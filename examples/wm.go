package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

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
	ctx      *gmf.FmtCtx
	finished bool
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

	// Create Video stream in output context
	if ost = oc.NewStream(codec); ost == nil {
		log.Fatal(errors.New("unable to create stream in output context"))
	}
	defer gmf.Release(ost)

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		log.Fatal(errors.New("unable to create codec context"))
	}
	defer gmf.Release(cc)

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
		ost.SetTimeBase(gmf.AVR{1, 25})
		cc.SetProfile(gmf.FF_PROFILE_MPEG4_SIMPLE)
		fmt.Printf("setup dims: %d, %d\n", ist.CodecCtx().Width(), ist.CodecCtx().Height())
		cc.SetDimension(ist.CodecCtx().Width(), ist.CodecCtx().Height())
		cc.SetPixFmt(ist.CodecCtx().PixFmt())
	}

	if err := cc.Open(nil); err != nil {
		log.Fatal(err)
	}

	ost.SetCodecCtx(cc)

	return ist.Index(), ost.Index()
}

func main() {
	var (
		src       arrayFlags
		dst       string
		streamMap map[int]int = make(map[int]int)
	)

	log.SetFlags(log.Lshortfile)

	flag.Var(&src, "src", "source files, e.g.: -src=1.mp4 -src=image.png")
	flag.StringVar(&dst, "dst", "", "destination file, e.g. -dst=result.mp4")
	flag.Parse()

	if len(src) == 0 || dst == "" {
		log.Fatal("at least one source and destination required")
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

		log.Printf("src[%d]=%s\n", i, name)
	}

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
	/*
			{
				Key: "pix_fmts", Val: []int32{gmf.AV_PIX_FMT_YUV420P},
			},
		}
	*/

	var (
		i   int = 0
		pkt *gmf.Packet
		ist *gmf.Stream
		ost *gmf.Stream
	)

	for i, stream := range srcStreams {
		fmt.Printf("srcStreams[%d] - %s, %d\n", i, stream.CodecCtx().Codec().LongName(), stream.CodecCtx().Width())
	}

	filtered := make([]*gmf.Frame, 0)

	ost, err = octx.GetStream(0)
	if err != nil {
		log.Fatalf("can't get stream - %s\n", err)
	}

	filter, err := gmf.NewFilter("overlay=10:main_h-overlay_h-10", srcStreams, ost, options)
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	if err := octx.WriteHeader(); err != nil {
		log.Fatalf("error writing header - %s\n", err)
	}

	fmt.Printf("ost TimeBase: %v, %s\n", ost.TimeBase(), ost.CodecCtx().Codec().LongName())

	total := 0
	init := false

	for {
		total++
		fmt.Printf("Total: %d\n", total)
		if total == 30 {
			break
		}
		fmt.Printf("i=%d\n", i)

		if finishedNb() == len(inputs) {
			log.Printf("finished all\n")
			break
		}

		if i == len(inputs) {
			i = 0
		}

		if inputs[i].finished {
			fmt.Printf("inputs[%d] - finished\n", i)
			i++
			continue
		}

		ictx := inputs[i].ctx

		pkt, err = ictx.GetNextPacket()
		if err != nil && err.Error() != "End of file" {
			log.Fatalf("error getting next packet - %s", err)
		} else if err != nil && pkt == nil {
			fmt.Printf("continue getting next packets - %s\n", err)
			inputs[i].finished = true
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

		frames, err := ist.CodecCtx().Decode(pkt)
		if err != nil && err.Error() != "End of file" {
			log.Fatalf("error decoding - %s\n", err)
		} else if err != nil {
			log.Printf("error decoding pkt - %s\n", err)
		}

		fmt.Printf("len(frames) = %d\n", len(frames))

		if len(frames) > 0 && !init {
			if err := filter.AddFrame(frames[0], i); err != nil {
				log.Fatalf("%s\n", err)
			}
			i++
			init = true
			continue
		}

		for _, frame := range frames {
			log.Printf("frame: %dX%d", frame.Width(), frame.Height())

			if err := filter.AddFrame(frame, i); err != nil {
				log.Fatalf("%s\n", err)
			}

			ff, err := filter.GetFrame()
			if err != nil {
				log.Printf("err != nil - %s\n", err)
				continue
			}
			if len(ff) == 0 {
				log.Printf("len(ff) = 0\n")
				filtered = nil
				break
			}

			for idx, f := range ff {
				log.Printf("ff[%d]: %dX%d\n", idx, f.Width(), f.Height())
			}

			filtered = append(filtered, ff...)
		}

		if len(filtered) == 0 {
			log.Printf("len(filtered) == 0\n")
			i++
			continue
		}

		packets, err := ost.CodecCtx().Encode(filtered, -1)
		if err != nil {
			log.Fatalf("%s\n", err)
		}

		for _, op := range packets {
			gmf.RescaleTs(op, ist.TimeBase(), ost.TimeBase())
			op.SetStreamIndex(ost.Index())

			if err = octx.WritePacket(op); err != nil {
				break
			}

			op.Free()
		}

		i++
		fmt.Printf("------------------------\n")
	}

}
