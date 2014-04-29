package gmf

import (
	"log"
	"testing"
)

// @todo write good test
func TestOptionSet(t *testing.T) {
	codec, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	cc := NewCodecCtx(codec, []*Option{{"refcounted_frames", 1}})
	if cc == nil {
		t.Fatal(err)
	}

	cc.Free()

	d := NewDict([]Pair{{"refcounted_frames", "1"}})
	cc2 := NewCodecCtx(codec, []*Option{{"dict", d}})
	if cc2 == nil {
		t.Fatal(err)
	}

	cc2.Free()

	log.Println("Options work")
}
