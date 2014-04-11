package gmf

import (
	"log"
	"os"
	// "fmt"
	"testing"
)

var (
	testVideoFile   = "tmp/src2s.mp4"
	testVideoOutput = "tmp/out.mp4"
	// testVideoOutput = "tmp/out.mpg"
)

func assert(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}

func _TestFmtCtx(t *testing.T) {
	ctx := NewCtx()

	if ctx.avCtx == nil {
		t.Fatal("AVContenxt is not initialized")
	}
}

func _TestStreamFail(t *testing.T) {
	ctx := NewCtx()

	if stream, err := ctx.GetStream(1); err != nil {
		log.Printf("Expected error got: %v", err)
	} else {
		t.Errorf("Expected error, stream with id %d got", stream.Id())
	}

	ctx.Free()
}

func _TestOutputCtx(t *testing.T) {
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

func _TestOutputCtx_shouldfail(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	if err := ctx.SetOformat(NewOutputFmt("-invalid-", "", "")); err == nil {
		t.Fatalf("Error is expected.")
	}
}

func _TestNewStream(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	if err := ctx.SetOformat(NewOutputFmt("mpeg", "", "")); err != nil {
		t.Fatalf("Unable to set 'oformat' to 'context'. Error:", err)
	}

	for i := 0; i < 2; i++ {
		stream := ctx.NewStream(assert(NewEncoder("mpeg4")).(*Codec))
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

func _TestOpenInput(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	err := ctx.OpenInput(testVideoFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("File '%s' opened.\n", testVideoFile)
	}
}

func _TestPacketsIterator(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	err := ctx.OpenInput(testVideoFile)
	if err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("File %s contains %d streams.\n", testVideoFile, ctx.StreamsCnt())
	}
	log.Println("start time:", ctx.StartTime())
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

		if stream.GetCodecCtx().Type() == AVMEDIA_TYPE_AUDIO {
			// skip for tests
			continue
		}

		if stream.GetCodecCtx().Id() == 0 {
			t.Fatal("Expected any non zero id for codecCtx")
		} else {
			// log.Println("codecCtx Id:", stream.GetCodecCtx().Id())
		}

		packet.SetPts(packet.Pts() + RescaleQ(ctx.TsOffset(ctx.StartTime()), AV_TIME_BASE_Q, stream.TimeBase()))

		log.Println("packet.Pts:", packet.Pts(), "duration:", packet.Duration(), "size:", packet.Size())

		frame, got, err := packet.Decode(stream.GetCodecCtx())
		if got != 0 {
			frame.SetBestPts()
			log.Println("frame. pktduration:", frame.PktDuration(), "pktpos:", frame.PktPos(), "pts:", frame.Pts(), "w:", frame.Width(), "h:", frame.Height(), "keyframe:", frame.KeyFrame())
		}

		if got == 0 || err != nil {
			log.Println("err:", err, "got:", got)
		}

		frame.Unref() // it could fail. if it does, try to move it into 'got != 0' block

		i++
		if i > 5 {
			os.Exit(0)
		}
	}

	if i == 0 {
		t.Error("Expected any non zero value")
	} else {
		log.Println(i, "frames iterated.")
	}
}

func _TestEncode(t *testing.T) {
	ctx := NewCtx()
	defer ctx.Free()

	if err := ctx.OpenInput(testVideoFile); err != nil {
		t.Fatalf("Unexpected error: %v\n", err)
	} else {
		log.Printf("Input file %s opened. %d streams found.\n", testVideoFile, ctx.StreamsCnt())
	}

	inputVideo, err := ctx.GetBestStream(AV_TIME_BASE)
	if err != nil {
		t.Fatal(err)
	} else {
		inputVideo.GetCodecCtx().SetOpt()
	}

	outCtx, err := NewOutputCtx(testVideoOutput)
	if err != nil {
		t.Fatal(err)
	}

	i := 0
	w := 0
	// audioEnc, err := NewEncoder("mp2")
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// videoEnc, err := NewEncoder(AV_CODEC_ID_MPEG1VIDEO)
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

	videoEncCtx.CopyCtx(assert(ctx.GetBestStream(AVMEDIA_TYPE_VIDEO)).(*Stream))

	if outCtx.IsGlobalHeader() {
		videoEncCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
		log.Println("AVFMT_GLOBALHEADER flag is set.")
	}

	// copy profile from source
	videoEncCtx.SetProfile(inputVideo.GetCodecCtx().GetProfile())

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

		if stream.GetCodecCtx().Type() == AVMEDIA_TYPE_AUDIO {
			// skip for tests
			continue
		}

		if stream.GetCodecCtx().Id() == 0 {
			t.Fatal("Expected any non zero id for codecCtx")
		} else {
			// log.Println("codecCtx Id:", stream.GetCodecCtx().Id())
		}

		frame, got, err := packet.Decode(stream.GetCodecCtx())
		if got != 0 {
			// frame.SetBestPts()
			frame.SetPts(i)
			log.Println("---frame---")
			log.Println("pkt_duration:", frame.PktDuration(), "pkt_pos:", frame.PktPos(), "pts:", frame.Pts(), "w:", frame.Width(), "h:", frame.Height(), "key_frame:", frame.KeyFrame())
			// log.Println(frame.avFrame)
			if p, ready, _ := frame.Encode(videoEncCtx); ready {
				if outSt, err := outCtx.GetStream(0); err == nil {
					if p.Pts() != AV_NOPTS_VALUE {
						p.SetPts(RescaleQ(p.Pts(), outSt.GetCodecCtx().TimeBase(), stream.TimeBase()))
					}

					if p.Dts() != AV_NOPTS_VALUE {
						p.SetDts(RescaleQ(p.Dts(), outSt.GetCodecCtx().TimeBase(), stream.TimeBase()))
					}
				}

				log.Println("---packet out---")
				log.Println("size:", p.Size(), "pts:", p.Pts(), "duration:", p.Duration())
				w++
				if err := outCtx.WritePacket(p); err != nil {
					t.Fatal(err)
				}
			}
		}

		if got == 0 || err != nil {
			log.Println("err:", err, "got:", got)
		}

		frame.Unref()

		i++
		// if i > 6 {
		// 	break
		// }
	}

	log.Println("output ctx duration:", outCtx.Duration())

	outCtx.CloseOutput()

	log.Println(i, "frames encoded", w, "written")
}
