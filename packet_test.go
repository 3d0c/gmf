package gmf

import (
	"log"
	"testing"
)

func TestFramesIterator(t *testing.T) {
	inputCtx, err := NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer inputCtx.CloseInput()

	for packet := range inputCtx.GetNewPackets() {
		if packet.Size() <= 0 {
			t.Fatal("Expected size > 0")
		}

		ist := assert(inputCtx.GetStream(0)).(*Stream)

		f := 0
		for frame := range packet.Frames(ist.CodecCtx()) {
			log.Println(frame.Pts(), frame.PktPts(), frame.PktPos())
			f++
		}

		if f >= 1 {
			log.Println(f, "frames decode.")
			break
		}

		Release(packet)
	}

}
