package main

import (
	. "github.com/3d0c/gmf"
	"log"
	"os"
	"runtime/debug"
	"strconv"
)

func fatal(err error) {
	debug.PrintStack()
	log.Fatal(err)
	os.Exit(0)
}

func assert(i interface{}, err error) interface{} {
	if err != nil {
		fatal(err)
	}

	return i
}

var i int = 0

func writeFile(b []byte) {
	name := "./tmp/" + strconv.Itoa(i) + ".jpg"

	fp, err := os.Create(name)
	if err != nil {
		fatal(err)
	}

	defer func() {
		if err := fp.Close(); err != nil {
			fatal(err)
		}
		i++
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

	codec, err := FindEncoder("mjpeg")
	if err != nil {
		fatal(err)
	}

	cc := NewCodecCtx(codec)
	defer Release(cc)

	cc.SetPixFmt(AV_PIX_FMT_YUVJ420P)
	cc.SetWidth(srcVideoStream.CodecCtx().Width())
	cc.SetHeight(srcVideoStream.CodecCtx().Height())
	cc.SetTimeBase(srcVideoStream.CodecCtx().TimeBase().AVR())

	if codec.IsExperimental() {
		cc.SetStrictCompliance(-2)
	}

	if err := cc.Open(nil); err != nil {
		fatal(err)
	}

	for packet := range inputCtx.GetNewPackets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			continue
		}
		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)

		for frame := range packet.Frames(ist.CodecCtx()) {
			if p, ready, _ := frame.EncodeNewPacket(cc); ready {
				writeFile(p.Data())
				defer Release(p)
			}
		}
		Release(packet)
	}
}
