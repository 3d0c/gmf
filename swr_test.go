package gmf

import (
	"log"
	"testing"
)

func TestSwrInit(t *testing.T) {
	options := []*Option{
		{"in_channel_count", 2},
		{"in_sample_rate", 44100},
		{"in_sample_fmt", AV_SAMPLE_FMT_S16},
		{"out_channel_count", 2},
		{"out_sample_rate", 44100},
		{"out_sample_fmt", AV_SAMPLE_FMT_S16},
	}

	swrCtx := NewSwrCtx(options, nil)
	if swrCtx == nil {
		t.Fatal("unable to create Swr Context")
	} else {
		Release(swrCtx)
	}

	log.Println("Swr context is createad")
}
