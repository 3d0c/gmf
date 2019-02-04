package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	. "github.com/3d0c/gmf"
)

var (
	section     *io.SectionReader
	srcfileName string
	seq         int
)

func customReader() ([]byte, int) {
	var file *os.File
	var err error

	if section == nil {
		file, err = os.Open(srcfileName)
		if err != nil {
			panic(err)
		}

		fi, err := file.Stat()
		if err != nil {
			panic(err)
		}

		section = io.NewSectionReader(file, 0, fi.Size())
	}

	b := make([]byte, IO_BUFFER_SIZE)

	n, err := section.Read(b)
	if err != nil && err == io.EOF {
		file.Close()
		return b, n
	}
	if err != nil {
		return nil, n
	}

	return b, n
}

func writeFile(b []byte) {
	name := "tmp-img/" + strconv.Itoa(seq) + ".jpg"

	fp, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := fp.Close(); err != nil {
			log.Fatal(err)
		}
		seq++
	}()

	if n, err := fp.Write(b); err != nil {
		log.Fatal(err)
	} else {
		log.Println(n, "bytes written to", name)
	}
}

func main() {
	if len(os.Args) > 1 {
		srcfileName = os.Args[1]
	} else {
		srcfileName = "tmp/big_buck_bunny.webm"
	}

	ictx := NewCtx()
	defer ictx.CloseInputAndRelease()

	if err := ictx.SetInputFormat("webm"); err != nil {
		log.Fatal(err)
	}

	avioCtx, err := NewAVIOContext(ictx, &AVIOHandlers{ReadPacket: customReader})
	defer Release(avioCtx)
	if err != nil {
		log.Fatal(err)
	}

	ictx.SetPb(avioCtx).OpenInput("")

	ist, err := ictx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Println("No video stream found in", srcfileName)
	}

	fmt.Println("ictx.Duration:", ictx.Duration())
	fmt.Printf("bitrate: %d/sec\n", ictx.BitRate())

	codec, err := FindEncoder(AV_CODEC_ID_JPEG2000)
	if err != nil {
		log.Fatal(err)
	}

	cc := NewCodecCtx(codec)
	defer Release(cc)

	cc.SetPixFmt(AV_PIX_FMT_RGB24).
		SetWidth(ist.CodecCtx().Width()).
		SetHeight(ist.CodecCtx().Height()).
		SetTimeBase(AVR{Num: 1, Den: 1})

	if codec.IsExperimental() {
		cc.SetStrictCompliance(FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		log.Fatal(err)
	}

	swsCtx := NewSwsCtx(ist.CodecCtx(), cc, SWS_BICUBIC)
	defer Release(swsCtx)

	dstFrame := NewFrame().
		SetWidth(ist.CodecCtx().Width()).
		SetHeight(ist.CodecCtx().Height()).
		SetFormat(AV_PIX_FMT_RGB24)
	defer Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil {
		log.Fatal(err)
	}

	i := 0
	for p := range ictx.GetNewPackets() {
		if p.StreamIndex() != ist.Index() {
			continue
		}

		fmt.Println(p)

	decode:
		frame, err := p.Frames(ist.CodecCtx())
		if err != nil {
			// Retry if EAGAIN
			if err.Error() == "Resource temporarily unavailable" {
				goto decode
			}
			log.Fatal(err)
		}

		swsCtx.Scale(frame, dstFrame)

		if p, err := dstFrame.Encode(cc); p != nil {
			writeFile(p.Data())
		} else if err != nil {
			log.Fatal(err)
		}

		i++

		Release(p)
	}

	fmt.Println(i, "packets")
}
