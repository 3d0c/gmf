package gmf

import (
	"log"
	"testing"
)

var codecsMap = make(map[int]*CodecDescriptor)

func TestCodecs(t *testing.T) {
	// Codecs glabal is initialized inside init() block in codec.go
	nbCodecs := len(Codecs)

	if nbCodecs == 0 {
		t.Fatal("No codecs found. Expected any non zero value.")
	} else {
		log.Println(nbCodecs, "codec registered")
	}
}

// func TestListCodecs(t *testing.T) {
// 	for _, codecDesc := range Codecs {
// 		log.Printf("%s, %s\n", codecDesc.Name(), codecDesc.LongName())
// 	}
// }
