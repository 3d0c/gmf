package gmf

/*

#cgo pkg-config: libswscale

#include "libswscale/swscale.h"

*/
import "C"

import (
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
	Width  int
	Height int
	PixFmt int32
	CgoMemoryManage
}

func NewSwsCtx(src *CodecCtx, dst *CodecCtx, method int) *SwsCtx {
	ctx := C.sws_getContext(C.int(src.Width()), C.int(src.Height()), src.PixFmt(), C.int(dst.Width()), C.int(dst.Height()), dst.PixFmt(), C.int(method), nil, nil, nil)

	if ctx == nil {
		return nil
	}

	return &SwsCtx{
		swsCtx: ctx,
		Width:  dst.Width(),
		Height: dst.Height(),
		PixFmt: dst.PixFmt(),
	}
}

func NewPicSwsCtx(srcWidth int, srcHeight int, srcPixFmt int32, dst *CodecCtx, method int) *SwsCtx {
	ctx := C.sws_getContext(C.int(srcWidth), C.int(srcHeight), srcPixFmt, C.int(dst.Width()), C.int(dst.Height()), dst.PixFmt(), C.int(method), nil, nil, nil)

	if ctx == nil {
		return nil
	}

	return &SwsCtx{swsCtx: ctx}
}

func (this *SwsCtx) Free() {
	C.sws_freeContext(this.swsCtx)
}

func (this *SwsCtx) Scale(src *Frame, dst *Frame) {
	C.sws_scale(
		this.swsCtx,
		(**C.uint8_t)(unsafe.Pointer(&src.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&src.avFrame.linesize)),
		0,
		C.int(src.Height()),
		(**C.uint8_t)(unsafe.Pointer(&dst.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&dst.avFrame.linesize)))
}

func (this *SwsCtx) Scale2(src *Frame) (*Frame, error) {
	dst := NewFrame().SetWidth(this.Width).SetHeight(this.Height).SetFormat(this.PixFmt)

	if err := dst.ImgAlloc(); err != nil {
		return nil, err
	}

	C.sws_scale(
		this.swsCtx,
		(**C.uint8_t)(unsafe.Pointer(&src.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&src.avFrame.linesize)),
		0,
		C.int(src.Height()),
		(**C.uint8_t)(unsafe.Pointer(&dst.avFrame.data)),
		(*_Ctype_int)(unsafe.Pointer(&dst.avFrame.linesize)))

	return dst, nil
}
