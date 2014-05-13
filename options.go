package gmf

/*

#cgo pkg-config: libavutil

#include <stdlib.h>
#include "libavutil/opt.h"

*/
import "C"

import (
	"log"
	"reflect"
	"unsafe"
)

// @todo remove from code
type Options struct{}

type Option struct {
	Key string
	Val interface{}
}

func (this *Option) Set(ctx interface{}) {
	ckey := C.CString(this.Key)
	defer C.free(unsafe.Pointer(ckey))

	var ret int = 0

	switch t := this.Val.(type) {
	case int:
		ret = int(C.av_opt_set_int(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, C.int64_t(this.Val.(int)), 0))

	case SampleFmt:
		ret = int(C.av_opt_set_sample_fmt(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (int32)(this.Val.(SampleFmt)), 0))

	case float64:
		ret = int(C.av_opt_set_double(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (C.double)(this.Val.(float64)), 0))

	case AVR:
		ret = int(C.av_opt_set_q(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, C.AVRational(this.Val.(AVR).AVRational()), 0))

	case []byte:
		ret = int(C.av_opt_set_bin(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (*C.uint8_t)(unsafe.Pointer(&this.Val.([]byte)[0])), C.int(len(ctx.([]byte))), 0))

	case *Dict:
		ret = int(C.av_opt_set_dict(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), &this.Val.(*Dict).avDict))

	default:
		log.Println("unsupported type:", t)
	}

	if ret < 0 {
		log.Printf("unable to set key '%s' value '%d', error: %s\n", this.Key, this.Val.(int), AvError(int(ret)))
	}
}
