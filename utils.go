package gmf

/*
#cgo pkg-config: libavcodec libavutil

#include "libavutil/avutil.h"
#include "libavutil/error.h"
#include "libavutil/mathematics.h"
#include "libavutil/rational.h"

*/
import "C"

import (
	"bytes"
	"errors"
	"unsafe"
)

type AVRational _Ctype_AVRational

type AVR struct {
	Num int
	Den int
}

var (
	AV_TIME_BASE   int        = C.AV_TIME_BASE
	AV_TIME_BASE_Q AVRational = AVRational{1, C.int(AV_TIME_BASE)}
)

func AvError(averr int) error {
	errlen := 1024
	b := make([]byte, errlen)

	C.av_strerror(C.int(averr), (*C.char)(unsafe.Pointer(&b[0])), C.size_t(errlen))

	return errors.New(string(b[:bytes.Index(b, []byte{0})]))
}

func RescaleQ(a int, encBase AVRational, stBase AVRational) int {
	return int(C.av_rescale_q(C.int64_t(a), _Ctype_AVRational(encBase), _Ctype_AVRational(stBase)))
}

func assert(i interface{}, err error) interface{} {
	if err != nil {
		panic(err)
	}

	return i
}
