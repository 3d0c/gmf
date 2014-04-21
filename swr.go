package gmf

/*

#cgo pkg-config: libswresample

#include "libswresample/swresample.h"

*/
import "C"

type SwrCtx struct {
	swrCtx *C.struct_SwrContext
}

func NewSwrCtx(options []*Option) *SwrCtx {
	this := &SwrCtx{swrCtx: C.swr_alloc()}

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
