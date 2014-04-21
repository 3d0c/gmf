package gmf

/*

#cgo pkg-config: libswresample

#include "libavutil/opt.h"

*/
import "C"

import (
	"log"
	"unsafe"
)

type Options struct{}

type Option struct {
	Key string
	Val interface{}
}

func (this *Option) Set(ctx interface{}) {
	ckey := C.CString(this.Key)
	defer C.free(unsafe.Pointer(ckey))

	switch t := this.Val.(type) {
	case int:
		C.av_opt_set_int((unsafe.Pointer)(ctx.(*[0]uint8)), ckey, C.int64_t(this.Val.(int)), 0)
	case SampleFmt:
		C.av_opt_set_sample_fmt((unsafe.Pointer)(ctx.(*[0]uint8)), ckey, (int32)(this.Val.(SampleFmt)), 0)
	default:
		log.Println("unsupported type:", t)
	}
}
