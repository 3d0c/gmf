package main

import (
	"flag"
	"fmt"
	. "github.com/3d0c/gmf"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
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

func encodeWorker(data chan *Frame, wg *sync.WaitGroup, srcCtx *CodecCtx) {
	defer wg.Done()
	log.Println("worker started")
	codec, err := FindEncoder(AV_CODEC_ID_JPEG2000)
	if err != nil {
		fatal(err)
	}

	cc := NewCodecCtx(codec)

	w, h := srcCtx.Width(), srcCtx.Height()

	cc.SetPixFmt(AV_PIX_FMT_RGB24).SetWidth(w).SetHeight(h)

	if codec.IsExperimental() {
		cc.SetStrictCompliance(-2)
	}

	if err := cc.Open(nil); err != nil {
		fatal(err)
	}

	swsCtx := NewSwsCtx(srcCtx, cc, SWS_BICUBIC)

	// convert to RGB, optionally resize could be here
	dstFrame := NewFrame().
		SetWidth(w).
		SetHeight(h).
		SetFormat(AV_PIX_FMT_RGB24)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}

	for {
		srcFrame, ok := <-data
		if !ok {
			break
		}

		swsCtx.Scale(srcFrame, dstFrame)

		if p, ready, _ := dstFrame.Encode(cc); ready {
			writeFile(p.Data())
		}
	}

	dstFrame.Free()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	os.Mkdir("./tmp", 0755)

	wnum := flag.Int("wnum", 10, "number of workers")
	srcFileName := flag.String("input", "tests-sample.mp4", "input file")

	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: %s [OPTIONS]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	inputCtx := assert(NewInputCtx(*srcFileName)).(*FmtCtx)
	defer inputCtx.CloseInput()

	srcVideoStream, err := inputCtx.GetBestStream(AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Println("No video stream found in", srcFileName)
	}

	wg := new(sync.WaitGroup)

	dataChan := make(chan *Frame)

	for i := 0; i < *wnum; i++ {
		wg.Add(1)
		go encodeWorker(dataChan, wg, srcVideoStream.CodecCtx())
	}

	for packet := range inputCtx.Packets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			continue
		}

		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)

		for frame := range packet.Frames(ist.CodecCtx()) {
			dataChan <- frame
		}
	}

	close(dataChan)

	wg.Wait()
}
