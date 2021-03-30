package gmf_test

import (
	"log"
	"testing"

	"github.com/3d0c/gmf"
)

// @todo export rescaled frame as jpeg and compare dimension.
func TestScale(t *testing.T) {
	srcWidth, srcHeight := 640, 480
	dstWidth, dstHeight := 320, 200

	codec, err := gmf.FindEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	srcEncCtx := gmf.NewCodecCtx(codec)
	if srcEncCtx == nil {
		t.Fatal("Unable to allocate codec context")
	}
	srcEncCtx.SetWidth(640).SetHeight(480).SetPixFmt(gmf.AV_PIX_FMT_YUV420P)

	dstCodecCtx := gmf.NewCodecCtx(codec)
	if dstCodecCtx == nil {
		t.Fatal("Unable to allocate codec context")
	}
	defer dstCodecCtx.Free()

	dstCodecCtx.SetBitRate(400000).SetWidth(dstWidth).SetHeight(dstHeight).SetTimeBase(gmf.AVR{Num: 1, Den: 25}).SetGopSize(10).SetMaxBFrames(1).SetPixFmt(gmf.AV_PIX_FMT_YUV420P)

	dstCodecCtx.SetProfile(gmf.FF_PROFILE_MPEG4_SIMPLE)

	outputCtx := gmf.NewCtx()
	defer outputCtx.Free()

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		t.Fatalf("Unable to create stream for videoEnc [%s]\n", codec.LongName())
	}
	defer videoStream.Free()

	if err := dstCodecCtx.Open(nil); err != nil {
		t.Fatal(err)
	}

	videoStream.DumpContexCodec(dstCodecCtx)
	swsCtx, err := gmf.NewSwsCtx(srcEncCtx.Width(), srcEncCtx.Height(), srcEncCtx.PixFmt(), dstCodecCtx.Width(), dstCodecCtx.Height(), dstCodecCtx.PixFmt(), gmf.SWS_BICUBIC)
	if err != nil {
		t.Fatal(err)
	}
	defer swsCtx.Free()

	dstFrame := gmf.NewFrame().SetWidth(dstWidth).SetHeight(dstHeight).SetFormat(gmf.AV_PIX_FMT_YUV420P)
	defer dstFrame.Free()

	if err := dstFrame.ImgAlloc(); err != nil {
		t.Fatal(err)
	}

	var frame *gmf.Frame

	for frame = range gmf.GenSyntVideoNewFrame(srcWidth, srcHeight, srcEncCtx.PixFmt()) {
		frame.SetPts(0)

		swsCtx.Scale(frame, dstFrame)

		frame.Free()
		break
	}

	log.Println("Swscale is OK")
}
