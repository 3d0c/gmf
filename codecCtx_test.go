package gmf

import (
	"log"
	"testing"
)

var CodecCtxTestData = struct {
	width    int
	height   int
	timebase AVR
	pixfmt   int32
	bitrate  int
}{
	100,
	200,
	AVR{1, 25},
	AV_PIX_FMT_YUV420P,
	400000,
}

func TestCodecCtx(t *testing.T) {
	td := CodecCtxTestData

	codec, err := FindEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	cc := NewCodecCtx(codec)
	if cc == nil {
		t.Fatal("Unable to allocate codec context")
	}

	cc.SetWidth(td.width).SetHeight(td.height).SetTimeBase(td.timebase).SetPixFmt(td.pixfmt).SetBitRate(td.bitrate)

	if cc.Width() != td.width {
		t.Fatalf("Expected width = %v, %v got.\n", td.width, cc.Width())
	}

	if cc.Height() != td.height {
		t.Fatalf("Expected height = %v, %v got.\n", td.height, cc.Height())
	}

	if cc.TimeBase().AVR().Num != td.timebase.Num || cc.TimeBase().AVR().Den != td.timebase.Den {
		t.Fatal("Expected AVR = %v, %v got", cc.TimeBase().AVR())
	}

	if cc.PixFmt() != td.pixfmt {
		t.Fatalf("Expected pixfmt = %v, %v got.\n", td.pixfmt, cc.PixFmt())
	}

	if err := cc.Open(nil); err != nil {
		t.Fatal(err)
	}

	log.Println("CodecCtx is successfully created and opened.")

	Release(cc)
}
