package gmf

import (
	"log"
	"testing"
)

func TestCodecCtx(t *testing.T) {
	testdata := struct {
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

	codec, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	cc := NewCodecCtx(codec)
	if cc == nil {
		t.Fatal("Unable to allocate codec context")
	}

	cc.SetWidth(testdata.width).SetHeight(testdata.height).SetTimeBase(testdata.timebase).SetPixFmt(testdata.pixfmt).SetBitRate(testdata.bitrate)

	if cc.Width() != testdata.width {
		t.Fatalf("Expected width = %v, %v got.\n", testdata.width, cc.Width())
	}

	if cc.Height() != testdata.height {
		t.Fatalf("Expected height = %v, %v got.\n", testdata.height, cc.Height())
	}

	if cc.TimeBase().AVR().Num != testdata.timebase.Num || cc.TimeBase().AVR().Den != testdata.timebase.Den {
		t.Fatal("Expected AVR = %v, %v got", cc.TimeBase().AVR())
	}

	if cc.PixFmt() != testdata.pixfmt {
		t.Fatalf("Expected pixfmt = %v, %v got.\n", testdata.pixfmt, cc.PixFmt())
	}

	if err := cc.Open(nil); err != nil {
		t.Fatal(err)
	}

	log.Println("CodecCtx is successfully created and opened.")

	cc.Release()
}
