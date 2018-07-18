package main

import (
	"image"
	"image/jpeg"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/3d0c/gmf"
)

func assert(i interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(err)
	}

	return i
}

var i int = 0

func writeFile(b image.Image) {
	name := "./tmp/" + strconv.Itoa(i) + ".jpg"

	fp, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}

	err = jpeg.Encode(fp, b, &jpeg.Options{Quality: 100})
	if err != nil {
		log.Fatal(err)
	}

	if err := fp.Close(); err != nil {
		log.Fatal(err)
	}
	i++
}

func main() {
	srcFileName := "tests-sample.mp4"

	os.MkdirAll("./tmp", 0755)

	if len(os.Args) > 1 {
		srcFileName = os.Args[1]
	}

	inputCtx := assert(gmf.NewInputCtx(srcFileName)).(*gmf.FmtCtx)
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

	// This should really be AV_PIX_FMT_RGB32, but then red and blue channels are switched
	cc.SetPixFmt(gmf.AV_PIX_FMT_BGR32).SetWidth(srcVideoStream.CodecCtx().Width()).SetHeight(srcVideoStream.CodecCtx().Height())
	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if err := cc.Open(nil); err != nil {
		log.Fatal(err)
	}

	swsCtx := gmf.NewSwsCtx(srcVideoStream.CodecCtx(), cc, gmf.SWS_BICUBIC)
	defer gmf.Release(swsCtx)

	dstFrame := gmf.NewFrame().
		SetWidth(srcVideoStream.CodecCtx().Width()).
		SetHeight(srcVideoStream.CodecCtx().Height()).
		SetFormat(gmf.AV_PIX_FMT_BGR32) // see above

	if err := dstFrame.ImgAlloc(); err != nil {
		log.Fatal(err)
	}

	ist := assert(inputCtx.GetStream(srcVideoStream.Index())).(*gmf.Stream)
	defer gmf.Release(ist)

	codecCtx := ist.CodecCtx()
	defer gmf.Release(codecCtx)

	start := time.Now()

	for packet := range inputCtx.GetNewPackets() {
		if packet.StreamIndex() != srcVideoStream.Index() {
			// skip non video streams
			continue
		}

		frame, err := packet.Frames(codecCtx)
		if err != nil {
			log.Fatal(err)
		}
		swsCtx.Scale(frame, dstFrame)

		p, err := dstFrame.Encode(cc)

		if err != nil {
			log.Fatal(err)
		}

		width, height := frame.Width(), frame.Height()
		img := new(image.RGBA)
		img.Pix = p.Data()
		img.Stride = 4 * width // 4 bytes per pixel (RGBA), width times per row
		img.Rect = image.Rect(0, 0, width, height)

		writeFile(img)

		i++
		gmf.Release(p)

		gmf.Release(frame)

		gmf.Release(packet)
	}

	since := time.Since(start)
	log.Printf("Average %.2f fps", float64(i)/since.Seconds())
	log.Println("Total time ", since)

	gmf.Release(dstFrame)
}
