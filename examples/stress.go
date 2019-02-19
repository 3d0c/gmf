package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/3d0c/gmf"
)

const (
	output = "/tmp/bbb.mp4"
)

var (
	input string
)

var cases map[string]func() = map[string]func(){
	"ref":           foo,
	"fmtctx":        fmtctx,
	"fmtctxavio":    fmtctxavio,
	"fmtinput":      fmtinput,
	"newfmtctx":     newfmtctx,
	"opencc":        opencc,
	"decode_step":   decodeStep,
	"decode":        decode,
	"encode_step":   encodeStep,
	"getnextpacket": getnextpacket,
	"rescale_step":  rescaleStep,
	"newframe":      newframe,
	"resample_step": resampleStep,
}

func foo() {
	var a int = 0
	b := map[int]*gmf.Stream{}
	_ = a
	_ = b
	return
}

func newfmtctx() {
	var (
		ictx *gmf.FmtCtx
	)

	if ictx = gmf.NewCtx(); ictx == nil {
		log.Fatalln("error creating format ctx")
	}

	ictx.Free()
	ictx = nil
}

func fmtctx() {
	var (
		ictx *gmf.FmtCtx
		err  error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatal(err)
	}

	ictx.Free()
}

var (
	fp  *os.File
	seq int
)

func reader() ([]byte, int) {
	var (
		err error
		n   int
	)

	if fp == nil {
		fp, err = os.Open(input)
		if err != nil {
			panic(err)
		}

		_, err := fp.Stat()
		if err != nil {
			panic(err)
		}
	}

	b := make([]byte, 500000)

	if n, err = fp.Read(b); err != nil {
		panic(err)
	}

	return b, n
}

func offset(off int64, whence int) int64 {
	r, _ := fp.Seek(off, whence)
	log.Printf("offset %d, %d >> %d\n", off, whence, r)
	return r
}

func fmtctxavio() {
	var (
		ictx    *gmf.FmtCtx
		avioCtx *gmf.AVIOContext
		err     error
	)

	if ictx = gmf.NewCtx([]gmf.Option{
		{Key: "avioflags", Val: 0x8000},
	}); ictx == nil {
		log.Fatalln("ictx is nil")
	}
	defer ictx.Free()

	if avioCtx, err = gmf.NewAVIOContext(ictx, &gmf.AVIOHandlers{ReadPacket: reader, Seek: offset}, 500000); err != nil {
		log.Fatalln(err)
	}
	defer avioCtx.Free()

	ictx.SetPb(avioCtx).OpenInput("")

	fp.Close()
	fp = nil
}

/*
==10255== LEAK SUMMARY:
==10255==    definitely lost: 0 bytes in 0 blocks
==10255==    indirectly lost: 0 bytes in 0 blocks
==10255==      possibly lost: 1,440 bytes in 5 blocks
==10255==    still reachable: 0 bytes in 0 blocks
==10255==         suppressed: 0 bytes in 0 blocks
*/
func fmtinput() {
	var (
		ictx *gmf.FmtCtx
		ist  *gmf.Stream
		err  error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	if ist, err = ictx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO); err != nil {
		log.Fatalln(err)
	}

	ist.Free()
}

/*
==10638== LEAK SUMMARY:
==10638==    definitely lost: 0 bytes in 0 blocks
==10638==    indirectly lost: 0 bytes in 0 blocks
==10638==      possibly lost: 1,440 bytes in 5 blocks
==10638==    still reachable: 0 bytes in 0 blocks
==10638==         suppressed: 0 bytes in 0 blocks
*/
func opencc() {
	var (
		octx    *gmf.FmtCtx
		codec   *gmf.Codec
		cc      *gmf.CodecCtx
		options []gmf.Option
		ost     *gmf.Stream
		err     error
	)

	octx = gmf.NewCtx()

	codec, err = gmf.FindEncoder("libx264")
	if err != nil {
		fmt.Printf("Error finding codec - %s\n", err)
		return
	}
	defer octx.Free()

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		fmt.Printf("unable to create codec context")
		return
	}
	defer cc.Free()

	options = []gmf.Option{
		{Key: "time_base", Val: gmf.AVR{Num: 1, Den: 25}},
		{Key: "pixel_format", Val: gmf.AV_PIX_FMT_YUV420P},
		{Key: "video_size", Val: "800x600"},
	}

	cc.SetProfile(gmf.FF_PROFILE_H264_MAIN)
	cc.SetOptions(options)

	if err := cc.Open(nil); err != nil {
		fmt.Printf("%s\n", err)
		return
	}

	if ost, err = octx.AddStreamWithCodeCtx(cc); err != nil {
		log.Fatalln(err)
	}

	ost.Free()
}

/*
==11968== LEAK SUMMARY:
==11968==    definitely lost: 0 bytes in 0 blocks
==11968==    indirectly lost: 0 bytes in 0 blocks
==11968==      possibly lost: 1,440 bytes in 5 blocks
==11968==    still reachable: 0 bytes in 0 blocks
==11968==         suppressed: 0 bytes in 0 blocks
*/
func getnextpacket() {
	var (
		ictx *gmf.FmtCtx
		pkt  *gmf.Packet
		err  error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	if pkt, err = ictx.GetNextPacket(); err != nil {
		if err == io.EOF {
		} else {
			log.Fatalln(pkt, err)
		}
	}

	pkt.Free()
}

/*
==13958== LEAK SUMMARY:
==13958==    definitely lost: 0 bytes in 0 blocks
==13958==    indirectly lost: 0 bytes in 0 blocks
==13958==      possibly lost: 1,440 bytes in 5 blocks
==13958==    still reachable: 0 bytes in 0 blocks
==13958==         suppressed: 0 bytes in 0 blocks
*/
func decodeStep() {
	var (
		ictx   *gmf.FmtCtx
		ist    *gmf.Stream
		pkt    *gmf.Packet
		frames []*gmf.Frame
		err    error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	if pkt, err = ictx.GetNextPacket(); err != nil {
		if err == io.EOF {
		} else {
			log.Fatalln(pkt, err)
		}
	}

	ist, err = ictx.GetStream(pkt.StreamIndex())
	if err != nil {
		log.Fatalln(err)
	}

	if frames, err = ist.CodecCtx().Decode(pkt); err != nil {
		log.Fatalln(err)
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	pkt.Free()

	if frames, err = ist.CodecCtx().Decode(nil); err != nil {
		log.Fatalln(err)
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	ist.CodecCtx().Free()
	ist.Free()

	frames = nil
}

/*
==7174== LEAK SUMMARY:
==7174==    definitely lost: 0 bytes in 0 blocks
==7174==    indirectly lost: 0 bytes in 0 blocks
==7174==      possibly lost: 1,152 bytes in 4 blocks
==7174==    still reachable: 0 bytes in 0 blocks
==7174==         suppressed: 0 bytes in 0 blocks
*/
func decode() {
	var (
		ictx   *gmf.FmtCtx
		ist    *gmf.Stream
		pkt    *gmf.Packet
		frames []*gmf.Frame
		err    error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	for {
		if pkt, err = ictx.GetNextPacket(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln(pkt, err)
			}
		}

		ist, err = ictx.GetStream(pkt.StreamIndex())
		if err != nil {
			log.Fatalln(err)
		}

		if frames, err = ist.CodecCtx().Decode(pkt); err != nil {
			log.Fatalln(err)
		}

		for i, _ := range frames {
			frames[i].Free()
		}

		pkt.Free()
	}

	if frames, err = ist.CodecCtx().Decode(nil); err != nil {
		log.Fatalln(err)
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	ist.CodecCtx().Free()
	ist.Free()

	frames = nil
}

/*
==10986== LEAK SUMMARY:
==10986==    definitely lost: 0 bytes in 0 blocks
==10986==    indirectly lost: 0 bytes in 0 blocks
==10986==      possibly lost: 1,440 bytes in 5 blocks
==10986==    still reachable: 0 bytes in 0 blocks
==10986==         suppressed: 0 bytes in 0 blocks
*/
func encodeStep() {
	var (
		ictx, octx *gmf.FmtCtx
		ist, ost   *gmf.Stream
		pkt        *gmf.Packet
		op         []*gmf.Packet
		codec      *gmf.Codec
		cc         *gmf.CodecCtx
		options    []gmf.Option
		frames     []*gmf.Frame
		pts        int64
		err        error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	if octx, err = gmf.NewOutputCtx(output); err != nil {
		log.Fatalln(err)
	}
	defer octx.Free()

	codec, err = gmf.FindEncoder("libx264")
	if err != nil {
		log.Fatalf("Error finding codec - %s\n", err)
	}

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		log.Fatalf("unable to create codec context")
	}
	defer cc.Free()

	options = []gmf.Option{
		{Key: "time_base", Val: gmf.AVR{Num: 1, Den: 25}},
		{Key: "pixel_format", Val: gmf.AV_PIX_FMT_YUV420P},
		{Key: "video_size", Val: "320x200"},
	}

	cc.SetProfile(gmf.FF_PROFILE_H264_MAIN)
	cc.SetOptions(options)

	if err := cc.Open(nil); err != nil {
		log.Fatalln(err)
	}

	if ost, err = octx.AddStreamWithCodeCtx(cc); err != nil {
		log.Fatalln(err)
	}
	defer ost.Free()

	ost.SetCodecCtx(cc)

	octx.WriteHeader()

	for {
		if pkt, err = ictx.GetNextPacket(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln(pkt, err)
			}
		}

		ist, err = ictx.GetStream(pkt.StreamIndex())
		if err != nil {
			log.Fatalln(err)
		}

		if ist.Type() != gmf.AVMEDIA_TYPE_VIDEO {
			pkt.Free()
			ist.Free()
			continue
		}

		break
	}

	if frames, err = ist.CodecCtx().Decode(pkt); err != nil {
		log.Fatalln(err)
	}
	pkt.Free()

	for i, _ := range frames {
		frames[i].SetPts(pts)
		pts++
	}

	if op, err = ost.CodecCtx().Encode(frames, -1); err != nil {
		log.Fatalln(err)
	}

	for i, _ := range op {
		octx.WritePacket(op[i])
		op[i].Free()
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	// drain decoder
	if frames, err = ist.CodecCtx().Decode(nil); err != nil {
		log.Fatalln(err)
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	// drain encoder
	if op, err = ost.CodecCtx().Encode(nil, 1); err != nil {
		log.Fatalln(err)
	}
	for i, _ := range op {
		op[i].Free()
	}

	frames = nil

	octx.WriteTrailer()

	ist.CodecCtx().Free()
	ost.CodecCtx().Free()
	ist.Free()
	ost.Free()
}

/*
==31970== LEAK SUMMARY:
==31970==    definitely lost: 0 bytes in 0 blocks
==31970==    indirectly lost: 0 bytes in 0 blocks
==31970==      possibly lost: 1,440 bytes in 5 blocks
==31970==    still reachable: 48 bytes in 2 blocks
==31970==         suppressed: 0 bytes in 0 blocks
*/
func newframe() {
	var (
		frame *gmf.Frame
		err   error
	)

	frame = gmf.NewFrame().SetWidth(640).SetHeight(320).SetFormat(gmf.AV_PIX_FMT_YUV420P)
	if err = frame.ImgAlloc(); err != nil {
		panic(err)
	}

	frame.Free()
}

/*
==1031== LEAK SUMMARY:
==1031==    definitely lost: 0 bytes in 0 blocks
==1031==    indirectly lost: 0 bytes in 0 blocks
==1031==      possibly lost: 1,440 bytes in 5 blocks
==1031==    still reachable: 0 bytes in 0 blocks
==1031==         suppressed: 0 bytes in 0 blocks
*/
func rescaleStep() {
	var (
		ictx, octx *gmf.FmtCtx
		ist, ost   *gmf.Stream
		pkt        *gmf.Packet
		op         []*gmf.Packet
		codec      *gmf.Codec
		cc         *gmf.CodecCtx
		options    []gmf.Option
		frames     []*gmf.Frame
		pts        int64
		err        error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	if octx, err = gmf.NewOutputCtx(output); err != nil {
		log.Fatalln(err)
	}
	defer octx.Free()

	codec, err = gmf.FindEncoder("libx264")
	if err != nil {
		log.Fatalf("Error finding codec - %s\n", err)
	}

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		log.Fatalf("unable to create codec context")
	}
	defer cc.Free()

	options = []gmf.Option{
		{Key: "time_base", Val: gmf.AVR{Num: 1, Den: 25}},
		{Key: "pixel_format", Val: gmf.AV_PIX_FMT_YUV420P},
		{Key: "video_size", Val: "320x200"},
	}

	cc.SetProfile(gmf.FF_PROFILE_H264_MAIN)
	cc.SetOptions(options)

	if err := cc.Open(nil); err != nil {
		log.Fatalln(err)
	}

	if ost, err = octx.AddStreamWithCodeCtx(cc); err != nil {
		log.Fatalln(err)
	}
	defer ost.Free()

	ost.SetCodecCtx(cc)

	octx.WriteHeader()

	for {
		if pkt, err = ictx.GetNextPacket(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln(pkt, err)
			}
		}

		ist, err = ictx.GetStream(pkt.StreamIndex())
		if err != nil {
			log.Fatalln(err)
		}

		if ist.Type() != gmf.AVMEDIA_TYPE_VIDEO {
			pkt.Free()
			continue
		} else {
			break
		}
	}

	if ost.SwsCtx == nil {
		icc := ist.CodecCtx()
		occ := ost.CodecCtx()
		if ost.SwsCtx, err = gmf.NewSwsCtx(icc.Width(), icc.Height(), icc.PixFmt(), occ.Width(), occ.Height(), occ.PixFmt(), gmf.SWS_BICUBIC); err != nil {
			panic(err)
		}
	}

	if frames, err = ist.CodecCtx().Decode(pkt); err != nil {
		log.Fatalln(err)
	}

	if frames, err = gmf.DefaultRescaler(ost.SwsCtx, frames); err != nil {
		panic(err)
	}

	for i, _ := range frames {
		frames[i].SetPts(pts)
		pts++
	}

	if op, err = ost.CodecCtx().Encode(frames, -1); err != nil {
		log.Fatalln(err)
	}

	for i, _ := range op {
		octx.WritePacket(op[i])
		op[i].Free()
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	pkt.Free()

	if frames, err = ist.CodecCtx().Decode(nil); err != nil {
		log.Fatalln(err)
	}

	if frames, err = gmf.DefaultRescaler(ost.SwsCtx, frames); err != nil {
		panic(err)
	}

	for i, _ := range frames {
		frames[i].Free()
	}

	if op, err = ost.CodecCtx().Encode(nil, 1); err != nil {
		log.Fatalln(err)
	}
	for i, _ := range op {
		op[i].Free()
	}

	frames = nil

	octx.WriteTrailer()

	ist.CodecCtx().Free()
	ist.Free()
}

/*
==14439== LEAK SUMMARY:
==14439==    definitely lost: 0 bytes in 0 blocks
==14439==    indirectly lost: 0 bytes in 0 blocks
==14439==      possibly lost: 1,440 bytes in 5 blocks
==14439==    still reachable: 0 bytes in 0 blocks
==14439==         suppressed: 0 bytes in 0 blocks
*/
func resampleStep() {
	var (
		ictx, octx *gmf.FmtCtx
		ist, ost   *gmf.Stream
		pkt        *gmf.Packet
		op         []*gmf.Packet
		codec      *gmf.Codec
		cc         *gmf.CodecCtx
		options    []gmf.Option
		frames     []*gmf.Frame
		// pts           int64
		err error
	)

	if ictx, err = gmf.NewInputCtx(input); err != nil {
		log.Fatalf("Error creating context - %s\n", err)
	}
	defer ictx.Free()

	if octx, err = gmf.NewOutputCtx(output); err != nil {
		log.Fatalln(err)
	}
	defer octx.Free()

	codec, err = gmf.FindEncoder("aac")
	if err != nil {
		log.Fatalf("Error finding codec - %s\n", err)
	}

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		log.Fatalf("unable to create codec context")
	}
	defer cc.Free()

	options = []gmf.Option{
		{Key: "time_base", Val: gmf.AVR{Num: 1, Den: 22050}},
		{Key: "ar", Val: 22050},
		{Key: "ac", Val: 2},
		{Key: "channel_layout", Val: 3},
	}

	cc.SetSampleFmt(8)
	cc.SetOptions(options)

	if err := cc.Open(nil); err != nil {
		log.Fatalln(err)
	}

	if ost, err = octx.AddStreamWithCodeCtx(cc); err != nil {
		log.Fatalln(err)
	}
	defer ost.Free()

	ost.SetCodecCtx(cc)

	octx.WriteHeader()

	for {
		if pkt, err = ictx.GetNextPacket(); err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln(pkt, err)
			}
		}

		ist, err = ictx.GetStream(pkt.StreamIndex())
		if err != nil {
			log.Fatalln(err)
		}

		if ist.Type() != gmf.AVMEDIA_TYPE_AUDIO {
			pkt.Free()
			continue
		} else {
			break
		}
	}

	if ost.SwrCtx == nil {
		icc := ist.CodecCtx()
		occ := ost.CodecCtx()
		options := []*gmf.Option{
			{Key: "in_channel_layout", Val: icc.ChannelLayout()},
			{Key: "out_channel_layout", Val: occ.ChannelLayout()},
			{Key: "in_sample_rate", Val: icc.SampleRate()},
			{Key: "out_sample_rate", Val: occ.SampleRate()},
			{Key: "in_sample_fmt", Val: gmf.SampleFormat(icc.SampleFmt())},
			{Key: "out_sample_fmt", Val: gmf.SampleFormat(gmf.AV_SAMPLE_FMT_FLTP)},
		}

		if ost.SwrCtx, err = gmf.NewSwrCtx(options, occ.Channels(), occ.SampleFmt()); err != nil {
			panic(err)
		}
		ost.AvFifo = gmf.NewAVAudioFifo(icc.SampleFmt(), ist.CodecCtx().Channels(), 1024)
	}

	if frames, err = ist.CodecCtx().Decode(pkt); err != nil {
		log.Fatalln(err)
	}

	frames = gmf.DefaultResampler(ost, frames, false)

	if op, err = ost.CodecCtx().Encode(frames, -1); err != nil {
		log.Fatalln(err)
	}
	for i, _ := range frames {
		frames[i].Free()
	}

	for i, _ := range op {
		octx.WritePacket(op[i])
		op[i].Free()
	}

	pkt.Free()

	if frames, err = ist.CodecCtx().Decode(nil); err != nil {
		log.Fatalln(err)
	}

	frames = gmf.DefaultResampler(ost, frames, true)

	for i, _ := range frames {
		frames[i].Free()
	}

	if op, err = ost.CodecCtx().Encode(nil, 1); err != nil {
		log.Fatalln(err)
	}
	for i, _ := range op {
		op[i].Free()
	}

	frames = nil

	octx.WriteTrailer()

	for i := 0; i < ictx.StreamsCnt(); i++ {
		st, _ := ictx.GetStream(i)
		st.CodecCtx().Free()
		st.Free()
	}
}

func main() {
	var (
		p  bool
		n  int
		fn string
	)

	flag.BoolVar(&p, "p", true, "do pause before exit")
	flag.IntVar(&n, "n", 10, "repeat n times")
	flag.StringVar(&fn, "fn", "", "function name to test")
	flag.StringVar(&input, "input", "bbb.mp4", "input file")
	flag.Parse()

	log.SetFlags(log.Lshortfile)

	if fn == "" {
		log.Fatalf("Please provide a function name")
	}

	if _, ok := cases[fn]; !ok {
		log.Fatalf("No such function found\n")
	}

	for i := 0; i < n; i++ {
		cases[fn]()
	}

	println("done, press enter to quit")

	if p {
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
}
