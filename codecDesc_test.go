package gmf

import (
	"log"
	"testing"
)

func TestCodecs(t *testing.T) {
	// Codecs glabal is initialized inside init() block in codec.go
	nbCodecs := len(Codecs)

	if nbCodecs == 0 {
		t.Fatal("No codecs found. Expected any non zero value.")
	} else {
		log.Println(nbCodecs, "codec registered")
	}
}
