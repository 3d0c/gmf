/*
2015  Sleepy Programmer <hunan@emsym.com>
*/
package gmf

/*

#cgo pkg-config: libswresample

#include "libswresample/swresample.h"
#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>

int gmf_sw_resample(SwrContext* ctx, AVFrame* dstFrame, AVFrame* srcFrame){
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		(const uint8_t **)srcFrame->data, srcFrame->nb_samples);
}

int gmf_swr_flush(SwrContext* ctx, AVFrame* dstFrame) {
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		NULL, 0);
}

*/
import "C"

import (
	"fmt"
	// "unsafe"
)

type SwrCtx struct {
	swrCtx *C.struct_SwrContext
	cc     *CodecCtx
	CgoMemoryManage
}

func NewSwrCtx(options []*Option, cc *CodecCtx) *SwrCtx {
	this := &SwrCtx{swrCtx: C.swr_alloc(), cc: cc}

	for _, option := range options {
		option.Set(this.swrCtx)
	}

	if ret := int(C.swr_init(this.swrCtx)); ret < 0 {
		fmt.Printf("error swr_init: %s\n", AvError(ret))
		return nil
	}

	return this
}

func (this *SwrCtx) Free() {
	C.swr_free(&this.swrCtx)
}

func (this *SwrCtx) Convert(input *Frame) *Frame {
	if this.cc == nil {
		return nil
	}

	srcSamples := input.NbSamples()
	channels := this.cc.Channels()
	format := this.cc.SampleFmt()
	dstFrame, _ := NewAudioFrame(format, channels, srcSamples)

	C.gmf_sw_resample(this.swrCtx, dstFrame.avFrame, input.avFrame)

	return dstFrame
}

func (this *SwrCtx) Flush(nbSamples int) *Frame {
	dstFrame, _ := NewAudioFrame(this.cc.SampleFmt(), this.cc.Channels(), nbSamples)

	C.gmf_swr_flush(this.swrCtx, dstFrame.avFrame)

	return dstFrame
}
