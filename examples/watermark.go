package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"strings"
	"syscall"

	"github.com/3d0c/gmf"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, " ")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

type Input struct {
	ctx       *gmf.FmtCtx
	lastFrame *gmf.Frame
	finished  bool
}

var (
	inputs map[int]*Input
)

func finishedNb() int {
	result := 0

	for _, input := range inputs {
		if input.finished {
			result++
		}
	}

	return result
}

func assert(i interface{}, err error) interface{} {
	if err != nil {
		log.Fatal(err)
	}

	return i
}

func addStream(codecName string, oc *gmf.FmtCtx, ist *gmf.Stream) (int, int) {
	var cc *gmf.CodecCtx
	var ost *gmf.Stream

	codec := assert(gmf.FindEncoder(codecName)).(*gmf.Codec)

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		log.Fatal(errors.New("unable to create codec context"))
	}

	if oc.IsGlobalHeader() {
		cc.SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	}

	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if cc.Type() == gmf.AVMEDIA_TYPE_AUDIO {
		cc.SetSampleFmt(ist.CodecCtx().SampleFmt())
		cc.SetSampleRate(ist.CodecCtx().SampleRate())
		cc.SetChannels(ist.CodecCtx().Channels())
		cc.SelectChannelLayout()
		cc.SelectSampleRate()

	}

	if cc.Type() == gmf.AVMEDIA_TYPE_VIDEO {
		cc.SetTimeBase(gmf.AVR{1, 25})
		cc.SetProfile(gmf.FF_PROFILE_MPEG4_SIMPLE)
		cc.SetDimension(ist.CodecCtx().Width(), ist.CodecCtx().Height())
		cc.SetPixFmt(ist.CodecCtx().PixFmt())
	}

	if err := cc.Open(nil); err != nil {
		log.Fatal(err)
	}

	par := gmf.NewCodecParameters()
	if err := par.FromContext(cc); err != nil {
		log.Fatalf("error creating codec parameters from context - %s", err)
	}
	defer par.Free()

	if ost = oc.NewStream(codec); ost == nil {
		log.Fatal(errors.New("unable to create stream in output context"))
	}

	ost.CopyCodecPar(par)
	ost.SetCodecCtx(cc)
	ost.SetTimeBase(gmf.AVR{Num: 1, Den: 25})
	ost.SetRFrameRate(gmf.AVR{Num: 25, Den: 1})

	return ist.Index(), ost.Index()
}

func main() {
	var (
		src       arrayFlags
		dst       string
		streamMap map[int]int = make(map[int]int)
	)

	log.SetFlags(log.Lshortfile)

	flag.Var(&src, "src", "source files, e.g.: -src=bbb.mp4 -src=image.png")
	flag.StringVar(&dst, "dst", "", "destination file, e.g. -dst=result.mp4")
	flag.Parse()

	if len(src) == 0 || dst == "" {
		log.Fatal("at least one source and destination required, e.g.\n./watermark -src=bbb.mp4 -src=test.png -dst=overlay.mp4")
	}

	octx, err := gmf.NewOutputCtx(dst)
	if err != nil {
		log.Fatal(err)
	}

	inputs = make(map[int]*Input)

	for i, name := range src {
		ictx, err := gmf.NewInputCtx(name)
		if err != nil {
			log.Fatal(err)
		}

		inputs[i] = &Input{}
		inputs[i].ctx = ictx

		log.Printf("Source #%d - %s\n", i, name)
	}
	defer func() {
		for i, _ := range inputs {
			inputs[i].ctx.Free()
		}
	}()

	srcVideoStream, err := inputs[0].ctx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Fatalf("No video stream found\n")
	} else {
		i, o := addStream("libx264", octx, srcVideoStream)
		streamMap[i] = o
	}

	srcStreams := []*gmf.Stream{}

	for i, _ := range inputs {
		for idx := 0; idx < inputs[i].ctx.StreamsCnt(); idx++ {
			stream, err := inputs[i].ctx.GetStream(idx)
			if err != nil {
				log.Fatalf("error getting stream - %s\n", err)
			}

			if !stream.IsVideo() {
				continue
			}

			srcStreams = append(srcStreams, stream)
		}
	}

	options := []*gmf.Option{}

	var (
		i, ret int = 0, 0
		pkt    *gmf.Packet
		ist    *gmf.Stream
		ost    *gmf.Stream
	)

	for i, stream := range srcStreams {
		log.Printf("stream #%d - %s, %s\n", i, stream.CodecCtx().Codec().LongName(), stream.CodecCtx().GetVideoSize())
	}

	ost, err = octx.GetStream(0)
	if err != nil {
		log.Fatalf("can't get stream - %s\n", err)
	}

	filter, err := gmf.NewFilter("overlay=10:main_h-overlay_h-10", srcStreams, ost, options)
	defer filter.Release()
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	if err := octx.WriteHeader(); err != nil {
		log.Fatalf("error writing header - %s\n", err)
	}
	defer octx.Free()

	init := false

	var (
		frame *gmf.Frame
		ff    []*gmf.Frame
	)

	for {
		if finishedNb() == len(inputs) {
			log.Printf("Finished all\n")
			break
		}

		if i == len(inputs) {
			i = 0
		}

		if inputs[i].finished {
			i++
			continue
		}

		ictx := inputs[i].ctx

		pkt, err = ictx.GetNextPacket()
		if err != nil && err != io.EOF {
			log.Fatalf("error getting next packet - %s", err)
		} else if err != nil && pkt == nil {
			if !inputs[i].finished {
				log.Printf("EOF input #%d, closing\n", i)
				filter.RequestOldest()
				filter.Close(i)
				inputs[i].finished = true
			}
			i++
			continue
		}

		ist, err = ictx.GetStream(pkt.StreamIndex())
		if err != nil {
			log.Fatalf("%s\n", err)
		}

		if ist.IsAudio() {
			continue
		}

		frame, ret = ist.CodecCtx().Decode2(pkt)
		if ret < 0 && gmf.AvErrno(ret) == syscall.EAGAIN {
			continue
		} else if ret == gmf.AVERROR_EOF {
			log.Fatalf("EOF in Decode2, handle it\n")
		} else if ret < 0 {
			log.Fatalf("Unexpected error - %s\n", gmf.AvError(ret))
		}

		frame.SetPts(ist.Pts)
		ist.Pts++

		if frame != nil && !init {
			if err := filter.AddFrame(frame, i, 0); err != nil {
				log.Fatalf("%s\n", err)
			}
			i++
			init = true
			continue
		}

		if err := filter.AddFrame(frame, i, 4); err != nil {
			log.Fatalf("%s\n", err)
		}

		frame.Free()

		if ff, err = filter.GetFrame(); err != nil && len(ff) == 0 {
			log.Printf("GetFrame() returned '%s', continue\n", err)
			i++
			continue
		}

		if len(ff) == 0 {
			i++
			continue
		}

		packets, err := ost.CodecCtx().Encode(ff, -1)
		if err != nil {
			log.Fatalf("%s\n", err)
		}

		for _, f := range ff {
			f.Free()
		}

		for _, op := range packets {
			gmf.RescaleTs(op, ost.CodecCtx().TimeBase(), ost.TimeBase())
			op.SetStreamIndex(ost.Index())

			if err = octx.WritePacket(op); err != nil {
				break
			}

			op.Free()
		}

		i++
	}

	octx.WriteTrailer()

	ost.CodecCtx().Free()
	ost.Free()

	for _, v := range inputs {
		for i := 0; i < v.ctx.StreamsCnt(); i++ {
			st, _ := v.ctx.GetStream(i)
			st.CodecCtx().Free()
			st.Free()
		}
	}
}
