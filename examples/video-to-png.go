package main

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"

	. "github.com/3d0c/gmf"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
}

func assert(i interface{}, err error) interface{} {
	if err != nil {
		fatal(err)
	}

	return i
}

var i, j int = 0, 0

func writeFile(b []byte) {
	name := fmt.Sprintf("./tmp/%d%d.png", j, i)

	fp, err := os.Create(name)
	if err != nil {
		fatal(err)
	}

	defer func() {
		if err := fp.Close(); err != nil {
			fatal(err)
		}
		i++
		if i == 9 {
			i = 0
			j++
		}
	}()

	if n, err := fp.Write(b); err != nil {
		fatal(err)
	} else {
		log.Println(n, "bytes written to", name)
	}
}

func main() {
	srcFileName := "tests-sample.mp4"

	os.Mkdir("./tmp", 0755)

	if len(os.Args) > 1 {
		srcFileName = os.Args[1]
	}

	inputCtx := assert(NewInputCtx(srcFileName)).(*FmtCtx)
	defer inputCtx.CloseInputAndRelease()

	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Println("No video stream found in", srcFileName)
	}

	codec, err := FindEncoder(AV_CODEC_ID_PNG)
	if err != nil {
		fatal(err)
	}

	cc := NewCodecCtx(codec)
	defer Release(cc)

	cc.SetPixFmt(AV_PIX_FMT_RGB24).SetWidth(srcVideoStream.CodecCtx().Width()).SetHeight(srcVideoStream.CodecCtx().Height()).SetTimeBase(AVR{1, 25})

	if codec.IsExperimental() {
		cc.SetStrictCompliance(FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		fatal(err)
	}

	swsCtx := NewSwsCtx(srcVideoStream.CodecCtx(), cc, SWS_BICUBIC)
	defer Release(swsCtx)

	dstFrame := NewFrame().
		SetWidth(srcVideoStream.CodecCtx().Width()).
		SetHeight(srcVideoStream.CodecCtx().Height()).
		SetFormat(AV_PIX_FMT_RGB24)
	defer Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}

	for packet := range inputCtx.GetNewPackets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			continue
		}

	decode:
		var errorCode int
		frame, errorCode := srcVideoStream.CodecCtx().Decode2(packet)
		if errorCode != 0 {
			// Retry if EAGAIN
			if errorCode == -35 {
				goto decode
			}

			log.Fatal(errorCode)
		}

		swsCtx.Scale(frame, dstFrame)

		if p, err := dstFrame.Encode(cc); p != nil {
			writeFile(p.Data())
			defer Release(p)
		} else if err != nil {
			fatal(err)
		}

		Release(packet)
	}

	Release(dstFrame)

}
