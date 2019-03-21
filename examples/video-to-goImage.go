package main

/* Valgrind report summary

==13531== LEAK SUMMARY:
==13531==    definitely lost: 0 bytes in 0 blocks
==13531==    indirectly lost: 0 bytes in 0 blocks
==13531==      possibly lost: 1,440 bytes in 5 blocks
==13531==    still reachable: 0 bytes in 0 blocks
==13531==         suppressed: 0 bytes in 0 blocks

*/

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
	var (
		srcFileName string
		swsctx      *gmf.SwsCtx
	)

	flag.StringVar(&srcFileName, "src", "tests-sample.mp4", "source video")
	flag.Parse()

	os.MkdirAll("./tmp", 0755)

	inputCtx, err := gmf.NewInputCtx(srcFileName)
	if err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer inputCtx.Free()

	srcVideoStream, err := inputCtx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Printf("No video stream found in '%s'\n", srcFileName)
		return
	}

	codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_RAWVIDEO)
	if err != nil {
		log.Fatalf("%s\n", err)
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
	defer cc.Free()

	ist, err := inputCtx.GetStream(srcVideoStream.Index())
	if err != nil {
		log.Fatalf("Error getting stream - %s\n", err)
	}
	defer ist.Free()

	// convert source pix_fmt into AV_PIX_FMT_RGBA
	// which is set up by codec context above
	icc := srcVideoStream.CodecCtx()
	if swsctx, err = gmf.NewSwsCtx(icc.Width(), icc.Height(), icc.PixFmt(), cc.Width(), cc.Height(), cc.PixFmt(), gmf.SWS_BICUBIC); err != nil {
		panic(err)
	}
	defer swsctx.Free()

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

		if frames, err = gmf.DefaultRescaler(swsctx, frames); err != nil {
			panic(err)
		}

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

	for i := 0; i < inputCtx.StreamsCnt(); i++ {
		st, _ := inputCtx.GetStream(i)
		st.CodecCtx().Free()
		st.Free()
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
