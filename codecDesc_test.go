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

func TestCodecsLongName(t *testing.T) {
	for _, codecDesc := range Codecs {
		codecsMap[codecDesc.Id()] = codecDesc
	}

	codec := codecsMap[28] // h264 codec
	expected := "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10"

	if actual := codec.LongName(); actual != expected {
		t.Fatalf("The long_name should '%s' not '%s'", expected, actual)
	}
}
