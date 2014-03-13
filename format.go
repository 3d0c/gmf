package gmf

/*

#cgo pkg-config: libavformat

#include "libavformat/avformat.h"

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type FmtCtx struct {
	avCtx *_Ctype_AVFormatContext
	ofmt  *OutputFmt
}

func init() {
	C.av_register_all()
}

func NewCtx() *FmtCtx {
	return &FmtCtx{
		avCtx: C.avformat_alloc_context(),
	}
}

// @todo avformat_close_input()
func (this *FmtCtx) OpenInput(filename string) error {
	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	if averr := C.avformat_open_input(&this.avCtx, cfilename, nil, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening input '%s': %s", filename, AvError(int(averr))))
	}

	if averr := C.avformat_find_stream_info(this.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to find stream info: %s", AvError(int(averr))))
	}

	// C.av_opt_set_int(this.avCtx.codec, "refcounted_frames", 1, 0)

	return nil
}

// Unlike OpenInput() it takes prepared outputfmt as an argument, e.g.
// ctx.OpenOutup(NewOutputFmt("mpeg", "/tmp/xxx.mpg", ""))
func (this *FmtCtx) OpenOutput(ofmt *OutputFmt) error {
	if ofmt == nil {
		return errors.New("Error opening output. OutputFmt is empty.")
	}

	this.ofmt = ofmt

	cfilename := C.CString(ofmt.Filename)
	defer C.free(unsafe.Pointer(cfilename))

	if averr := C.avformat_alloc_output_context2(&this.avCtx, ofmt.avOutputFmt, nil, cfilename); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening output '%s': %s", ofmt.Filename, AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) CloseOutput() {
	C.av_write_trailer(this.avCtx)
	C.avio_close(this.avCtx.pb)
}

func (this *FmtCtx) IsFileOpened() bool {
	if int(this.avCtx.flags&C.AVFMT_NOFILE) == 0 {
		return false
	}

	return true
}

func (this *FmtCtx) WriteHeader() error {
	// check is file opened

	cfilename := C.CString(this.ofmt.Filename)
	defer C.free(unsafe.Pointer(cfilename))

	if averr := C.avio_open(&this.avCtx.pb, cfilename, C.AVIO_FLAG_WRITE); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to open '%s': %s", this.ofmt.Filename, AvError(int(averr))))
	}
	fmt.Println("avCtx.pb:", this.avCtx.pb)

	if averr := C.avformat_write_header(this.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write header to '%s': %s", this.ofmt.Filename, AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) WritePacket(p *Packet) error {
	if averr := C.av_interleaved_write_frame(this.avCtx, &p.avPacket); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write packet to '%s': %s", this.ofmt.Filename, AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) SetOformat(ofmt *OutputFmt) error {
	if ofmt == nil {
		return errors.New("'ofmt' is not initialized.")
	}

	if averr := C.avformat_alloc_output_context2(&this.avCtx, ofmt.avOutputFmt, nil, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Error creating output context: %s", AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) Dump() {
	cfilename := C.CString(this.ofmt.Filename)
	defer C.free(unsafe.Pointer(cfilename))

	C.av_dump_format(this.avCtx, 0, cfilename, 1)
}

func (this *FmtCtx) StreamsCnt() int {
	return int(this.avCtx.nb_streams)
}

func (this *FmtCtx) GetStream(idx int) (*Stream, error) {
	if idx > this.StreamsCnt() || this.StreamsCnt() == 0 {
		return nil, errors.New(fmt.Sprintf("Stream index '%d' is out of range. There is only '%d' streams.", idx, this.StreamsCnt()))
	}

	return &Stream{
		avStream: C.gmf_get_stream(this.avCtx, C.int(idx)),
	}, nil
}

func (this *FmtCtx) GetVideoStream() (*Stream, error) {
	idx := C.av_find_best_stream(this.avCtx, C.AVMEDIA_TYPE_VIDEO, -1, -1, nil, 0)
	if int(idx) < 0 {
		return nil, errors.New(fmt.Sprintf("Can't find video stream"))
	}

	return this.GetStream(int(idx))
}

func (this *FmtCtx) Packets() chan *Packet {
	yield := make(chan *Packet)

	go func() {
		for {
			p := NewPacket()

			if ret := C.av_read_frame(this.avCtx, &p.avPacket); int(ret) < 0 {
				break
			}

			yield <- p
		}

		close(yield)
	}()

	return yield
}

func (this *FmtCtx) NewStream(c *Codec, _ error) *Stream {
	var avCodec *_Ctype_AVCodec = nil

	if c != nil {
		avCodec = c.avCodec
	}

	if st := C.avformat_new_stream(this.avCtx, avCodec); st == nil {
		return nil
	} else {
		return &Stream{avStream: st}
	}

}

func (this *FmtCtx) Free() {
	C.avformat_free_context(this.avCtx)
}

// OutputFmt
type OutputFmt struct {
	Filename    string
	avOutputFmt *_Ctype_AVOutputFormat
}

func NewOutputFmt(format string, filename string, mime string) *OutputFmt {
	cformat := C.CString(format)
	defer C.free(unsafe.Pointer(cformat))

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cmime := C.CString(mime)
	defer C.free(unsafe.Pointer(cmime))

	var ofmt *_Ctype_AVOutputFormat

	if ofmt = C.av_guess_format(cformat, nil, cmime); ofmt == nil {
		ofmt = C.av_guess_format(nil, cfilename, cmime)
	}

	if ofmt == nil {
		return nil
	}

	return &OutputFmt{Filename: filename, avOutputFmt: ofmt}
}

func (this *OutputFmt) Name() string {
	return C.GoString(this.avOutputFmt.name)
}
