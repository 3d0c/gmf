package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"log"
	"os"
	"time"

	"github.com/3d0c/gmf"
)

var fileCount int = 0

func main() {
	var srcFileName string

	flag.StringVar(&srcFileName, "src", "tests-sample.mp4", "source video")
	flag.Parse()

	os.MkdirAll("./tmp", 0755)

	inputCtx, err := gmf.NewInputCtx(srcFileName)
	if err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer inputCtx.CloseInputAndRelease()

	srcVideoStream, err := inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Println("No video stream found in", srcFileName)
	}

	codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_RAWVIDEO)
	if err != nil {
		log.Fatal(err)
	}

	cc := gmf.NewCodecCtx(codec)
	defer gmf.Release(cc)

	cc.SetTimeBase(gmf.AVR{Num: 1, Den: 1})

	cc.SetPixFmt(gmf.AV_PIX_FMT_RGBA).SetWidth(srcVideoStream.CodecCtx().Width()).SetHeight(srcVideoStream.CodecCtx().Height())
	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		log.Fatal(err)
	}

	ist, err := inputCtx.GetStream(srcVideoStream.Index())
	if err != nil {
		log.Fatalf("Error getting stream - %s\n", err)
	}
	defer gmf.Release(ist)

	codecCtx := ist.CodecCtx()
	defer gmf.Release(codecCtx)

	// convert source pix_fmt into AV_PIX_FMT_RGBA
	// which is set up by codec context above
	ist.SwsCtx = gmf.NewSwsCtx(srcVideoStream.CodecCtx(), cc, gmf.SWS_BICUBIC)
	// gmf.Stream.Rescaler is a function pointer
	// default handler is "DefaultRescaler"
	ist.Rescaler = gmf.DefaultRescaler

	start := time.Now()

	var (
		pkt        *gmf.Packet
		frames     []*gmf.Frame
		drain      int = -1
		frameCount int = 0
	)

	for {
		if drain >= 0 {
			break
		}

		pkt, err = inputCtx.GetNextPacket()
		if err != nil && err != io.EOF {
			if pkt != nil {
				pkt.Free()
			}
			log.Printf("error getting next packet - %s", err)
			break
		} else if err != nil && pkt == nil {
			drain = 0
		}

		if pkt != nil && pkt.StreamIndex() != srcVideoStream.Index() {
			continue
		}

		frames, err = ist.CodecCtx().Decode(pkt)
		if err != nil {
			log.Printf("Fatal error during decoding - %s\n", err)
			break
		}

		// Decode() method doesn't treat EAGAIN and EOF as errors
		// it returns empty frames slice instead. Countinue until
		// input EOF or frames received.
		if len(frames) == 0 && drain < 0 {
			continue
		}

		frames = ist.Rescaler(ist, frames)

		encode(cc, frames, drain)

		for i, _ := range frames {
			frames[i].Free()
			frameCount++
		}

		if pkt != nil {
			pkt.Free()
			pkt = nil
		}
	}

	since := time.Since(start)
	log.Printf("Finished in %v, avg %.2f fps", since, float64(frameCount)/since.Seconds())
}

func encode(cc *gmf.CodecCtx, frames []*gmf.Frame, drain int) {
	packets, err := cc.Encode(frames, drain)
	if err != nil {
		log.Fatalf("Error encoding - %s\n", err)
	}
	if len(packets) == 0 {
		return
	}

	for _, p := range packets {
		width, height := cc.Width(), cc.Height()

		img := new(image.RGBA)
		img.Pix = p.Data()
		img.Stride = 4 * width
		img.Rect = image.Rect(0, 0, width, height)

		writeFile(img)

		p.Free()
	}

	return
}

func writeFile(b image.Image) {
	name := fmt.Sprintf("tmp/%d.jpg", fileCount)
	fp, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error opening file '%s' - %s\n", name, err)
	}
	defer fp.Close()

	fileCount++

	log.Printf("Saving file %s\n", name)

	if err = jpeg.Encode(fp, b, &jpeg.Options{Quality: 80}); err != nil {
		log.Fatal(err)
	}
}
