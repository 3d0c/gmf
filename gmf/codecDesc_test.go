package gmf_test

import (
	"github.com/3d0c/gmf"
	"log"
	"testing"
)

var codecsMap = make(map[int]*gmf.CodecDescriptor)

func TestCodecs(t *testing.T) {
	gmf.InitDesc()
	nbCodecs := len(gmf.Codecs)

	if nbCodecs == 0 {
		t.Fatal("No codecs found. Expected any non zero value.")
	} else {
		log.Println(nbCodecs, "codec registered")
	}
}
