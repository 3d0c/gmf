package gmf_test

import (
	"github.com/3d0c/gmf"
	"log"
	"testing"
)

func TestSwrInit(t *testing.T) {
	options := []*gmf.Option{
		{"in_channel_count", 2},
		{"in_sample_rate", 44100},
		{"in_sample_fmt", gmf.AV_SAMPLE_FMT_S16},
		{"out_channel_count", 2},
		{"out_sample_rate", 44100},
		{"out_sample_fmt", gmf.AV_SAMPLE_FMT_S16},
	}

	swrCtx, err := gmf.NewSwrCtx(options, 2, gmf.AV_SAMPLE_FMT_S16)
	if err != nil {
		t.Fatal("unable to create Swr Context")
	} else {
		swrCtx.Free()
	}

	log.Println("Swr context is createad")
}
