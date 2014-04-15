package gmf

/*

#cgo pkg-config: libavformat

#include "libavformat/avformat.h"

static AVStream* gmf_get_stream(AVFormatContext *ctx, int idx) {
    return ctx->streams[idx];
}

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var ()

type FmtCtx struct {
	avCtx   *_Ctype_AVFormatContext
	ofmt    *OutputFmt
	streams map[int]*Stream
}

func init() {
	C.av_register_all()
}

// @todo return error if avCtx is null
// @todo start_time is it needed?
func NewCtx() *FmtCtx {
	ctx := &FmtCtx{
		avCtx:   C.avformat_alloc_context(),
		streams: make(map[int]*Stream),
	}

	ctx.avCtx.start_time = 0

	return ctx
}

func NewOutputCtx(i interface{}) (*FmtCtx, error) {
	this := &FmtCtx{streams: make(map[int]*Stream)}

	switch t := i.(type) {
	case string:
		this.ofmt = NewOutputFmt("", i.(string), "")

	case *OutputFmt:
		this.ofmt = i.(*OutputFmt)

	default:
		return nil, errors.New(fmt.Sprintf("unexpected type %v", t))
	}

	if this.ofmt == nil {
		return nil, errors.New(fmt.Sprintf("output format is not initialized. Unable to allocate context"))
	}

	cfilename := C.CString(this.ofmt.Filename)
	defer C.free(unsafe.Pointer(cfilename))

	C.avformat_alloc_output_context2(&this.avCtx, this.ofmt.avOutputFmt, nil, cfilename)
	if this.avCtx == nil {
		return nil, errors.New(fmt.Sprintf("unable to allocate context"))
	}

	return this, nil
}

// Just a helper for NewCtx().OpenInput()
func NewInputCtx(filename string) (*FmtCtx, error) {
	ctx := NewCtx()

	if ctx.avCtx == nil {
		return nil, errors.New(fmt.Sprintf("unable to allocate context"))
	}

	if err := ctx.OpenInput(filename); err != nil {
		return nil, err
	}

	return ctx, nil
}

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
	if this.avCtx == nil {
		return
	}

	if this.avCtx.pb != nil {
		C.av_write_trailer(this.avCtx)
		C.avio_close(this.avCtx.pb)
	}

	this.Free()
}

func (this *FmtCtx) CloseInput() {
	C.avformat_close_input(&this.avCtx)
	this.Free()
}

func (this *FmtCtx) IsNoFile() bool {
	return this.avCtx.oformat != nil && (this.avCtx.oformat.flags&C.AVFMT_NOFILE) != 0
}

func (this *FmtCtx) IsGlobalHeader() bool {
	return this.avCtx != nil && this.avCtx.oformat != nil && (this.avCtx.oformat.flags&C.AVFMT_GLOBALHEADER) != 0
}

func (this *FmtCtx) WriteHeader() error {
	cfilename := C.CString(this.ofmt.Filename)
	defer C.free(unsafe.Pointer(cfilename))

	if !this.IsNoFile() {
		if averr := C.avio_open(&this.avCtx.pb, cfilename, C.AVIO_FLAG_WRITE); averr < 0 {
			return errors.New(fmt.Sprintf("Unable to open '%s': %s", this.ofmt.Filename, AvError(int(averr))))
		}
	}

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

	if _, ok := this.streams[idx]; !ok {
		this.streams[idx] = &Stream{avStream: C.gmf_get_stream(this.avCtx, C.int(idx))}
	}

	return this.streams[idx], nil
}

func (this *FmtCtx) GetBestStream(typ int32) (*Stream, error) {
	idx := C.av_find_best_stream(this.avCtx, typ, -1, -1, nil, 0)
	if int(idx) < 0 {
		return nil, errors.New(fmt.Sprintf("stream type %d not found", typ))
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

func (this *FmtCtx) NewStream(c *Codec) *Stream {
	var avCodec *_Ctype_AVCodec = nil

	if c != nil {
		avCodec = c.avCodec
	}

	if st := C.avformat_new_stream(this.avCtx, avCodec); st == nil {
		return nil
	} else {
		return &Stream{avStream: st, Pts: 0}
	}

}

func (this *FmtCtx) Free() {
	if this.avCtx != nil {
		C.avformat_free_context(this.avCtx)
	}
}

func (this *FmtCtx) Duration() int {
	return int(this.avCtx.duration)
}

func (this *FmtCtx) StartTime() int {
	return int(this.avCtx.start_time)
}

func (this *FmtCtx) SetStartTime(val int) *FmtCtx {
	this.avCtx.start_time = C.int64_t(val)
	return this
}

func (this *FmtCtx) TsOffset(stime int) int {
	// temp solution. see ffmpeg_opt.c:899
	return (0 - stime)
}

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
