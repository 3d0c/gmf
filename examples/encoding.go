package main

import (
	"errors"
	"fmt"
	. "github.com/3d0c/gmf"
	"log"
	"os"
)

func fatal(err error) {
	log.Fatal(err)
	os.Exit(0)
}

func main() {
	outputfilename := "sample-encoding.mpg"
	codec, err := NewEncoder(AV_CODEC_ID_MPEG1VIDEO)
	// codec, err := NewEncoder("mpeg4")
	if err != nil {
		fatal(err)
	}

	videoEncCtx := NewCodecCtx(codec)
	if videoEncCtx == nil {
		fatal(err)
	}

	outputCtx, err := NewOutputCtx(outputfilename)
	if err != nil {
		fatal(err)
	}

	videoEncCtx.
		SetBitRate(400000).
		SetWidth(dstWidth).
		SetHeight(dstHeight).
		SetTimeBase(AVR{1, 25}).
		SetPixFmt(AV_PIX_FMT_YUV420P).
		SetGopSize(10).
		SetMaxBFrames(1).
		SetProfile(FF_PROFILE_MPEG4_SIMPLE)

	// videoEncCtx.SetProfile(C.FF_PROFILE_MPEG4_SIMPLE)

	if outputCtx.IsGlobalHeader() {
		videoEncCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
		log.Println("AVFMT_GLOBALHEADER flag is set.")
	}

	videoStream := outputCtx.NewStream(codec, nil)
	if videoStream == nil {
		fatal(errors.New(fmt.Sprintf("Unable to create stream for videoEnc [%s]\n", codec.LongName())))
	}

	if err := videoEncCtx.Open(nil); err != nil {
		fatal(err)
	}

	if err := videoStream.SetCodecCtx(videoEncCtx); err != nil {
		fatal(err)
	}

	outputCtx.SetStartTime(0)

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	var frame *Frame
	i := 0

	for frame = range GenSyntVideo(videoEncCtx.Width(), videoEncCtx.Height(), videoEncCtx.PixFmt()) {
		frame.SetPts(i)

		if p, ready, err := frame.Encode(videoStream.GetCodecCtx()); ready {
			log.Println("pkt[orig].Pts/Dts/Size:", p.Pts(), p.Dts(), p.Size())

			if p.Pts() != AV_NOPTS_VALUE {
				p.SetPts(RescaleQ(p.Pts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if p.Dts() != AV_NOPTS_VALUE {
				p.SetDts(RescaleQ(p.Dts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				fatal(err)
			}

			log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())

		} else if err != nil {
			fatal(err)
		}

		i++
	}

	frame.SetPts(i)

	if p, ready, _ := frame.Encode(videoStream.GetCodecCtx()); ready {
		p.SetPts(RescaleQ(p.Pts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))

		p.SetDts(RescaleQ(p.Dts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))

		if err := outputCtx.WritePacket(p); err != nil {
			fatal(err)
		}
		log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())
	}

	outputCtx.CloseOutput()

	log.Println(i, "frames written to", outputfilename)
}
