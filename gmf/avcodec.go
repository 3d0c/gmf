package gmf

///*go could not map the memory correct from ReSampleContext to AVResampleContext*/
///*for that case we build this wrapper function with gmf_resample_compensate*/
//#include "libavcodec/avcodec.h"
//ReSampleContext * gmf_audio_resample_init(int output_channels, int input_channels,
//                                         int output_rate, int input_rate,
//                                         int sample_fmt_out,
//                                         int sample_fmt_in,
//                                         int filter_length, int log2_phase_count,
//                                         int linear){
//  void * ctx=av_audio_resample_init(output_channels,input_channels,
//                                         output_rate, input_rate,
//                                         sample_fmt_out,
//                                         sample_fmt_in,
//                                         filter_length, log2_phase_count,
//                                         linear, 0.8);
//						return ctx;
//}
//void gmf_resample_compensate(ReSampleContext *s, int delta, int distance){
//  av_resample_compensate(*(struct AVResampleContext**)s, delta, distance);
//}
//void gmf_audio_resample_close(ReSampleContext *s){
//  audio_resample_close(s);
//}
import "C"
import (
	"fmt"
	"sync"
	"unsafe"
)

var CODEC_TYPE_VIDEO int32 = C.AVMEDIA_TYPE_VIDEO
var CODEC_TYPE_AUDIO int32 = C.AVMEDIA_TYPE_AUDIO
var AV_NOPTS_VALUE int64 = -9223372036854775808 //C.AV_NOPTS_VALUE
var CODEC_TYPE_ENCODER int = 1
var CODEC_TYPE_DECODER int = 2
var AVCODEC_MAX_AUDIO_FRAME_SIZE int = C.AVCODEC_MAX_AUDIO_FRAME_SIZE
var TIME_BASE_Q = Rational{1, 1000000}

func init() {
	fmt.Println("Register all Codecs")
	C.avcodec_register_all()
	//C.av_log_set_level(48);
}

type avPacket C.AVPacket

var av_resample_mutex sync.Mutex

func av_audio_resample_init(trgch, srcch, trgrate, srcrate, trgfmt, srcfmt int) *ResampleContext {
	data := C.gmf_audio_resample_init(
		C.int(trgch),
		C.int(srcch),
		C.int(trgrate),
		C.int(srcrate),
		C.int(1),
		C.int(1),
		16, 10, 0)
	ctx := ResampleContext{ctx: data}
	return &ctx
}

func audio_resample(ctx *ResampleContext, outbuffer, inbuffer []byte, size int) int {
	av_resample_mutex.Lock()
	if ctx.ctx == nil {
		fmt.Printf("no ReSampleContext here!")
		return 0
	}
	result := int(C.audio_resample(ctx.ctx,
		(*C.short)(unsafe.Pointer(&outbuffer[0])),
		(*C.short)(unsafe.Pointer(&inbuffer[0])),
		C.int(size)))
	av_resample_mutex.Unlock()
	return result
}

func audio_resample_close(ctx *ResampleContext) {
	C.gmf_audio_resample_close(ctx.ctx)
}

func av_resample_compensate(ctx *ResampleContext, delta, distance int) {
	/*go could not map the memory correct from ReSampleContext to AVResampleContext*/
	/*for that case we build a wrapper function on top of this file with gmf_resample_compensate*/
	C.gmf_resample_compensate(ctx.ctx, C.int(delta), C.int(distance))
}

type Packet struct {
	avpacket  *C.AVPacket
	time_base Rational
	Dts       Timestamp
	Pts       Timestamp
	Duration  Timestamp
	Data      []byte
	Size      int
	Stream    int
	Flags     int
	Pos       int64
}

type Frame struct {
	avframe     *C.AVFrame
	buffer      []byte
	isFinished  bool
	width       int
	height      int
	size        int
	duration    int
	Pts         Timestamp
	Duration    Timestamp
	frame_count int
}

type _Codec struct {
	codec *C.AVCodec
}

type _CodecContext struct {
	ctx *C.AVCodecContext
}

type ResampleContext struct {
	ctx *C.ReSampleContext
}

func NewPacket() *Packet {
	result := new(Packet)
	result.avpacket = new(C.AVPacket)
	av_init_packet(result)
	return result
}

func (p *Packet) String() string {
	return fmt.Sprintf("S:%d;Pts:%s;Dts:%s;Idx:%d;Dur:%s|avp|S:%d;Pts:%d;Dts:%d;Idx:%d;Dur:%d", p.Size, p.Pts, p.Dts, p.Stream, p.Duration, p.avpacket.size, int64(p.avpacket.pts), int64(p.avpacket.dts), p.avpacket.stream_index, int(p.avpacket.duration))
}

func (p *Packet) Free() {
	av_free_packet(p)
}

func (p *Frame) String() string {
	return fmt.Sprintf("Size:%d:pts:%s", p.size, p.Pts)
}

func (p *Frame) destroy() {
	if p.avframe != nil {
		C.av_free(unsafe.Pointer(p.avframe))
		p.avframe = nil
	}
}

func (p *Frame) IsFinished() bool {
	return p.isFinished
}

func free_frame(frame *Frame) {
	frame.destroy()
}

func NewFrame(fmt, width, height int) *Frame {
	var frame *Frame = new(Frame)
	frame.avframe = new(C.AVFrame)
	frame.isFinished = false
	numBytes := avpicture_get_size(int32(fmt), width, height)
	if numBytes > 0 {
		b := make([]byte, numBytes)
		frame.buffer = b
		avpicture_fill(frame, frame.buffer, fmt, width, height)
		//runtime.SetFinalizer(frame, free_frame)
	}
	frame.size = numBytes
	frame.width = width
	frame.height = height
	frame.frame_count = 1
	return frame
}

var avcodec_mutex sync.Mutex

func avcodec_open(cctx _CodecContext, codec _Codec) int {
	avcodec_mutex.Lock()
	res := C.avcodec_open2(cctx.ctx, codec.codec, nil)
	avcodec_mutex.Unlock()
	return int(res)

}

func avcodec_close(cctx _CodecContext) {
	if cctx.ctx != nil {
		avcodec_mutex.Lock()
		C.avcodec_close(cctx.ctx)
		avcodec_mutex.Unlock()
	}
}

func av_free_packet(p *Packet) {
	if p.avpacket != nil {
		C.av_free_packet(p.avpacket)
		p.avpacket = nil
	}
}
func av_free_packet2(p *avPacket) {
	if p != nil {
		C.av_free_packet((*C.AVPacket)(unsafe.Pointer(p)))
		p = nil
	}
}

func av_free_codec_context(p *coder) {
	if p.Ctx.ctx != nil {
		C.av_free(unsafe.Pointer(p.Ctx.ctx))
	}
}

func avcodec_alloc_context() *_CodecContext {
	return &_CodecContext{ctx: C.avcodec_alloc_context3(nil)}
}

func avcodec_get_context_defaults2(ctx *_CodecContext, codec _Codec) {
	C.avcodec_get_context_defaults2(ctx.ctx, codec.codec._type)
}

func avpicture_fill(frame *Frame, buffer []byte, format, width, height int) int {
	return int(C.avpicture_fill((*C.AVPicture)(unsafe.Pointer(frame.avframe)), (*C.uint8_t)(unsafe.Pointer(&buffer[0])), int32(format), C.int(width), C.int(height)))
}

func alloc_avframe(frame *Frame) {
	if frame.avframe == nil {
		frame.avframe = new(C.AVFrame)
	}
}

func av_init_packet(packet *Packet) {
	if packet.avpacket == nil {
		packet.avpacket = new(C.AVPacket)
	}
	C.av_init_packet(packet.avpacket)
}

func av_init_packet2(packet *avPacket) {
	//if(packet==nil){
	//packet=new(AVPacket)
	//}
	C.av_init_packet((*C.AVPacket)(unsafe.Pointer(packet)))
}

func av_dup_packet(packet *avPacket) {
	C.av_dup_packet((*C.AVPacket)(unsafe.Pointer(packet)))
}

func avcodec_find_decoder(codec_id int32) _Codec {
	return _Codec{codec: C.avcodec_find_decoder(uint32(codec_id))}
}
func avcodec_find_decoder_by_name(name string) _Codec {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return _Codec{codec: C.avcodec_find_decoder_by_name(cname)}
}
func avcodec_find_encoder(codec_id int32) _Codec {
	return _Codec{codec: C.avcodec_find_encoder(uint32(codec_id))}
}

func avcodec_find_encoder_by_name(name string) _Codec {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	return _Codec{codec: C.avcodec_find_encoder_by_name(cname)}
}

func avpicture_get_size(fmt int32, width, height int) int {
	return int(C.avpicture_get_size(fmt, C.int(width), C.int(height)))
}

func avcodec_get_frame_defaults(frame *Frame) {
	alloc_avframe(frame)
	C.avcodec_get_frame_defaults((*C.AVFrame)(unsafe.Pointer(frame.avframe)))
}

func avpicture_alloc(frame *Frame, fmt, width, height int) int {
	return int(C.avpicture_alloc((*C.AVPicture)(unsafe.Pointer(frame.avframe)), int32(fmt), C.int(width), C.int(height)))
}

func avcodec_decode_video(ctx *_CodecContext, frame *Frame, finished *int, packet *avPacket) int {
	return int(C.avcodec_decode_video2(
		ctx.ctx,
		(*C.AVFrame)(unsafe.Pointer(frame.avframe)),
		(*C.int)(unsafe.Pointer(finished)),
		(*C.AVPacket)(unsafe.Pointer(packet))))
}

func avcodec_decode_audio(ctx *_CodecContext, buffer []byte, size *int, packet *avPacket) int {
	return int(C.avcodec_decode_audio3(
		ctx.ctx,
		(*C.int16_t)(unsafe.Pointer(&buffer[0])),
		(*C.int)(unsafe.Pointer(size)),
		(*C.AVPacket)(unsafe.Pointer(packet))))
}

func avcodec_encode_video(ctx *_CodecContext, buffer []byte, size *int, frame *Frame) int {
	return int(C.avcodec_encode_video(
		ctx.ctx,
		(*C.uint8_t)(unsafe.Pointer(&buffer[0])),
		C.int(*size),
		(*C.AVFrame)(frame.avframe)))
}

func avcodec_encode_audio(ctx *_CodecContext, outbuffer []byte, size *int, inbuffer []byte) int {
	out_size := C.avcodec_encode_audio(
		ctx.ctx,
		(*C.uint8_t)(unsafe.Pointer(&outbuffer[0])),
		C.int(*size),
		(*C.short)(unsafe.Pointer(&inbuffer[0])))

	return int(out_size)
}

func avpicture_deinterlace(outframe, inframe *Frame, fmt, width, height int) int {
	return int(C.avpicture_deinterlace(
		(*C.AVPicture)(unsafe.Pointer(outframe.avframe)),
		(*C.AVPicture)(unsafe.Pointer(inframe.avframe)),
		int32(fmt),
		C.int(width),
		C.int(height)))
}
