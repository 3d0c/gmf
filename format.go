// Format.
package gmf

/*

#cgo pkg-config: libavformat

#include <stdlib.h>
#include "libavformat/avformat.h"
#include "libavutil/opt.h"

static AVStream* gmf_get_stream(AVFormatContext *ctx, int idx) {
	return ctx->streams[idx];
}

static int gmf_alloc_priv_data(AVFormatContext *s, AVDictionary **options) {
	AVDictionary *tmp = NULL;

    if (options)
        av_dict_copy(&tmp, *options, 0);

	if (s->iformat->priv_data_size > 0) {
		if (!(s->priv_data = av_mallocz(s->iformat->priv_data_size))) {
			return -1;
		 }

		 if (s->iformat->priv_class) {
			*(const AVClass**)s->priv_data = s->iformat->priv_class;
			av_opt_set_defaults(s->priv_data);
			if (av_opt_set_dict(s->priv_data, &tmp) < 0)
				return -1;
		}

		return (s->iformat->priv_data_size);
	}

	return 0;
}

*/
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

var (
	AVFMT_FLAG_GENPTS int = C.AVFMT_FLAG_GENPTS
	AVFMTCTX_NOHEADER int = C.AVFMTCTX_NOHEADER
)

type FmtCtx struct {
	avCtx    *C.struct_AVFormatContext
	ofmt     *OutputFmt
	streams  map[int]*Stream
	customPb bool
	CgoMemoryManage
}

func init() {
	C.av_register_all()
}

// @todo return error if avCtx is null
// @todo start_time is it needed?
func NewCtx() *FmtCtx {
	ctx := &FmtCtx{
		avCtx:    C.avformat_alloc_context(),
		streams:  make(map[int]*Stream),
		customPb: false,
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
	var cfilename *_Ctype_char

	if filename == "" {
		cfilename = nil
	} else {
		cfilename = C.CString(filename)
		defer C.free(unsafe.Pointer(cfilename))
	}

	if averr := C.avformat_open_input(&this.avCtx, cfilename, nil, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening input '%s': %s", filename, AvError(int(averr))))
	}

	if averr := C.avformat_find_stream_info(this.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to find stream info: %s", AvError(int(averr))))
	}
	// fmt.Println(this.avCtx.pb)
	// C.av_opt_set_int(this.avCtx.codec, "refcounted_frames", 1, 0)

	return nil
}

func (this *FmtCtx) CloseOutputAndRelease() {
	if this.avCtx == nil || this.IsNoFile() {
		return
	}

	if this.avCtx.pb != nil && !this.customPb {
		this.WriteTrailer()
		C.avio_close(this.avCtx.pb)
	}

	Release(this)
}

func (this *FmtCtx) WriteTrailer() {
	C.av_write_trailer(this.avCtx)
}

func (this *FmtCtx) CloseInputAndRelease() {
	C.avformat_close_input(&this.avCtx)
	Release(this)
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

	// If NOFILE flag isn't set and we don't use custom IO, open it
	if !this.IsNoFile() && !this.customPb {
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

func (this *FmtCtx) DumpAv() {
	fmt.Println("AVCTX:\n", this.avCtx, "\niformat:\n", this.avCtx.iformat)
	fmt.Println("raw_packet_buffer:", this.avCtx.raw_packet_buffer)
	fmt.Println("flags:", this.avCtx.flags)
	fmt.Println("packet_buffer:", this.avCtx.packet_buffer)
}

func (this *FmtCtx) GetNewPackets() chan *Packet {
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
	var avCodec *C.struct_AVCodec = nil

	if c != nil {
		avCodec = c.avCodec
	}

	if st := C.avformat_new_stream(this.avCtx, avCodec); st == nil {
		return nil
	} else {
		this.streams[int(st.index)] = &Stream{avStream: st, Pts: 0}
//		Retain(this.streams[int(st.index)])
		return this.streams[int(st.index)]
	}

}

// Original structure member is called instead of len(this.streams)
// because there is no initialized Stream wrappers in input context.
func (this *FmtCtx) StreamsCnt() int {
	return int(this.avCtx.nb_streams)
}

func (this *FmtCtx) GetStream(idx int) (*Stream, error) {
	if idx > this.StreamsCnt() || this.StreamsCnt() == 0 {
		return nil, errors.New(fmt.Sprintf("Stream index '%d' is out of range. There is only '%d' streams.", idx, this.StreamsCnt()))
	}

	if _, ok := this.streams[idx]; !ok {
		// create instance of Stream wrapper, when stream was initialized
		// by demuxer. it means that this is an input context.
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

func (this *FmtCtx) FindStreamInfo() error {
	if averr := C.avformat_find_stream_info(this.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("unable to find stream info: %s", AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) SetInputFormat(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if this.avCtx.iformat = (*C.struct_AVInputFormat)(C.av_find_input_format(cname)); this.avCtx.iformat == nil {
		return errors.New("unable to find format for name: " + name)
	}

	if int(C.gmf_alloc_priv_data(this.avCtx, nil)) < 0 {
		return errors.New("unable to allocate priv_data")
	}

	return nil
}

func (this *FmtCtx) Free() {
//	Release(this.ofmt)
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

func (this *FmtCtx) SetDebug(val int) *FmtCtx {
	this.avCtx.debug = C.int(val)
	return this
}

func (this *FmtCtx) SetFlag(flag int) *FmtCtx {
	this.avCtx.flags |= C.int(flag)
	return this
}

func (this *FmtCtx) SeekFile(ist *Stream, minTs, maxTs int, flag int) error {
	if ret := int(C.avformat_seek_file(this.avCtx, C.int(ist.Index()), C.int64_t(0), C.int64_t(minTs), C.int64_t(maxTs), C.int(flag))); ret < 0 {
		return errors.New(fmt.Sprintf("Error creating output context: %s", AvError(ret)))
	}

	return nil
}

func (this *FmtCtx) SeekFrameAt(sec int, streamIndex int) error {
	ist, err := this.GetStream(streamIndex)
	if err != nil {
		return err
	}

	frameTs := Rescale(sec*1000, ist.TimeBase().AVR().Den, ist.TimeBase().AVR().Num) / 1000

	if err := this.SeekFile(ist, frameTs, frameTs, C.AVSEEK_FLAG_FRAME); err != nil {
		return err
	}

	ist.CodecCtx().FlushBuffers()

	return nil
}

func (this *FmtCtx) SetPb(val *AVIOContext) *FmtCtx {
	this.avCtx.pb = val.avAVIOContext
	this.customPb = true
	return this
}

type OutputFmt struct {
	Filename    string
	avOutputFmt *C.struct_AVOutputFormat
}

func NewOutputFmt(format string, filename string, mime string) *OutputFmt {
	cformat := C.CString(format)
	defer C.free(unsafe.Pointer(cformat))

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cmime := C.CString(mime)
	defer C.free(unsafe.Pointer(cmime))

	var ofmt *C.struct_AVOutputFormat

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

