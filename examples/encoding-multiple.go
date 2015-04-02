package main

// BUG(3d0c):
// 1.
//   Insufficient thread locking around avcodec_open/close()
//   [NULL @ 0x6001200] No lock manager is set, please see av_lockmgr_register()
//   Assertion ff_avcodec_locked failed at libavcodec/utils.c:3271
//
// 2. Last frame isn't processed

import (
	"errors"
	"fmt"
	"log"
	"sync"

	. "github.com/3d0c/gmf"
)

func fatal(err error) {
	log.Fatal(err)
}

type output struct {
	filename string
	codec    interface{}
	data     chan *Frame
}

func encodeWorker(o output, wg *sync.WaitGroup) {
	defer wg.Done()

	codec, err := FindEncoder(o.codec)
	if err != nil {
		fatal(err)
	}

	videoEncCtx := NewCodecCtx(codec)
	if videoEncCtx == nil {
		fatal(err)
	}
	defer Release(videoEncCtx)

	outputCtx, err := NewOutputCtx(o.filename)
	if err != nil {
		fatal(err)
	}

	videoEncCtx.
		SetBitRate(400000).
		SetWidth(320).
		SetHeight(200).
		SetTimeBase(AVR{1, 25}).
		SetPixFmt(AV_PIX_FMT_YUV420P)

	if o.codec == AV_CODEC_ID_MPEG1VIDEO {
		videoEncCtx.SetMbDecision(FF_MB_DECISION_RD)
	}

	if o.codec == AV_CODEC_ID_MPEG4 {
		videoEncCtx.SetProfile(FF_PROFILE_MPEG4_SIMPLE)
	}

	if outputCtx.IsGlobalHeader() {
		videoEncCtx.SetFlag(CODEC_FLAG_GLOBAL_HEADER)
	}

	videoStream := outputCtx.NewStream(codec)
	if videoStream == nil {
		fatal(errors.New(fmt.Sprintf("Unable to create stream for videoEnc [%s]\n", codec.LongName())))
	}
	defer Release(videoStream)

	if err := videoEncCtx.Open(nil); err != nil {
		fatal(err)
	}

	videoStream.SetCodecCtx(videoEncCtx)

	outputCtx.SetStartTime(0)

	if err := outputCtx.WriteHeader(); err != nil {
		fatal(err)
	}

	i, w := 0, 0

	for {
		frame, ok := <-o.data
		if !ok {
			break
		}

		if p, ready, err := frame.EncodeNewPacket(videoStream.CodecCtx()); ready {
			if p.Pts() != AV_NOPTS_VALUE {
				p.SetPts(RescaleQ(p.Pts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if p.Dts() != AV_NOPTS_VALUE {
				p.SetDts(RescaleQ(p.Dts(), videoStream.CodecCtx().TimeBase(), videoStream.TimeBase()))
			}

			if err := outputCtx.WritePacket(p); err != nil {
				fatal(err)
			} else {
				w++
			}
			Release(p)
		} else if err != nil {
			fatal(err)
		}

		Release(frame)
		i++
	}

	outputCtx.CloseOutputAndRelease()

	log.Printf("done [%s], %d frames, %d written\n", o.filename, i, w)
}

func main() {
	o := []output{
		{"sample-enc-mpeg1.mpg", AV_CODEC_ID_MPEG1VIDEO, make(chan *Frame)},
		{"sample-enc-mpeg2.mpg", AV_CODEC_ID_MPEG2VIDEO, make(chan *Frame)},
		{"sample-enc-mpeg4.mp4", AV_CODEC_ID_MPEG4, make(chan *Frame)},
	}

	wg := new(sync.WaitGroup)
	wCount := 0

	for _, item := range o {
		wg.Add(1)
		go encodeWorker(item, wg)
		wCount++
	}

	var srcFrame *Frame
	j := 0

	for srcFrame = range GenSyntVideoNewFrame(320, 200, AV_PIX_FMT_YUV420P) {
		srcFrame.SetPts(j)
		for i := 0; i < wCount; i++ {
			Retain(srcFrame)
			o[i].data <- srcFrame
		}
		j += 1
		Release(srcFrame)
	}

	for _, item := range o {
		close(item.data)
	}

	wg.Wait()
}
