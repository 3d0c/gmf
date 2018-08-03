package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"sync/atomic"

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

var i int32 = 0

func writeFile(b []byte) {
	name := "./tmp/" + strconv.Itoa(int(atomic.AddInt32(&i, 1))) + ".jpg"

	fp, err := os.Create(name)
	if err != nil {
		fatal(err)
	}

	defer func() {
		if err := fp.Close(); err != nil {
			fatal(err)
		}
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
	defer Release(cc)

	w, h := srcCtx.Width(), srcCtx.Height()

	cc.SetPixFmt(AV_PIX_FMT_RGB24).SetWidth(w).SetHeight(h)
	cc.SetTimeBase(AVR{1, 25})
	
	if codec.IsExperimental() {
		cc.SetStrictCompliance(FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		fatal(err)
	}

	swsCtx := NewSwsCtx(srcCtx, cc, SWS_BICUBIC)
	defer Release(swsCtx)

	// convert to RGB, optionally resize could be here
	dstFrame := NewFrame().
		SetWidth(w).
		SetHeight(h).
		SetFormat(AV_PIX_FMT_RGB24)
	defer Release(dstFrame)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}

	for {
		srcFrame, ok := <-data
		if !ok {
			break
		}
		//		log.Printf("srcFrome = %p",srcFrame)
		swsCtx.Scale(srcFrame, dstFrame)
		p, err := dstFrame.Encode(cc)
		if err == nil {
			writeFile(p.Data())
		} else {
			Release(srcFrame)
			fatal(err)
		}
		Release(srcFrame)
	}

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
	defer inputCtx.CloseInputAndRelease()

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

	for packet := range inputCtx.GetNewPackets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			continue
		}

		ist := assert(inputCtx.GetStream(packet.StreamIndex())).(*Stream)

		frame, err := packet.Frames(ist.CodecCtx())
		if err != nil {
			log.Fatal(err)
		}

		dataChan <- frame.CloneNewFrame()

		Release(packet)
	}

	close(dataChan)

	wg.Wait()
}
