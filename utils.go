package gmf

/*

#cgo pkg-config: libavcodec libavutil

#include "libavutil/avutil.h"
#include "libavutil/error.h"
#include "libavutil/mathematics.h"
#include "libavutil/rational.h"
#include "libavutil/samplefmt.h"

*/
import "C"

import (
	"bytes"
	"errors"
	"unsafe"
)

type AVRational C.struct_AVRational

type AVR struct {
	Num int
	Den int
}

func (this AVR) AVRational() AVRational {
	return AVRational{C.int(this.Num), C.int(this.Den)}
}

func (this AVRational) AVR() AVR {
	return AVR{Num: int(this.num), Den: int(this.den)}
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
	return int(C.av_rescale_q(C.int64_t(a), C.struct_AVRational(encBase), C.struct_AVRational(stBase)))
}

func CompareTimeStamp(aTimestamp int, aTimebase AVRational, bTimestamp int, bTimebase AVRational) int {
	return int(C.av_compare_ts(C.int64_t(aTimestamp), C.struct_AVRational(aTimebase),
		C.int64_t(bTimestamp), C.struct_AVRational(bTimebase)))
}
func RescaleDelta(inTb AVRational, inTs int, fsTb AVRational, duration int, last *int, outTb AVRational) int {
	return int(C.av_rescale_delta(C.struct_AVRational(inTb), C.int64_t(inTs), C.struct_AVRational(fsTb), C.int(duration), (*C.int64_t)(unsafe.Pointer(&last)), C.struct_AVRational(outTb)))
}

func Rescale(a, b, c int) int {
	return int(C.av_rescale(C.int64_t(a), C.int64_t(b), C.int64_t(c)))
}

func GetSampleFmtName(fmt int32) string {
	return C.GoString(C.av_get_sample_fmt_name(fmt))
}

// Synthetic video generator. It produces 25 iteratable frames.
// Used for tests.
func GenSyntVideoNewFrame(w, h int, fmt int32) chan *Frame {
	yield := make(chan *Frame)

	go func() {
		defer close(yield)
		for i := 0; i < 25; i++ {
			frame := NewFrame().SetWidth(w).SetHeight(h).SetFormat(fmt)

			if err := frame.ImgAlloc(); err != nil {
				return
			}

			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					frame.SetData(0, y*frame.LineSize(0)+x, x+y+i*3)
				}
			}

			// Cb and Cr
			for y := 0; y < h/2; y++ {
				for x := 0; x < w/2; x++ {
					frame.SetData(1, y*frame.LineSize(1)+x, 128+y+i*2)
					frame.SetData(2, y*frame.LineSize(2)+x, 64+x+i*5)
				}
			}

			yield <- frame
		}
	}()
	return yield
}
