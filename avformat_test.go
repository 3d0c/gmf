package gmf

import (
	"log"
	"testing"
)

var (
	testVideoFile   = "tmp/video.mpg"
	testVideoOutput = "tmp/out.mpg"
)

func init() {
	// log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func TestFmtCtx(t *testing.T) {
	ctx := NewCtx()

	if ctx.avCtx == nil {
		t.Fatal("AVContenxt is not initialized")
	}
}

func TestStreamFail(t *testing.T) {
	ctx := NewCtx()

	if stream, err := ctx.GetStream(1); err != nil {
		log.Printf("Expected error got: %v", err)
	} else {
		t.Errorf("Expected error, stream with id %d got", stream.Id())
	}

	ctx.Free()
}

func TestOutputCtx(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	formatName := "mpeg"

	ofmt := NewOutputFmt(formatName, "", "")
	if ofmt == nil {
		t.Fatal("OutputFormat is not initialized.")
	}

	if ofmt.Name() != formatName {
		t.Fatalf("Expected name '%s', '%s' got\n", formatName, ofmt.Name())
	}

	if err := ctx.SetOformat(ofmt); err != nil {
		t.Fatalf("Unable to set 'oformat' to 'context'. Error:", err)
	}
}

func TestOutputCtx_shouldfail(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	if err := ctx.SetOformat(NewOutputFmt("-invalid-", "", "")); err == nil {
		t.Fatalf("Error is expected.")
	}
}

func TestNewStream(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	if err := ctx.SetOformat(NewOutputFmt("mpeg", "", "")); err != nil {
		t.Fatalf("Unable to set 'oformat' to 'context'. Error:", err)
	}

	for i := 0; i < 2; i++ {
		stream := ctx.NewStream(NewEncoder("mpeg4"))
		if stream == nil {
			t.Fatalf("Unable to create new stream")
		}

		if stream.Index() != i {
			t.Fatalf("Expected stream index = %d, %d got\n", i, stream.Index())
		}
	}

	if ctx.StreamsCnt() != 2 {
		t.Fatalf("Expected streams count = 2, %d got\n", ctx.StreamsCnt())
	}
}

func TestOpenInput(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	err := ctx.OpenInput(testVideoFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("File '%s' opened.\n", testVideoFile)
	}
}

func TestPacketsIterator(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	err := ctx.OpenInput(testVideoFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("File %s contains %d streams.\n", testVideoFile, ctx.StreamsCnt())
	}

	i := 0

	for packet := range ctx.Packets() {
		if packet.Size() == 0 {
			t.Fatalf("Expected any non zero size.")
		} else {
			// log.Println("Size =", packet.Size(), "Stream index:", packet.StreamIndex())
		}

		stream, err := ctx.GetStream(packet.StreamIndex())
		if err != nil {
			t.Fatal("Unexpected error:", err)
		}

		if stream.CodecCtx().Id() == 0 {
			t.Fatal("Expected any non zero id for codecCtx")
		} else {
			// log.Println("codecCtx Id:", stream.CodecCtx().Id())
		}

		_, _, err = packet.Decode(stream.CodecCtx())
		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		} else {
			// log.Println("Frame data:", frame.Format(), frame.Width(), frame.Height())
		}

		i++
	}

	if i == 0 {
		t.Error("Expected any non zero value")
	} else {
		log.Println(i, "frames iterated.")
	}
}

func TestEncode(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	outCtx := NewCtx()
	defer outCtx.Free()

	if err := ctx.OpenInput(testVideoFile); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("Input file %s opened. %d streams found.\n", testVideoFile, ctx.StreamsCnt())
	}

	i := 0

	audioEnc, err := NewEncoder("mp2")
	if err != nil {
		t.Fatal(err)
	}

	videoEnc, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	if outCtx.NewStream(audioEnc, nil) == nil {
		t.Fatalf("Unable to create stream for audioEnc [%s]\n", audioEnc.LongName())
	}

	if outCtx.NewStream(videoEnc, nil) == nil {
		t.Fatalf("Unable to create stream for videoEnc [%s]\n", videoEnc.LongName())
	}

	if cnt := outCtx.StreamsCnt(); cnt != 2 {
		t.Fatalf("Expected stream count in output context = 2, %d got", cnt)
	} else {
		log.Println("2 streams created in output context")
	}

	// STOP HERE
	// check for avio_open
	// check for AVFMT_NOFILE
	// see muxing.c:506

	if err := outCtx.OpenOutput(NewOutputFmt("mpeg", testVideoOutput, "")); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("Output file %s opened", testVideoOutput)
	}

	audioEncCtx := NewCodecCtx(audioEnc)
	if audioEncCtx == nil {
		t.Fatal("Unable to create audio encoder context.")
	}

	if err := audioEncCtx.Open(nil); err != nil {
		t.Fatal(err)
	}

	videoEncCtx := NewCodecCtx(videoEnc)
	if videoEncCtx == nil {
		t.Fatal("Unable to create video encoder context.")
	}

	if err := videoEncCtx.Open(nil); err != nil {
		t.Fatal(err)
	}

	for packet := range ctx.Packets() {
		if packet.Size() == 0 {
			t.Fatalf("Expected any non zero size.")
		} else {
			// log.Println("Size =", packet.Size(), "Stream index:", packet.StreamIndex())
		}

		stream, err := ctx.GetStream(packet.StreamIndex())
		if err != nil {
			t.Fatal("Unexpected error:", err)
		}

		if stream.CodecCtx().Id() == 0 {
			t.Fatal("Expected any non zero id for codecCtx")
		} else {
			// log.Println("codecCtx Id:", stream.CodecCtx().Id())
		}

		frame, output, err := packet.Decode(stream.CodecCtx())
		if err != nil {
			t.Errorf("Unexpected error: %v\n", err)
		} else {
			// log.Println("Frame data:", frame.Format(), frame.Width(), frame.Height())
		}

		if frame.mediaType == CODEC_TYPE_AUDIO && output > 0 {
			if _, err := frame.Encode(audioEncCtx); err != nil {
				t.Fatal("Unexpected error:", err)
			} else {
				// log.Println(p.avPacket)
			}
		}

		if frame.mediaType == CODEC_TYPE_VIDEO && output > 0 {
			if _, err := frame.Encode(videoEncCtx); err != nil {
				t.Fatal("Unexpected error:", err)
			} else {
				// log.Println(p.avPacket)
			}
		}

		i++
	}

	if i == 0 {
		t.Error("Expected any non zero value")
	} else {
		log.Println(i, "frames encoded")
	}

}
