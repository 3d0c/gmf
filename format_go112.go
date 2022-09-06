//go:build go1.12
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

// @todo start_time is it needed?
func NewCtx(options ...[]Option) *FmtCtx {
	ctx := &FmtCtx{
		avCtx:    C.avformat_alloc_context(),
		streams:  make(map[int]*Stream),
		customPb: false,
	}

	if ctx.avCtx == nil {
		return nil
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

func (ctx *FmtCtx) SetOptions(options []*Option) {
	for _, option := range options {
		option.Set(ctx.avCtx)
	}
}

func (ctx *FmtCtx) OpenInputWithOption(filename string, inputOptions *Option) error {
	var (
		cfilename *C.char
		options   *C.struct_AVDictionary = inputOptions.Val.(*Dict).avDict
	)

	if filename == "" {
		cfilename = nil
	} else {
		cfilename = C.CString(filename)
		defer C.free(unsafe.Pointer(cfilename))
	}

	if averr := C.avformat_open_input(&ctx.avCtx, cfilename, nil, &options); averr < 0 {
		return errors.New(fmt.Sprintf("Error opening input '%s': %s", filename, AvError(int(averr))))
	}

	if averr := C.avformat_find_stream_info(ctx.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to find stream info: %s", AvError(int(averr))))
	}

	return nil
}

func (ctx *FmtCtx) OpenInput(filename string) error {
	//Create an empty Option object to pass to the open input
	inputOptionsDict := NewDict([]Pair{})
	inputOption := &Option{Key: "input_options", Val: inputOptionsDict}
	if err := ctx.OpenInputWithOption(filename, inputOption); err != nil {
		return err
	}

	return nil
}

func (ctx *FmtCtx) AddStreamWithCodeCtx(codeCtx *CodecCtx) (*Stream, error) {
	var ost *Stream

	// Create Video stream in output context
	if ost = ctx.NewStream(codeCtx.Codec()); ost == nil {
		return nil, fmt.Errorf("unable to create stream in context, filename: %s", ctx.Filename)
	}

	ost.DumpContexCodec(codeCtx)

	if ctx.avCtx.oformat != nil && int(ctx.avCtx.oformat.flags&C.AVFMT_GLOBALHEADER) > 0 {
		ost.SetCodecFlags()
	}

	return ost, nil
}

func (ctx *FmtCtx) WriteTrailer() {
	C.av_write_trailer(ctx.avCtx)
}

func (ctx *FmtCtx) IsNoFile() bool {
	return ctx.avCtx.oformat != nil && (ctx.avCtx.oformat.flags&C.AVFMT_NOFILE) != 0
}

func (ctx *FmtCtx) IsGlobalHeader() bool {
	return ctx.avCtx != nil && ctx.avCtx.oformat != nil && (ctx.avCtx.oformat.flags&C.AVFMT_GLOBALHEADER) != 0
}

func (ctx *FmtCtx) WriteHeader() error {
	cfilename := &(ctx.avCtx.url)

	// If NOFILE flag isn't set and we don't use custom IO, open it
	if !ctx.IsNoFile() && !ctx.customPb {
		if averr := C.avio_open(&ctx.avCtx.pb, *cfilename, C.AVIO_FLAG_WRITE); averr < 0 {
			return errors.New(fmt.Sprintf("Unable to open '%s': %s", ctx.Filename, AvError(int(averr))))
		}
	}

	if averr := C.avformat_write_header(ctx.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write header to '%s': %s", ctx.Filename, AvError(int(averr))))
	}

	return nil
}

func (ctx *FmtCtx) WritePacket(p *Packet) error {
	if averr := C.av_interleaved_write_frame(ctx.avCtx, &p.avPacket); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write packet to '%s': %s", ctx.Filename, AvError(int(averr))))
	}

	return nil
}

func (ctx *FmtCtx) WritePacketNoBuffer(p *Packet) error {
	if averr := C.av_write_frame(ctx.avCtx, &p.avPacket); averr < 0 {
		return errors.New(fmt.Sprintf("Unable to write packet to '%s': %s", ctx.Filename, AvError(int(averr))))
	}

	return nil
}

func (ctx *FmtCtx) SetOformat(ofmt *OutputFmt) error {
	if ofmt == nil {
		return errors.New("'ofmt' is not initialized.")
	}

	if averr := C.avformat_alloc_output_context2(&ctx.avCtx, ofmt.avOutputFmt, nil, nil); averr < 0 {
		return errors.New(fmt.Sprintf("Error creating output context: %s", AvError(int(averr))))
	}

	return nil
}

func (ctx *FmtCtx) Dump() {
	if ctx.ofmt == nil {
		C.av_dump_format(ctx.avCtx, 0, ctx.avCtx.url, 0)
	} else {
		C.av_dump_format(ctx.avCtx, 0, ctx.avCtx.url, 1)
	}
}

func (ctx *FmtCtx) DumpAv() {
	fmt.Println("AVCTX:\n", ctx.avCtx, "\niformat:\n", ctx.avCtx.iformat)
	fmt.Println("flags:", ctx.avCtx.flags)
}

func (ctx *FmtCtx) GetNextPacket() (*Packet, error) {
	pkt := NewPacket()

	for {
		ret := int(C.av_read_frame(ctx.avCtx, &pkt.avPacket))

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

func (ctx *FmtCtx) GetNextPacketForStreamIndex(streamIndex int) (*Packet, error) {
	for {
		packet, err := ctx.GetNextPacket()
		if err != nil {
			return nil, fmt.Errorf("failed to get next packet: %s", err)
		}

		if packet.StreamIndex() == streamIndex {
			return packet, nil
		}

		packet.Free()
	}
}

func (ctx *FmtCtx) GetFirstPacketForStreamIndex(streamIndex int) (*Packet, error) {
	currentPacket, err := ctx.GetNextPacketForStreamIndex(streamIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get current packet: %s", err)
	}
	originalPosition := currentPacket.Pos()

	err = ctx.SeekFrameAt(0, streamIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to first frame: %s", err)
	}

	packet, err := ctx.GetNextPacketForStreamIndex(streamIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get next packet: %s", err)
	}

	err = ctx.SeekFrameAt(originalPosition, streamIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to original position: %s", err)
	}

	return packet, nil
}

func (ctx *FmtCtx) GetLastPacketForStreamIndex(streamIndex int) (*Packet, error) {
	err := ctx.SeekFrameAt(int64(ctx.Duration()), streamIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to seek to last frame: %s", err)
	}

	return ctx.GetNextPacketForStreamIndex(streamIndex)
}

func (ctx *FmtCtx) GetNewPackets() chan *Packet {
	yield := make(chan *Packet)

	go func() {
		for {
			p := NewPacket()

			if ret := C.av_read_frame(ctx.avCtx, &p.avPacket); int(ret) < 0 {
				break
			}

			yield <- p
		}

		close(yield)
	}()

	return yield
}

func (ctx *FmtCtx) NewStream(c *Codec) *Stream {
	var avCodec *C.struct_AVCodec = nil

	if c != nil {
		avCodec = c.avCodec
	}

	if st := C.avformat_new_stream(ctx.avCtx, avCodec); st == nil {
		return nil
	} else {
		ctx.streams[int(st.index)] = &Stream{avStream: st}
		Retain(ctx.streams[int(st.index)])
		return ctx.streams[int(st.index)]
	}

}

// Original structure member is called instead of len(this.streams)
// because there is no initialized Stream wrappers in input context.
func (ctx *FmtCtx) StreamsCnt() int {
	return int(ctx.avCtx.nb_streams)
}

func (ctx *FmtCtx) GetStream(idx int) (*Stream, error) {
	if idx > ctx.StreamsCnt()-1 || ctx.StreamsCnt() == 0 {
		return nil, errors.New(fmt.Sprintf("Stream index '%d' is out of range. There is only '%d' streams.", idx, ctx.StreamsCnt()))
	}

	if _, ok := ctx.streams[idx]; !ok {
		// create instance of Stream wrapper, when stream was initialized
		// by demuxer. it means that ctx is an input context.
		ctx.streams[idx] = &Stream{
			avStream: C.gmf_get_stream(ctx.avCtx, C.int(idx)),
		}
	}

	return ctx.streams[idx], nil
}

func (ctx *FmtCtx) GetBestStream(typ int32) (*Stream, error) {
	idx := C.av_find_best_stream(ctx.avCtx, typ, -1, -1, nil, 0)
	if int(idx) < 0 {
		return nil, errors.New(fmt.Sprintf("stream type %d not found", typ))
	}

	return ctx.GetStream(int(idx))
}

func (ctx *FmtCtx) FindStreamInfo() error {
	if averr := C.avformat_find_stream_info(ctx.avCtx, nil); averr < 0 {
		return errors.New(fmt.Sprintf("unable to find stream info: %s", AvError(int(averr))))
	}

	return nil
}

func (ctx *FmtCtx) SetInputFormat(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	if ctx.avCtx.iformat = (*C.struct_AVInputFormat)(C.av_find_input_format(cname)); ctx.avCtx.iformat == nil {
		return errors.New("unable to find format for name: " + name)
	}

	if int(C.gmf_alloc_priv_data(ctx.avCtx, nil)) < 0 {
		return errors.New("unable to allocate priv_data")
	}

	return nil
}

func (ctx *FmtCtx) Close() {
	if ctx.ofmt == nil {
		ctx.CloseInput()
	} else {
		ctx.CloseOutput()
	}
}

func (ctx *FmtCtx) CloseInput() {
	if ctx.avCtx != nil {
		C.avformat_close_input(&ctx.avCtx)
	}
}

func (ctx *FmtCtx) CloseOutput() {
	if ctx.avCtx == nil {
		return
	}
	if ctx.IsNoFile() {
		return
	}

	if ctx.avCtx.pb != nil && !ctx.customPb {
		C.avio_close(ctx.avCtx.pb)
	}
}

func (ctx *FmtCtx) Free() {
	ctx.Close()

	if ctx.avCtx != nil {
		C.avformat_free_context(ctx.avCtx)
	}
}
func (ctx *FmtCtx) Duration() float64 {
	return float64(ctx.avCtx.duration) / float64(AV_TIME_BASE)
}

// Total stream bitrate in bit/s
func (ctx *FmtCtx) BitRate() int64 {
	return int64(ctx.avCtx.bit_rate)
}

func (ctx *FmtCtx) StartTime() int {
	return int(ctx.avCtx.start_time)
}

func (ctx *FmtCtx) SetStartTime(val int) *FmtCtx {
	ctx.avCtx.start_time = C.int64_t(val)
	return ctx
}

func (ctx *FmtCtx) TsOffset(stime int) int {
	// temp solution. see ffmpeg_opt.c:899
	return (0 - stime)
}

func (ctx *FmtCtx) SetDebug(val int) *FmtCtx {
	ctx.avCtx.debug = C.int(val)
	return ctx
}

func (ctx *FmtCtx) SetFlag(flag int) *FmtCtx {
	ctx.avCtx.flags |= C.int(flag)
	return ctx
}

func (ctx *FmtCtx) SeekFile(ist *Stream, minTs, maxTs int64, flag int) error {
	if ret := int(C.avformat_seek_file(ctx.avCtx, C.int(ist.Index()), C.int64_t(0), C.int64_t(minTs), C.int64_t(maxTs), C.int(flag))); ret < 0 {
		return errors.New(fmt.Sprintf("Error creating output context: %s", AvError(ret)))
	}

	return nil
}

func (ctx *FmtCtx) SeekFrameAt(sec int64, streamIndex int) error {
	ist, err := ctx.GetStream(streamIndex)
	if err != nil {
		return err
	}

	frameTs := Rescale(sec*1000, int64(ist.TimeBase().AVR().Den), int64(ist.TimeBase().AVR().Num)) / 1000

	if err := ctx.SeekFile(ist, frameTs, frameTs, C.AVSEEK_FLAG_FRAME); err != nil {
		return err
	}

	ist.CodecCtx().FlushBuffers()

	return nil
}

func (ctx *FmtCtx) SeekFrameAtTimeCode(timecode string, streamIndex int) error {
	ist, err := ctx.GetStream(streamIndex)
	if err != nil {
		return err
	}

	istCodecCtx := ist.CodecCtx()
	istAvgFrameRate := ist.GetAvgFrameRate()

	// check if timecode is valid (i.e. if the timecode is between the first and last frame of the video)
	firstPacket, err := ctx.GetFirstPacketForStreamIndex(streamIndex)
	if err != nil {
		return err
	}
	lastPacket, err := ctx.GetLastPacketForStreamIndex(streamIndex)
	if err != nil {
		return err
	}
	firstFrameTimeCode, err := getPacketTimeCode(firstPacket, istCodecCtx, istAvgFrameRate)
	if err != nil {
		return err
	}
	lastFrameTimeCode, err := getPacketTimeCode(lastPacket, istCodecCtx, istAvgFrameRate)
	if err != nil {
		return err
	}

	tcBetween, err := IsTimeCodeBetween(timecode, firstFrameTimeCode, lastFrameTimeCode)
	if err != nil {
		return err
	}
	if !tcBetween {
		return errors.New("timecode is not between the first and last frame of the video")
	}

	// get comparable timecodes
	hhToTest, hhStart, hhEnd, err := TimeCodeToComparable(timecode, firstFrameTimeCode, lastFrameTimeCode)
	if err != nil {
		return err
	}

	// calculate a good approximation of the position of the needed frame
	var relativePos float64 = float64(float64(hhToTest-hhStart) / float64(hhEnd-hhStart))

	// rewind a bit to be sure we are not too far
	if relativePos-0.001 > 0 {
		relativePos -= 0.001
	}

	sec := ctx.Duration() * relativePos
	err = ctx.SeekFrameAt(int64(sec), streamIndex)
	if err != nil {
		return err
	}

	found := false
	for !found {
		packet, err := ctx.GetNextPacketForStreamIndex(streamIndex)
		if err != nil {
			return err
		}

		frameTimeCode, err := getPacketTimeCode(packet, istCodecCtx, istAvgFrameRate)
		if err != nil {
			return err
		}

		hhFrameTimeCode, _, _, err := TimeCodeToComparable(frameTimeCode, firstFrameTimeCode, lastFrameTimeCode)
		if err != nil {
			return err
		}

		if hhFrameTimeCode >= hhToTest {
			return nil
		}

		if hhFrameTimeCode < hhToTest {
			sec += 1
			err = ctx.SeekFrameAt(int64(sec), streamIndex)
			if err != nil {
				return err
			}
		}

		packet.Free()
	}

	return nil
}

func isPacketLaterThanTimeCode(packet *Packet, timecode string, codecCtx *CodecCtx, avgFrameRate AVRational) (bool, error) {
	frameTimeCode, err := getPacketTimeCode(packet, codecCtx, avgFrameRate)
	if err != nil {
		return false, err
	}
	if frameTimeCode >= timecode {
		return true, nil
	}
	return false, nil
}

func getPacketTimeCode(packet *Packet, codecCtx *CodecCtx, avgFrameRate AVRational) (string, error) {
	frames, err := codecCtx.Decode(packet)
	if err != nil {
		return "", err
	}
	for _, frame := range frames {
		timecode, err := frame.GetTimeCode(avgFrameRate)
		if err != nil {
			return "", err
		}
		frame.Free()
		return timecode, nil
	}
	return "", nil
}

func (ctx *FmtCtx) SetPb(val *AVIOContext) *FmtCtx {
	ctx.avCtx.pb = val.avAVIOContext
	ctx.customPb = true
	return ctx
}

func (ctx *FmtCtx) GetSDPString() (sdp string) {
	sdpChar := C.gmf_sprintf_sdp(ctx.avCtx)
	defer C.free(unsafe.Pointer(sdpChar))

	return C.GoString(sdpChar)
}

func (ctx *FmtCtx) WriteSDPFile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return errors.New(fmt.Sprintf("Error open file:%s,error message:%s", filename, err))
	}
	defer file.Close()

	file.WriteString(ctx.GetSDPString())
	return nil
}

func (ctx *FmtCtx) Position() int {
	return int(ctx.avCtx.pb.pos)
}

func (ctx *FmtCtx) SetProbeSize(v int64) {
	ctx.avCtx.probesize = C.int64_t(v)
}

func (ctx *FmtCtx) GetProbeSize() int64 {
	return int64(ctx.avCtx.probesize)
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

func (f *OutputFmt) Free() {
	fmt.Printf("(f *OutputFmt)Free()\n")
}

func (f *OutputFmt) Name() string {
	return C.GoString(f.avOutputFmt.name)
}

func (f *OutputFmt) LongName() string {
	return C.GoString(f.avOutputFmt.long_name)
}

func (f *OutputFmt) MimeType() string {
	return C.GoString(f.avOutputFmt.mime_type)
}

func (f *OutputFmt) Infomation() string {
	return f.Filename + ":" + f.Name() + "#" + f.LongName() + "#" + f.MimeType()
}
