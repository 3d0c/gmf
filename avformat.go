package gmf

/*
#cgo pkg-config: libavformat

#include "libavformat/avformat.h"

AVStream* gmf_get_stream(AVFormatContext *ctx, int idx) {
    return ctx->streams[idx];
}

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

type FmtCtx struct {
	avCtx *_Ctype_AVFormatContext
}

func init() {
	C.av_register_all()
}

// Format Context
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

	return nil
}

// Unlike OpenInput() it takes prepared outputfmt as an argument, e.g.
// ctx.OpenOutup(NewOutputFmt("mpeg", "/tmp/xxx.mpg", ""))
func (this *FmtCtx) OpenOutput(ofmt *OutputFmt) error {
	if ofmt == nil {
		return errors.New("Error opening output. OutputFmt is empty.")
	}

	cfilename := C.CString(ofmt.Filename)
	defer C.free(unsafe.Pointer(cfilename))

	if averr := C.avformat_alloc_output_context2(&this.avCtx, ofmt.avOutputFmt, nil, cfilename); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening output '%s': %s", ofmt.Filename, AvError(int(averr))))
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

// Stream
type Stream struct {
	avStream *_Ctype_AVStream
	cc       *CodecCtx
}

func (this *Stream) Index() int {
	return int(this.avStream.index)
}

func (this *Stream) Id() int {
	return int(this.avStream.id)
}

func (this *Stream) NbFrames() int {
	return int(this.avStream.nb_frames)
}

func (this *Stream) CodecCtx() *CodecCtx {
	if this.cc != nil {
		return this.cc
	}

	c, err := NewDecoder(int(this.avStream.codec.codec_id))
	if err != nil {
		panic(fmt.Sprintf("Can't init codec for stream '%d', error:", this.Index(), err))
	}

	this.cc = &CodecCtx{
		codec:      c,
		avCodecCtx: this.avStream.codec,
	}

	if err := this.cc.Open(nil); err != nil {
		panic(fmt.Sprintf("Can't open code for stream '%d', error: %v", this.Index(), err))
	}

	return this.cc
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
