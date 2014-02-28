package gmf

/*
#cgo pkg-config: libavcodec libavutil

#include "libavutil/error.h"

*/
import "C"

import (
	"bytes"
	"errors"
	"unsafe"
)

func AvError(averr int) error {
	errlen := 1024
	b := make([]byte, errlen)

	C.av_strerror(C.int(averr), (*C.char)(unsafe.Pointer(&b[0])), C.size_t(errlen))

	return errors.New(string(b[:bytes.Index(b, []byte{0})]))
}
