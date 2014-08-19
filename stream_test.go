package gmf

import (
	"log"
	"testing"
)

func TestStream(t *testing.T) {
	ctx := NewCtx()

	vc, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	cc := NewCodecCtx(vc)
	if cc == nil {
		t.Fatal("Unable to allocate codec context")
	}

	if ctx.NewStream(vc) == nil {
		t.Fatal("Unable to create new stream")
	}

	td := CodecCtxTestData

	cc.SetWidth(td.width).SetHeight(td.height).SetTimeBase(td.timebase).SetPixFmt(td.pixfmt).SetBitRate(td.bitrate)

	if err := cc.Open(nil); err != nil {
		t.Fatal(err)
	}

	st := assert(ctx.GetStream(0)).(*Stream)

	st.SetCodecCtx(cc)

	if st.CodecCtx().Height() != td.height || st.CodecCtx().Width() != td.width {
		t.Fatalf("Expected dimension = %dx%d, %dx%d got\n", td.width, td.height, st.CodecCtx().Width(), st.CodecCtx().Height())
	}

	ctx.Free()

	log.Println("Stream is OK")
}

func TestStreamInputCtx(t *testing.T) {
	inputCtx, err := NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	ist := assert(inputCtx.GetStream(0)).(*Stream)

	if ist.CodecCtx().Width() != inputSampleWidth || ist.CodecCtx().Height() != inputSampleHeight {
		t.Fatalf("Expected dimension = %dx%d, %dx%d got\n", inputSampleWidth, inputSampleHeight, ist.CodecCtx().Width(), ist.CodecCtx().Height())
	}

	log.Printf("Input stream is OK, cnt: %d, %dx%d\n", inputCtx.StreamsCnt(), ist.CodecCtx().Width(), ist.CodecCtx().Height())

	inputCtx.CloseInputAndRelease()
}
