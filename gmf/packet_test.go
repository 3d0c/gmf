package gmf_test

import (
	"testing"

	"github.com/3d0c/gmf"
)

func TestFramesIterator(t *testing.T) {
	inputCtx, err := gmf.NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}
	defer inputCtx.Free()

	cnt := 0
	ist := assert(inputCtx.GetStream(0)).(*gmf.Stream)
	par := ist.CodecPar()
	for frame := range gmf.GenSyntVideoNewFrame(par.Width(), par.Height(), par.Format()) {
		cnt++
		frame.Free()
	}

	if cnt != 25 {
		t.Fatalf("Expected %d frames, obtained %d\n", 25, cnt)
	}
}

func ExamplePacket_SetData() {
	p := gmf.NewPacket()
	defer p.Free()
	p.SetData([]byte{0x00, 0x00, 0x00, 0x01, 0x67})
	defer p.FreeData()
}

func ExamplePacket_SetFlags() {
	p := gmf.NewPacket()
	p.SetFlags(p.Flags() | gmf.AV_PKT_FLAG_KEY)
	defer p.Free()
}
