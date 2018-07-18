package main

import (
	"log"

	. "github.com/3d0c/gmf"
)

func fatal(err interface{}) {
	log.Fatal(err)
}

func main() {
	outputfilename := "sample-scale.mp4"
	srcWidth, srcHeight := 640, 480
	dstWidth, dstHeight := 320, 200

	// codec, err := FindEncoder(AV_CODEC_ID_MPEG1VIDEO)
	codec, err := FindEncoder("mpeg4")
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
	defer Release(dstCodecCtx)

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
	defer Release(videoStream)

	if err := dstCodecCtx.Open(nil); err != nil {
		fatal(err)
	}

	videoStream.SetCodecCtx(dstCodecCtx)

	// outputCtx.SetStartTime(0)

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	swsCtx := NewSwsCtx(srcEncCtx, dstCodecCtx, SWS_BICUBIC)
	defer Release(swsCtx)

	dstFrame := NewFrame().SetWidth(dstWidth).SetHeight(dstHeight).SetFormat(AV_PIX_FMT_YUV420P)

	if err := dstFrame.ImgAlloc(); err != nil {
		fatal(err)
	}

	var frame *Frame

	i := int64(0)
	for frame = range GenSyntVideoNewFrame(srcWidth, srcHeight, srcEncCtx.PixFmt()) {
		frame.SetPts(i)

		swsCtx.Scale(frame, dstFrame)

		if p, err := dstFrame.Encode(videoStream.CodecCtx()); err == nil {

			p.SetPts(RescaleQ(i, videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			p.SetDts(RescaleQ(i, videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))

			if err := outputCtx.WritePacket(p); err != nil {
				fatal(err)
			}

			Release(p)
		} else {
			fatal(err)
		}

		i++
		Release(frame)
	}

	Release(dstFrame)

	outputCtx.CloseOutputAndRelease()

	log.Println(i, "frames written to", outputfilename)
}
