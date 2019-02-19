package gmf

/*

#cgo pkg-config: libswscale

#include "libswscale/swscale.h"

*/
import "C"

import (
	"fmt"
	"unsafe"
)

var (
	SWS_FAST_BILINEAR int = C.SWS_FAST_BILINEAR
	SWS_BILINEAR      int = C.SWS_BILINEAR
	SWS_BICUBIC       int = C.SWS_BICUBIC
	SWS_X             int = C.SWS_X
	SWS_POINT         int = C.SWS_POINT
	SWS_AREA          int = C.SWS_AREA
	SWS_BICUBLIN      int = C.SWS_BICUBLIN
	SWS_GAUSS         int = C.SWS_GAUSS
	SWS_SINC          int = C.SWS_SINC
	SWS_LANCZOS       int = C.SWS_LANCZOS
	SWS_SPLINE        int = C.SWS_SPLINE
)

type SwsCtx struct {
	swsCtx *C.struct_SwsContext
	width  int
	height int
	pixfmt int32
}

func NewSwsCtx(srcW, srcH int, srcPixFmt int32, dstW, dstH int, dstPixFmt int32, method int) (*SwsCtx, error) {
	ctx := C.sws_getContext(
		C.int(srcW),
		C.int(srcH),
		srcPixFmt,
		C.int(dstW),
		C.int(dstH),
		dstPixFmt,
		C.int(method), nil, nil, nil,
	)

	if ctx == nil {
		return nil, fmt.Errorf("error creating sws context\n")
	}

	return &SwsCtx{
		swsCtx: ctx,
		width:  dstW,
		height: dstH,
		pixfmt: dstPixFmt,
	}, nil
}

func (ctx *SwsCtx) Scale(src *Frame, dst *Frame) {
	C.sws_scale(
		ctx.swsCtx,
		(**C.uint8_t)(unsafe.Pointer(&src.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&src.avFrame.linesize)),
		0,
		C.int(src.Height()),
		(**C.uint8_t)(unsafe.Pointer(&dst.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&dst.avFrame.linesize)))
}

func (ctx *SwsCtx) Free() {
	if ctx.swsCtx != nil {
		C.sws_freeContext(ctx.swsCtx)
	}
}

func DefaultRescaler(ctx *SwsCtx, frames []*Frame) ([]*Frame, error) {
	var (
		result []*Frame = make([]*Frame, 0)
		tmp    *Frame
		err    error
	)

	for i, _ := range frames {
		tmp = NewFrame().SetWidth(ctx.width).SetHeight(ctx.height).SetFormat(ctx.pixfmt)
		if err = tmp.ImgAlloc(); err != nil {
			return nil, fmt.Errorf("error allocation tmp frame - %s", err)
		}

		ctx.Scale(frames[i], tmp)

		tmp.SetPts(frames[i].Pts())
		tmp.SetPktDts(frames[i].PktDts())

		result = append(result, tmp)
	}

	for i := 0; i < len(frames); i++ {
		if frames[i] != nil {
			frames[i].Free()
		}
	}

	return result, nil
}
