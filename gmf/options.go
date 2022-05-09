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

type Option struct {
	Key string
	Val interface{}
}

func (this Option) Set(ctx interface{}) {
	ckey := C.CString(this.Key)
	defer C.free(unsafe.Pointer(ckey))

	var ret int = 0

	switch t := this.Val.(type) {
	case int:
		ret = int(C.av_opt_set_int(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, C.int64_t(this.Val.(int)), C.AV_OPT_SEARCH_CHILDREN))

	case []int32:
		ret = int(C.av_opt_set_bin(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (*C.uchar)(unsafe.Pointer(&this.Val.([]int)[0])), C.int(this.Val.([]int)[0]), C.AV_OPT_SEARCH_CHILDREN))

	case int32:
		ret = int(C.av_opt_set_int(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, C.int64_t(this.Val.(int32)), C.AV_OPT_SEARCH_CHILDREN))

	case SampleFormat:
		ret = int(C.av_opt_set_sample_fmt(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (int32)(this.Val.(SampleFormat)), C.AV_OPT_SEARCH_CHILDREN))

	case float64:
		ret = int(C.av_opt_set_double(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (C.double)(this.Val.(float64)), C.AV_OPT_SEARCH_CHILDREN))

	case AVR:
		ret = int(C.av_opt_set_q(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, C.AVRational(this.Val.(AVR).AVRational()), C.AV_OPT_SEARCH_CHILDREN))

	case []byte:
		ret = int(C.av_opt_set_bin(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, (*C.uint8_t)(unsafe.Pointer(&this.Val.([]byte)[0])), C.int(len(ctx.([]byte))), C.AV_OPT_SEARCH_CHILDREN))

	case string:
		cval := C.CString(this.Val.(string))
		defer C.free(unsafe.Pointer(cval))
		ret = int(C.av_opt_set(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), ckey, cval, C.AV_OPT_SEARCH_CHILDREN))

	case *Dict:
		ret = int(C.av_opt_set_dict(unsafe.Pointer(reflect.ValueOf(ctx).Pointer()), &this.Val.(*Dict).avDict))

	default:
		log.Println("unsupported type:", t)
	}

	if ret < 0 {
		log.Printf("unable to set key '%s' value '%v', error: %s\n", this.Key, this.Val, AvError(int(ret)))
	}
}
