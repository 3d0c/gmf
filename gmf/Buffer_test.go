package gmf


import (
	"testing"
	"log"
)

func TestBufferFormat(t *testing.T) {
	buffer := Buffer{}
	buffer.Format = VideoFormat{}
	switch v := buffer.Format.(type) {
	case VideoFormat:
		log.Printf("VideoFormat")
		f := buffer.Format.(VideoFormat)
		f.Width = 10
		//bla.Width=10
	case AudioFormat:
		log.Printf("AudioFormat")
	default:
		log.Printf("unknown Format")
	}

}
