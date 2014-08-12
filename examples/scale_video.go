package main

import (
	. "github.com/3d0c/gmf"
	"log"
	"os"
)

func fatal(err interface{}) {
	log.Fatal(err)
	os.Exit(0)
}

func main() {
	outputfilename := "sample-scale.mp4"
	srcWidth, srcHeight := 640, 480
	dstWidth, dstHeight := 320, 200

	// codec, err := NewEncoder(AV_CODEC_ID_MPEG1VIDEO)
	codec, err := NewEncoder("mpeg4")
	if err != nil {
		fatal(err)
	}

	srcEncCtx := NewCodecCtx(codec)
	if srcEncCtx == nil {
		fatal("Unable to allocate codec context")
	}
	srcEncCtx.SetWidth(640).SetHeight(480).SetPixFmt(AV_PIX_FMT_YUV420P)

	dstCodecCtx := NewCodecCtx(codec)
	if dstCodecCtx == nil {
		fatal("Unable to allocate codec context")
	}

	dstCodecCtx.
		SetBitRate(400000).
		SetWidth(dstWidth).
		SetHeight(dstHeight).
		SetTimeBase(AVR{1, 25}).
		SetPixFmt(AV_PIX_FMT_YUV420P).
		SetProfile(FF_PROFILE_MPEG4_SIMPLE)

	outputCtx, err := NewOutputCtx(outputfilename)
	if err != nil {
		fatal(err)
	}

	if outputCtx.IsGlobalHeader() {
		dstCodecCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	}

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		fatal("Unable to create stream for videoEnc " + codec.LongName())
	}

	if err := dstCodecCtx.Open(nil); err != nil {
		fatal(err)
	}

	videoStream.SetCodecCtx(dstCodecCtx)

	// outputCtx.SetStartTime(0)

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	swsCtx := NewSwsCtx(srcEncCtx, dstCodecCtx, SWS_BICUBIC)

	dstFrame := NewFrame().SetWidth(dstWidth).SetHeight(dstHeight).SetFormat(AV_PIX_FMT_YUV420P)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}

	var frame *Frame

	i := 0
	for frame = range GenSyntVideo(srcWidth, srcHeight, srcEncCtx.PixFmt()) {
		frame.SetPts(i)

		swsCtx.Scale(frame, dstFrame)

		if p, ready, err := dstFrame.Encode(videoStream.CodecCtx()); ready {
			if p.Pts() != AV_NOPTS_VALUE {
				p.SetPts(RescaleQ(p.Pts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if p.Dts() != AV_NOPTS_VALUE {
				p.SetDts(RescaleQ(p.Dts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				fatal(err)
			}

			p.Free()

			log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())

		} else if err != nil {
			fatal(err)
		}

		i++
	}

	frame.SetPts(i)

	if p, ready, _ := frame.Encode(videoStream.CodecCtx()); ready {
		p.SetPts(RescaleQ(p.Pts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))

		p.SetDts(RescaleQ(p.Dts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))

		if err := outputCtx.WritePacket(p); err != nil {
			fatal(err)
		}
		p.Free()

		log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())
	}

	dstFrame.Free()
	frame.Free()

	outputCtx.CloseOutput()

	log.Println(i, "frames written to", outputfilename)
}
