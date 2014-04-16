package gmf

import (
	"log"
	"testing"
)

func TestStream(t *testing.T) {
	ctx := NewCtx()

	codec, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	cc := NewCodecCtx(codec)
	if cc == nil {
		t.Fatal("Unable to allocate codec context")
	}

	if ctx.NewStream(codec) == nil {
		t.Fatal("Unable to create new stream")
	}

	if assert(ctx.GetStream(0)).(*Stream).IsAudio() {
		t.Fatal("Expected type is video, audio got")
	}

	if !assert(ctx.GetStream(0)).(*Stream).IsVideo() {
		t.Fatal("Expected type id video")
	}

	ctx.Free()

	log.Println("Stream is OK")
}
