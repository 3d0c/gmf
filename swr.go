package gmf

/*

#cgo pkg-config: libswresample

#include "libswresample/swresample.h"

*/
import "C"

//
// Unfinished.
//

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

	if int(C.swr_init(this.swrCtx)) < 0 {
		return nil
	}

	return this
}

func (this *SwrCtx) Free() {
	C.swr_free(&this.swrCtx)
}

func (this *SwrCtx) Convert(input *Frame) *Frame {
	panic("This stuff is unfinished.")
	// frame := NewFrame()

	// dstNbSamples := C.av_rescale_rnd(C.swr_get_delay(this.swrCtx, this.cc.avCodecCtx.sample_rate)+input.avFrame.nb_samples, C.int64_t(this.cc.SampleRate()), this.cc.SampleRate(), C.AV_ROUND_UP)

	// if dstNbSamples > input.NbSamples() {
	// }

	return nil
}
