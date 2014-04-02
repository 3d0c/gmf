package main

/*
#cgo pkg-config: libavformat libavutil

#include "libavutil/frame.h"
#include "libavformat/avformat.h"
#include "libavformat/avio.h"

void set_data(AVFrame *frame, int idx, int y, int x, uint8_t data) {
    if(!frame) {
        fprintf(stderr, "frame is NULL\n");
    }

    if(!frame->linesize[idx]) {
        fprintf(stderr, "wrong index, %d\n", idx);
    }

    frame->data[idx][y * frame->linesize[idx] + x] = data;
}

*/
import "C"

import (
	"errors"
	"fmt"
	. "github.com/3d0c/gmf"
	"log"
	"os"
)

func fatal(err error) {
	log.Fatal(err)
	os.Exit(0)
}

func videoEncodeExample() {
	codec, err := NewEncoder(AV_CODEC_ID_MPEG1VIDEO)
	// codec, err := NewEncoder("mpeg4")
	if err != nil {
		fatal(err)
	}

	videoEncCtx := NewCodecCtx(codec)
	if videoEncCtx == nil {
		fatal(err)
	}

	outputCtx, err := NewOutputCtx("./test-ctx.mpg")
	if err != nil {
		fatal(err)
	}

	videoEncCtx.SetBitRate(400000)
	videoEncCtx.SetWidth(352)
	videoEncCtx.SetHeight(288)
	videoEncCtx.SetTimeBase(AVR{1, 25})
	videoEncCtx.SetGopSize(10)
	videoEncCtx.SetMaxBFrames(1)
	videoEncCtx.SetPixFmt(AV_PIX_FMT_YUV420P)

	// videoEncCtx.SetProfile(C.FF_PROFILE_MPEG4_SIMPLE)

	if outputCtx.IsGlobalHeader() {
		videoEncCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
		log.Println("AVFMT_GLOBALHEADER flag is set.")
	}

	videoStream := outputCtx.NewStream(codec, nil)
	if videoStream == nil {
		fatal(errors.New(fmt.Sprintf("Unable to create stream for videoEnc [%s]\n", codec.LongName())))
	}

	if err := videoEncCtx.Open(nil); err != nil {
		fatal(err)
	}

	if err := videoStream.SetCodecCtx(videoEncCtx); err != nil {
		fatal(err)
	}

	// This is what WriteHeader() does:

	// if ret := int(C.avio_open(&((*_Ctype_AVFormatContext)(outputCtx.AvPtr()).pb), C.CString("./test-ctx.mpg"), C.AVIO_FLAG_WRITE)); ret < 0 {
	// 	fatal(errors.New("avio_open error"))
	// }

	// if ret := int(C.avformat_write_header((*_Ctype_AVFormatContext)(outputCtx.AvPtr()), nil)); ret < 0 {
	// 	fatal(errors.New("write_header error"))
	// }

	// (*_Ctype_AVFormatContext)(outputCtx.AvPtr()).start_time = 0
	// (*_Ctype_AVFormatContext)(outputCtx.AvPtr()).duration = 0
	// (*_Ctype_AVFormatContext)(outputCtx.AvPtr()).duration_estimation_method = C.AVFMT_DURATION_FROM_PTS

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	// We're creating dummy image by hand, so we should create
	// frame too. Generally, packet.Decode() returns frame.
	frame := NewFrame()
	frame.SetWidth(videoEncCtx.Width())
	frame.SetHeight(videoEncCtx.Height())
	frame.SetFormat(videoEncCtx.PixFmt())

	if err := frame.ImgAlloc(); err != nil {
		fatal(err)
	}

	i := 0

	for i = 0; i < 25; i++ {
		for y := 0; y < videoEncCtx.Height(); y++ {
			for x := 0; x < videoEncCtx.Width(); x++ {
				C.set_data((*_Ctype_AVFrame)(frame.AvPtr()), 0, C.int(y), C.int(x), C.uint8_t(x+y+i*3))
			}
		}

		// Cb and Cr
		for y := 0; y < videoEncCtx.Height()/2; y++ {
			for x := 0; x < videoEncCtx.Width()/2; x++ {
				C.set_data((*_Ctype_AVFrame)(frame.AvPtr()), 1, C.int(y), C.int(x), C.uint8_t(128+y+i*2))
				C.set_data((*_Ctype_AVFrame)(frame.AvPtr()), 2, C.int(y), C.int(x), C.uint8_t(64+x+i*5))
			}
		}

		frame.SetPts(i)

		if p, ready, err := frame.Encode(videoStream.GetCodecCtx()); ready {
			log.Println("pkt[orig].Pts/Dts/Size:", p.Pts(), p.Dts(), p.Size())

			if p.Pts() != AV_NOPTS_VALUE {
				p.SetPts(RescaleQ(p.Pts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if p.Dts() != AV_NOPTS_VALUE {
				p.SetDts(RescaleQ(p.Dts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				fatal(err)
			}

			log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())

		} else if err != nil {
			fatal(err)
		}
	}

	frame.SetPts(i)

	if p, ready, _ := frame.Encode(videoStream.GetCodecCtx()); ready {
		p.SetPts(RescaleQ(p.Pts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))

		p.SetDts(RescaleQ(p.Dts(), videoStream.GetCodecCtx().TimeBase(), videoStream.TimeBase()))

		if err := outputCtx.WritePacket(p); err != nil {
			fatal(err)
		}
		log.Printf("Write frame=%d size=%v pts=%v dts=%v\n", i, p.Size(), p.Pts(), p.Dts())
	}

	outputCtx.CloseOutput()
}

func main() {
	videoEncodeExample()
}
