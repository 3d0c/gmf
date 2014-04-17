package gmf

import (
	"log"
	"testing"
)

// @todo export rescaled frame as jpeg and compare dimension.
func TestScale(t *testing.T) {
	srcWidth, srcHeight := 640, 480
	dstWidth, dstHeight := 320, 200

	codec, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	srcEncCtx := NewCodecCtx(codec)
	if srcEncCtx == nil {
		t.Fatal("Unable to allocate codec context")
	}
	srcEncCtx.SetWidth(640).SetHeight(480).SetPixFmt(AV_PIX_FMT_YUV420P)

	dstCodecCtx := NewCodecCtx(codec)
	if dstCodecCtx == nil {
		t.Fatal("Unable to allocate codec context")
	}

	dstCodecCtx.SetBitRate(400000).SetWidth(dstWidth).SetHeight(dstHeight).SetTimeBase(AVR{1, 25}).SetGopSize(10).SetMaxBFrames(1).SetPixFmt(AV_PIX_FMT_YUV420P)

	dstCodecCtx.SetProfile(FF_PROFILE_MPEG4_SIMPLE)

	outputCtx := NewCtx()

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		t.Fatalf("Unable to create stream for videoEnc [%s]\n", codec.LongName())
	}

	if err := dstCodecCtx.Open(nil); err != nil {
		t.Fatal(err)
	}

	videoStream.SetCodecCtx(dstCodecCtx)

	swsCtx := NewSwsCtx(srcEncCtx, dstCodecCtx, SWS_BICUBIC)

	dstFrame := NewFrame().SetWidth(dstWidth).SetHeight(dstHeight).SetFormat(AV_PIX_FMT_YUV420P)

	if err := dstFrame.ImgAlloc(); err != nil {
		t.Fatal(err)
	}

	var frame *Frame

	i := 0
	for frame = range GenSyntVideo(srcWidth, srcHeight, srcEncCtx.PixFmt()) {
		frame.SetPts(i)

		swsCtx.Scale(frame, dstFrame)

		i++
		if i >= 1 {
			break
		}
	}

	outputCtx.Free()

	log.Println("Swscale is OK")
}
