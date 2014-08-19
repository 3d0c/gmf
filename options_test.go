package gmf

import (
	"log"
	"testing"
)

// @todo write good test
func TestOptionSet(t *testing.T) {
	codec, err := FindEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	cc := NewCodecCtx(codec, []*Option{{"refcounted_frames", 1}})
	if cc == nil {
		t.Fatal(err)
	}

	Release(cc)

	d := NewDict([]Pair{{"refcounted_frames", "1"}})
	cc2 := NewCodecCtx(codec, []*Option{{"dict", d}})
	if cc2 == nil {
		t.Fatal(err)
	}

	Release(cc2)
	log.Println("Options work")
}
