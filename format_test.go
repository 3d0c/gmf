package gmf

import (
	"log"
	// "os"
	// "fmt"
	"testing"
)

var (
	testVideoFile   = "tmp/v.mp4"
	testVideoOutput = "tmp/out.mp4"
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

		if stream.GetCodecCtx().Id() == 0 {
			t.Fatal("Expected any non zero id for codecCtx")
		} else {
			// log.Println("codecCtx Id:", stream.GetCodecCtx().Id())
		}

		_, _, err = packet.Decode(stream.GetCodecCtx())
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

	if err := outCtx.OpenOutput(NewOutputFmt("mpeg", testVideoOutput, "")); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("Output file %s opened", testVideoOutput)
	}

	i := 0
	w := 0
	// audioEnc, err := NewEncoder("mp2")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	videoEnc, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	// audioStream := outCtx.NewStream(audioEnc, nil)
	// if audioStream == nil {
	// 	t.Fatalf("Unable to create stream for audioEnc [%s]\n", audioEnc.LongName())
	// } else {
	// 	log.Println("Stream for", audioEnc.LongName(), "created")
	// }

	videoStream := outCtx.NewStream(videoEnc, nil)
	if videoStream == nil {
		t.Fatalf("Unable to create stream for videoEnc [%s]\n", videoEnc.LongName())
	} else {
		log.Println("Stream for", videoEnc.LongName(), "created")
	}

	if cnt := outCtx.StreamsCnt(); cnt != 1 {
		t.Fatalf("Expected stream count in output context = 1, %d got", cnt)
	} else {
		log.Println("1 streams created in output context")
	}

	// audioEncCtx := NewCodecCtx(audioEnc)
	// if audioEncCtx == nil {
	// 	t.Fatal("Unable to create audio encoder context.")
	// }

	// if err := audioEncCtx.Open(nil); err != nil {
	// 	t.Fatal(err)
	// }

	videoEncCtx := NewCodecCtx(videoEnc)
	if videoEncCtx == nil {
		t.Fatal("Unable to create video encoder context.")
	}

	if err := videoEncCtx.Open(nil); err != nil {
		t.Fatal(err)
	}

	// if err := audioStream.SetCodecCtx(audioEncCtx); err != nil {
	// 	t.Fatal("Unexpected error:", err)
	// } else {
	// 	log.Println("Audio context is set")
	// }

	if err := videoStream.SetCodecCtx(videoEncCtx); err != nil {
		t.Fatal("Unexpected error:", err)
	} else {
		log.Println("Video context is set")
	}

	if err := outCtx.WriteHeader(); err != nil {
		t.Fatal("Unexpected error:", err)
	} else {
		log.Println("Header is written to:", outCtx.ofmt.Filename)
	}

	outCtx.Dump()

	inputVideo, err := ctx.GetVideoStream()
	if err != nil {
		t.Fatal(err)
	} else {
		inputVideo.GetCodecCtx().SetOpt()
	}

	// image, err := NewImage(
	// 	inputVideo.GetCodecCtx().Width(),
	// 	inputVideo.GetCodecCtx().Height(),
	// 	inputVideo.GetCodecCtx().PixFmt(),
	// 	1,
	// )
	// if err != nil {
	// 	t.Fatal(err)
	// }

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

		if stream.GetCodecCtx().Type() == CODEC_TYPE_AUDIO {
			// skip for tests
			continue
		}

		if stream.GetCodecCtx().Id() == 0 {
			t.Fatal("Expected any non zero id for codecCtx")
		} else {
			// log.Println("codecCtx Id:", stream.GetCodecCtx().Id())
		}

		// frame, output, err := packet.Decode(stream.GetCodecCtx())
		// if err != nil {
		// 	t.Errorf("Unexpected error: %v\n", err)
		// } else {
		// 	// log.Println("Frame data:", frame.Format(), frame.Width(), frame.Height())
		// }

		// if frame.mediaType == CODEC_TYPE_AUDIO && output > 0 {
		// if p, err := frame.Encode(audioEncCtx); err != nil {
		// 	t.Fatal("Unexpected error:", err)
		// } else {
		// 	if err := outCtx.WritePacket(p); err != nil {
		// 		t.Fatal(err)
		// 	}
		// }
		// }

		// v1
		// f := 0
		// for frame := range packet.DecodeV(stream.GetCodecCtx()) {
		// 	// log.Println("f:", f)
		// 	if frame.mediaType != CODEC_TYPE_VIDEO {
		// 		t.Fatal("Wrong frame.mediaType")
		// 	}

		// 	frame.SetPts(videoStream.RescaleTimestamp())
		// 	if p, ready, err := frame.Encode(videoEncCtx); ready {
		// 		if err := outCtx.WritePacket(p); err != nil {
		// 			t.Fatal(err)
		// 		}
		// 		// break

		// 	} else if err != nil {
		// 		t.Fatal(err)
		// 	} else if !ready {
		// 		log.Println("!ready")
		// 	}
		// 	f++
		// }

		log.Println("orig packet:", packet.Dts(), packet.Pts(), packet.Duration())
		if frame := packet.DecodeV2(stream.GetCodecCtx()); frame != nil {
			i++
			// image.Copy(frame)
			log.Println("frame:", frame.TimeStamp(), frame.PktPos(), frame.PktDuration())
			if p, ready, err := frame.Encode(videoEncCtx); ready {
				w++
				log.Println("packet:", p.Dts(), p.Pts(), p.Duration())
				if err := outCtx.WritePacket(p); err != nil {
					t.Fatal(err)
				}
			} else if err != nil {
				t.Fatal(err)
			}

			frame.Unref()
		}

		// if frame.mediaType == CODEC_TYPE_VIDEO && output > 0 {
		// 	frame.SetPts(videoStream.RescaleTimestamp())
		// 	if p, ready, err := frame.Encode(videoEncCtx); ready > 0 {
		// 		if err := outCtx.WritePacket(p); err != nil {
		// 			t.Fatal(err)
		// 		}

		// 	} else if err != nil {
		// 		t.Fatal(err)
		// 	}
		// }

		// log.Println("i:", i)
	}
	// image.Free()

	outCtx.CloseOutput()

	if i == 0 {
		t.Error("Expected any non zero value")
	} else {
		log.Println(i, "frames encoded", w, "written")
	}

}
