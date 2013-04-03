package gmf

//#include "libswscale/swscale.h"
//#include <stdint.h>
import "C"
import "unsafe"

type SwsContext struct {
	sws *C.struct_SwsContext
}

func sws_free_context(ctx *SwsContext) {
	C.sws_freeContext(ctx.sws)
}

func sws_scale_getcontext(ctx *SwsContext, srcwidth, srcheight, srcfmt, trgwidth, trgheight, trgfmt, flags int) {
	ctx.sws = C.sws_getContext(C.int(srcwidth), C.int(srcheight), int32(srcfmt), C.int(trgwidth), C.int(trgheight), int32(trgfmt), C.int(flags), nil, nil, nil)
	if ctx.sws == nil {
		println("error!")
	}
}

func sws_scale(ctx *SwsContext, src *Frame, trg *Frame) int {
	in_data := (**C.uint8_t)(unsafe.Pointer(&src.avframe.data))
	in_line := (*_Ctype_int)(unsafe.Pointer(&src.avframe.linesize))
	out_data := (**C.uint8_t)(unsafe.Pointer(&trg.avframe.data))
	out_line := (*_Ctype_int)(unsafe.Pointer(&trg.avframe.linesize))
	result := int(C.sws_scale(ctx.sws,
		in_data,
		in_line,
		0,
		C.int(src.height),
		out_data,
		out_line))
	return result
}
