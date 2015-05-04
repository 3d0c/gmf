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

	defer inputCtx.CloseInputAndRelease()

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

func TestGetNextFrame(t *testing.T) {
	inputCtx, err := NewInputCtx(inputSampleFilename)
	if err != nil {
		t.Fatal(err)
	}

	defer inputCtx.CloseInputAndRelease()

	for {
		packet := inputCtx.GetNextPacket()
		if packet == nil {
			break
		}
		if packet.Size() <= 0 {
			t.Fatal("Expected size > 0")
		}

		ist := assert(inputCtx.GetStream(0)).(*Stream)

		f := 0
		for {
			frame, err := packet.GetNextFrame(ist.CodecCtx())
			if frame == nil && err == nil {
				break
			}
			Release(frame)
			f++
		}

		if f >= 1 {
			log.Println(f, "frames decode.")
			Release(packet)
			break
		}

		Release(packet)
	}
}
