package gmf_test

import (
	"github.com/3d0c/gmf"
	"testing"
)

func TestFramesIterator(t *testing.T) {
	inputCtx, err := gmf.NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer inputCtx.Free()

	cnt := 0
	ist := assert(inputCtx.GetStream(0)).(*gmf.Stream)
	t.Log("NbFrames", ist.NbFrames())
	ctx := ist.CodecCtx()
	for frame := range gmf.GenSyntVideoNewFrame(ctx.Width(), ctx.Height(), ctx.PixFmt()) {
		cnt++
		frame.Free()
	}

	if cnt != 25 {
		t.Fatalf("Expected %d frames, obtained %d\n", 25, cnt)
	}
}
