// +build go1.12

// Format.
package gmf

/*

#cgo pkg-config: libavformat libavdevice libavfilter

#include <stdlib.h>
#include "libavformat/avformat.h"
#include <libavdevice/avdevice.h>
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

static char *gmf_sprintf_sdp(AVFormatContext *ctx) {
	char *sdp = malloc(sizeof(char)*16384);
	av_sdp_create(&ctx, 1, sdp, sizeof(char)*16384);
	return sdp;
}
*/
import "C"

import (
	"errors"
	"fmt"
	"io"
	"os"
	"syscall"
	"time"
	"unsafe"
)

const (
	AV_LOG_QUIET   int = C.AV_LOG_QUIET
	AV_LOG_PANIC   int = C.AV_LOG_PANIC
	AV_LOG_FATAL   int = C.AV_LOG_FATAL
	AV_LOG_ERROR   int = C.AV_LOG_ERROR
	AV_LOG_WARNING int = C.AV_LOG_WARNING
	AV_LOG_INFO    int = C.AV_LOG_INFO
	AV_LOG_VERBOSE int = C.AV_LOG_VERBOSE
	AV_LOG_DEBUG   int = C.AV_LOG_DEBUG

	FF_MOV_FLAG_FASTSTART = (1 << 7)

	AVFMT_FLAG_GENPTS int = C.AVFMT_FLAG_GENPTS
	AVFMTCTX_NOHEADER int = C.AVFMTCTX_NOHEADER
)

const (
	PLAYLIST_TYPE_NONE int = iota
	PLAYLIST_TYPE_EVENT
	PLAYLIST_TYPE_VOD
	PLAYLIST_TYPE_NB
)

type FmtCtx struct {
	Filename string

	avCtx    *C.struct_AVFormatContext
	ofmt     *OutputFmt
	streams  map[int]*Stream
	customPb bool
}

func init() {
	C.avformat_network_init()
	C.avdevice_register_all()
}

func LogSetLevel(level int) {
	C.av_log_set_level(C.int(level))
}

// @todo return error if avCtx is null
// @todo start_time is it needed?
func NewCtx(options ...[]Option) *FmtCtx {
	ctx := &FmtCtx{
		avCtx:    C.avformat_alloc_context(),
		streams:  make(map[int]*Stream),
		customPb: false,
	}

	ctx.avCtx.start_time = 0

	if len(options) == 1 {
		for _, option := range options[0] {
			option.Set(ctx.avCtx)
		}
	}

	return ctx
}

func NewOutputCtx(i interface{}, options ...[]Option) (*FmtCtx, error) {
	this := &FmtCtx{streams: make(map[int]*Stream)}

	switch t := i.(type) {
	case string:
		this.ofmt = FindOutputFmt("", i.(string), "")

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

	C.av_opt_set_defaults(unsafe.Pointer(this.avCtx))

	if len(options) == 1 {
		for _, option := range options[0] {
			option.Set(this.avCtx)
		}
	}

	this.Filename = this.ofmt.Filename

	return this, nil
}

func NewOutputCtxWithFormatName(filename, format string) (*FmtCtx, error) {
	this := &FmtCtx{streams: make(map[int]*Stream)}

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	cFormat := C.CString(format)
	defer C.free(unsafe.Pointer(cFormat))

	C.avformat_alloc_output_context2(&this.avCtx, nil, cFormat, cfilename)

	if this.avCtx == nil {
		return nil, errors.New(fmt.Sprintf("unable to allocate context"))
	}

	this.Filename = filename

	this.ofmt = &OutputFmt{Filename: filename, avOutputFmt: this.avCtx.oformat}

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

func NewInputCtxWithFormatName(filename, format string) (*FmtCtx, error) {
	ctx := NewCtx()

	if ctx.avCtx == nil {
		return nil, errors.New(fmt.Sprintf("unable to allocate context"))
	}
	if err := ctx.SetInputFormat(format); err != nil {
		return nil, err
	}
	if err := ctx.OpenInput(filename); err != nil {
		return nil, err
	}
	return ctx, nil
}

func (this *FmtCtx) SetOptions(options []*Option) {
	for _, option := range options {
		option.Set(this.avCtx)
	}
}

func (this *FmtCtx) OpenInput(filename string) error {
	var (
		cfilename *C.char
		options   *C.struct_AVDictionary = nil
	)

	if filename == "" {
		cfilename = nil
	} else {
		cfilename = C.CString(filename)
		defer C.free(unsafe.Pointer(cfilename))
	}

	if averr := C.avformat_open_input(&this.avCtx, cfilename, nil, &options); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening input '%s': %s", filename, AvError(int(averr))))
	}

	if averr := C.avformat_find_stream_info(this.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to find stream info: %s", AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) AddStreamWithCodeCtx(codeCtx *CodecCtx) (*Stream, error) {
	var ost *Stream

	// Create Video stream in output context
	if ost = this.NewStream(codeCtx.Codec()); ost == nil {
		return nil, fmt.Errorf("unable to create stream in context, filename: %s", this.Filename)
	}

	ost.DumpContexCodec(codeCtx)

	if this.avCtx.oformat != nil && int(this.avCtx.oformat.flags&C.AVFMT_GLOBALHEADER) > 0 {
		ost.SetCodecFlags()
	}

	return ost, nil
}

func (this *FmtCtx) WriteTrailer() {
	C.av_write_trailer(this.avCtx)
}

func (this *FmtCtx) IsNoFile() bool {
	return this.avCtx.oformat != nil && (this.avCtx.oformat.flags&C.AVFMT_NOFILE) != 0
}

func (this *FmtCtx) IsGlobalHeader() bool {
	return this.avCtx != nil && this.avCtx.oformat != nil && (this.avCtx.oformat.flags&C.AVFMT_GLOBALHEADER) != 0
}

func (this *FmtCtx) WriteHeader() error {
	cfilename := &(this.avCtx.filename[0])

	// If NOFILE flag isn't set and we don't use custom IO, open it
	if !this.IsNoFile() && !this.customPb {
		if averr := C.avio_open(&this.avCtx.pb, cfilename, C.AVIO_FLAG_WRITE); averr < 0 {
			return errors.New(fmt.Sprintf("Unable to open '%s': %s", this.Filename, AvError(int(averr))))
		}
	}

	if averr := C.avformat_write_header(this.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write header to '%s': %s", this.Filename, AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) WritePacket(p *Packet) error {
	if averr := C.av_interleaved_write_frame(this.avCtx, &p.avPacket); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write packet to '%s': %s", this.Filename, AvError(int(averr))))
	}

	return nil
}

func (this *FmtCtx) WritePacketNoBuffer(p *Packet) error {
	if averr := C.av_write_frame(this.avCtx, &p.avPacket); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write packet to '%s': %s", this.Filename, AvError(int(averr))))
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
	if this.ofmt == nil {
		C.av_dump_format(this.avCtx, 0, &(this.avCtx.filename[0]), 0)
	} else {
		C.av_dump_format(this.avCtx, 0, &(this.avCtx.filename[0]), 1)
	}
}

func (this *FmtCtx) DumpAv() {
	fmt.Println("AVCTX:\n", this.avCtx, "\niformat:\n", this.avCtx.iformat)
	fmt.Println("flags:", this.avCtx.flags)
}

func (this *FmtCtx) GetNextPacket() (*Packet, error) {
	pkt := NewPacket()

	for {
		ret := int(C.av_read_frame(this.avCtx, &pkt.avPacket))

		if AvErrno(ret) == syscall.EAGAIN {
			time.Sleep(10000 * time.Microsecond)
			continue
		}
		if ret == AVERROR_EOF {
			return nil, io.EOF
		}
		if ret < 0 {
			return nil, AvError(ret)
		}

		break
	}

	return pkt, nil
}

func (this *FmtCtx) NewStream(c *Codec) *Stream {
	var avCodec *C.struct_AVCodec = nil

	if c != nil {
		avCodec = c.avCodec
	}

	if st := C.avformat_new_stream(this.avCtx, avCodec); st == nil {
		return nil
	} else {
		this.streams[int(st.index)] = &Stream{avStream: st}
		Retain(this.streams[int(st.index)])
		return this.streams[int(st.index)]
	}

}

// Original structure member is called instead of len(this.streams)
// because there is no initialized Stream wrappers in input context.
func (this *FmtCtx) StreamsCnt() int {
	return int(this.avCtx.nb_streams)
}

func (this *FmtCtx) GetStream(idx int) (*Stream, error) {
	if idx > this.StreamsCnt()-1 || this.StreamsCnt() == 0 {
		return nil, errors.New(fmt.Sprintf("Stream index '%d' is out of range. There is only '%d' streams.", idx, this.StreamsCnt()))
	}

	if _, ok := this.streams[idx]; !ok {
		// create instance of Stream wrapper, when stream was initialized
		// by demuxer. it means that this is an input context.
		this.streams[idx] = &Stream{
			avStream: C.gmf_get_stream(this.avCtx, C.int(idx)),
		}
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

func (this *FmtCtx) Close() {
	if this.ofmt == nil {
		this.CloseInput()
	} else {
		this.CloseOutput()
	}
}

func (this *FmtCtx) CloseInput() {
	if this.avCtx != nil {
		C.avformat_close_input(&this.avCtx)
	}
}

func (this *FmtCtx) CloseOutput() {
	if this.avCtx == nil {
		return
	}
	if this.IsNoFile() {
		return
	}

	if this.avCtx.pb != nil && !this.customPb {
		C.avio_close(this.avCtx.pb)
	}
}

func (this *FmtCtx) Free() {
	this.Close()

	if this.avCtx != nil {
		C.avformat_free_context(this.avCtx)
	}
}
func (this *FmtCtx) Duration() float64 {
	return float64(this.avCtx.duration) / float64(AV_TIME_BASE)
}

// Total stream bitrate in bit/s
func (this *FmtCtx) BitRate() int64 {
	return int64(this.avCtx.bit_rate)
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

func (this *FmtCtx) SeekFile(ist *Stream, minTs, maxTs int64, flag int) error {
	if ret := int(C.avformat_seek_file(this.avCtx, C.int(ist.Index()), C.int64_t(0), C.int64_t(minTs), C.int64_t(maxTs), C.int(flag))); ret < 0 {
		return errors.New(fmt.Sprintf("Error creating output context: %s", AvError(ret)))
	}

	return nil
}

func (this *FmtCtx) SeekFrameAt(sec int64, streamIndex int) error {
	ist, err := this.GetStream(streamIndex)
	if err != nil {
		return err
	}

	frameTs := Rescale(sec*1000, int64(ist.TimeBase().AVR().Den), int64(ist.TimeBase().AVR().Num)) / 1000

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

func (this *FmtCtx) GetSDPString() (sdp string) {
	sdpChar := C.gmf_sprintf_sdp(this.avCtx)
	defer C.free(unsafe.Pointer(sdpChar))

	return C.GoString(sdpChar)
}

func (this *FmtCtx) WriteSDPFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Error open file:%s,error message:%s", filename, err))
	}
	defer file.Close()

	file.WriteString(this.GetSDPString())
	return nil
}

func (this *FmtCtx) Position() int {
	return int(this.avCtx.pb.pos)
}

func (this *FmtCtx) SetProbeSize(v int64) {
	this.avCtx.probesize = C.int64_t(v)
}

func (this *FmtCtx) GetProbeSize() int64 {
	return int64(this.avCtx.probesize)
}

type OutputFmt struct {
	Filename    string
	avOutputFmt *C.struct_AVOutputFormat
	CgoMemoryManage
}

func FindOutputFmt(format string, filename string, mime string) *OutputFmt {
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

func (this *OutputFmt) Free() {
	fmt.Printf("(this *OutputFmt)Free()\n")
}

func (this *OutputFmt) Name() string {
	return C.GoString(this.avOutputFmt.name)
}

func (this *OutputFmt) LongName() string {
	return C.GoString(this.avOutputFmt.long_name)
}

func (this *OutputFmt) MimeType() string {
	return C.GoString(this.avOutputFmt.mime_type)
}

func (this *OutputFmt) Infomation() string {
	return this.Filename + ":" + this.Name() + "#" + this.LongName() + "#" + this.MimeType()
}
