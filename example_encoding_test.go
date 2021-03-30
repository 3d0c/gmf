package gmf_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/3d0c/gmf"
)

func Example() {
	outputfilename := "examples/sample-encoding1.mpg"
	dstWidth, dstHeight := 640, 480

	codec, err := gmf.FindEncoder(gmf.AV_CODEC_ID_MPEG1VIDEO)
	if err != nil {
		log.Fatal(err)
	}

	videoEncCtx := gmf.NewCodecCtx(codec)
	if videoEncCtx == nil {
		log.Fatal(errors.New("failed to create a new codec context"))
	}
	defer videoEncCtx.Free()

	outputCtx, err := gmf.NewOutputCtx(outputfilename)
	if err != nil {
		log.Fatal(errors.New("failed to create a new output context"))
	}
	defer outputCtx.Free()

	videoEncCtx.
		SetBitRate(400000).
		SetWidth(dstWidth).
		SetHeight(dstHeight).
		SetTimeBase(gmf.AVR{Num: 1, Den: 25}).
		SetPixFmt(gmf.AV_PIX_FMT_YUV420P).
		SetProfile(gmf.FF_PROFILE_MPEG4_SIMPLE).
		SetMbDecision(gmf.FF_MB_DECISION_RD)

	if outputCtx.IsGlobalHeader() {
		videoEncCtx.SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	}

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		log.Fatal(errors.New(fmt.Sprintf("Unable to create stream for videoEnc [%s]\n", codec.LongName())))
	}
	defer videoStream.Free()

	if err = videoEncCtx.Open(nil); err != nil {
		log.Fatal(err)
	}
	videoStream.DumpContexCodec(videoEncCtx)
	// videoStream.SetCodecCtx(videoEncCtx)

	outputCtx.SetStartTime(0)

	if err = outputCtx.WriteHeader(); err != nil {
		log.Fatal(err)
	}
	outputCtx.Dump()
	i := int64(0)
	n := 0

	for frame := range gmf.GenSyntVideoNewFrame(videoEncCtx.Width(), videoEncCtx.Height(), videoEncCtx.PixFmt()) {
		frame.SetPts(i)

		if p, err := frame.Encode(videoEncCtx); p != nil {
			if p.Pts() != gmf.AV_NOPTS_VALUE {
				p.SetPts(gmf.RescaleQ(p.Pts(), videoEncCtx.TimeBase(), videoStream.TimeBase()))
			}

			if p.Dts() != gmf.AV_NOPTS_VALUE {
				p.SetDts(gmf.RescaleQ(p.Dts(), videoEncCtx.TimeBase(), videoStream.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				log.Fatal(err)
			}

			n++

			log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", frame.Pts(), p.Size(), p.Pts(), p.Dts())

			p.Free()
		} else if err != nil {
			log.Fatal(err)
		}

		frame.Free()
		i++
	}
	fmt.Println("frames written to examples/sample-encoding1.mpg")
	// Output: frames written to examples/sample-encoding1.mpg
}
