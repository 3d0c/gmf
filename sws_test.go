package gmf

import (
	"log"
	"testing"
)

// @todo rewrite it
func _TestScale(t *testing.T) {
	srcWidth, srcHeight := 640, 480
	dstWidth, dstHeight := 320, 200

	// codec, err := NewEncoder(AV_CODEC_ID_MPEG1VIDEO)
	codec, err := NewEncoder("mpeg4")
	if err != nil {
		t.Fatal(err)
	}

	srcEncCtx := NewCodecCtx(codec)
	if srcEncCtx == nil {
		t.Fatal("Unable to allocate codec context")
	}
	srcEncCtx.SetWidth(640).SetHeight(480).SetPixFmt(AV_PIX_FMT_YUV420P)

	dstCodecCtx := NewCodecCtx(codec)
	if dstCodecCtx == nil {
		t.Fatal("Unable to allocate codec context")
	}

	dstCodecCtx.SetBitRate(400000).SetWidth(dstWidth).SetHeight(dstHeight).SetTimeBase(AVR{1, 25}).SetGopSize(10).SetMaxBFrames(1).SetPixFmt(AV_PIX_FMT_YUV420P)

	dstCodecCtx.SetProfile(FF_PROFILE_MPEG4_SIMPLE)

	outputCtx, err := NewOutputCtx("tmp/test-ctx.mp4")
	if err != nil {
		t.Fatal(err)
	}

	if outputCtx.IsGlobalHeader() {
		dstCodecCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
		log.Println("AVFMT_GLOBALHEADER flag is set.")
	}

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		t.Fatalf("Unable to create stream for videoEnc [%s]\n", codec.LongName())
	}

	if err := dstCodecCtx.Open(nil); err != nil {
		t.Fatal(err)
	}

	videoStream.SetCodecCtx(dstCodecCtx)

	if err := outputCtx.WriteHeader(); err != nil {
		t.Fatal(err)
	}

	swsCtx := NewSwsCtx(srcEncCtx, dstCodecCtx, SWS_BICUBIC)

	dstFrame := NewFrame().SetWidth(dstWidth).SetHeight(dstHeight).SetFormat(AV_PIX_FMT_YUV420P)

	if err := dstFrame.ImgAlloc(); err != nil {
		t.Fatal(err)
	}

	var frame *Frame

	i := 0
	for frame = range GenSyntVideo(srcWidth, srcHeight, srcEncCtx.PixFmt()) {
		frame.SetPts(i)

		swsCtx.Scale(frame, dstFrame)

		if p, ready, err := dstFrame.Encode(videoStream.CodecCtx()); ready {
			log.Println("pkt[orig].Pts/Dts/Size:", p.Pts(), p.Dts(), p.Size())

			if p.Pts() != AV_NOPTS_VALUE {
				p.SetPts(RescaleQ(p.Pts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if p.Dts() != AV_NOPTS_VALUE {
				p.SetDts(RescaleQ(p.Dts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				t.Fatal(err)
			}

			log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())

		} else if err != nil {
			t.Fatal(err)
		}

		i++
	}

	frame.SetPts(i)

	if p, ready, _ := frame.Encode(videoStream.CodecCtx()); ready {
		p.SetPts(RescaleQ(p.Pts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))

		p.SetDts(RescaleQ(p.Dts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))

		if err := outputCtx.WritePacket(p); err != nil {
			t.Fatal(err)
		}
		log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())
	}

	outputCtx.CloseOutput()
}
