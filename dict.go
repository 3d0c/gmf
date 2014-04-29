package gmf

/*

#cgo pkg-config: libavutil

#include "libavutil/dict.h"

*/
import "C"

import (
	"log"
)

type Pair struct {
	Key string
	Val string
}

type Dict struct {
	avDict *C.struct_AVDictionary
}

func NewDict(pairs []Pair) *Dict {
	this := &Dict{avDict: nil}

	for _, pair := range pairs {
		ckey := C.CString(pair.Key)
		cval := C.CString(pair.Val)

		if ret := C.av_dict_set(&this.avDict, ckey, cval, 0); int(ret) < 0 {
			log.Printf("unable to set key '%s' value '%d', error: %s\n", pair.Key, pair.Val, AvError(int(ret)))
		}

		// C.free(unsafe.Pointer(ckey))
		// C.free(unsafe.Pointer(cval))
	}

	return this
}
