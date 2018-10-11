package gmf

import (
	"log"
	"testing"
)

var codecsMap = make(map[int]*CodecDescriptor)

func TestCodecs(t *testing.T) {
	InitDesc()
	nbCodecs := len(Codecs)

	if nbCodecs == 0 {
		t.Fatal("No codecs found. Expected any non zero value.")
	} else {
		log.Println(nbCodecs, "codec registered")
	}
}
