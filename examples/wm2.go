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

func multiplex(inputs []*gmf.FmtCtx, filter *gmf.Filter) <-chan *gmf.Frame {
	var (
		out      chan *gmf.Frame = make(chan *gmf.Frame)
		lf       []*gmf.Frame    = make([]*gmf.Frame, len(inputs), len(inputs))
		ff       []*gmf.Frame    = make([]*gmf.Frame, 0)
		finished map[int]bool    = map[int]bool{}
		pkt      *gmf.Packet
		ist      *gmf.Stream
		err      error
		ret      int
	)

	go func(out chan *gmf.Frame) {
	loop:
		for {
			for y := 0; y < len(inputs); y++ {
				ictx := inputs[y]

				pkt, err = ictx.GetNextPacket()
				if err != nil && err.Error() != "End of file" {
					log.Fatalf("error getting next packet - %s", err)
					break loop
				} else if err != nil && pkt == nil {
					if lf == nil {
						log.Fatalf("last frame is not initialized\n")
					}

					finished[y] = true

					tmp := lf[y].CloneNewFrame()
					if err := filter.AddFrame(tmp, y); err != nil {
						log.Fatalf("%s\n", err)
					}

					ff, ret = filter.GetFrame2()
				} else {
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

					for i, _ := range frames {
						lf[y] = frames[i].CloneNewFrame()

						if err := filter.AddFrame(frames[i], y); err != nil {
							log.Fatalf("%s\n", err)
						}
					}

					ff, ret = filter.GetFrame2()
				}

				log.Printf("ret=%d\n", ret)

				if len(finished) == len(inputs) {
					break loop
				}
			}

			// for i, _ := range ff {
			// 	out <- ff[i]
			// }
		}
		close(out)
	}(out)

	return out
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

	inputs := make([]*gmf.FmtCtx, 0)

	for i, name := range src {
		ictx, err := gmf.NewInputCtx(name)
		if err != nil {
			log.Fatal(err)
		}

		inputs = append(inputs, ictx)
		log.Printf("src[%d]=%s\n", i, name)
	}

	srcVideoStream, err := inputs[0].GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
	if err != nil {
		log.Fatalf("No video stream found\n")
	} else {
		i, o := addStream("libx264", octx, srcVideoStream)
		streamMap[i] = o
	}

	srcStreams := []*gmf.Stream{}

	for i, _ := range inputs {
		for idx := 0; idx < inputs[i].StreamsCnt(); idx++ {
			stream, err := inputs[i].GetStream(idx)
			if err != nil {
				log.Fatalf("error getting stream - %s\n", err)
			}

			if !stream.IsVideo() {
				continue
			}

			srcStreams = append(srcStreams, stream)
		}
	}

	options := []*gmf.Option{
		/*
			{
				Key: "pix_fmts", Val: []int32{gmf.AV_PIX_FMT_YUV420P},
			},
		*/
	}

	var (
		ist *gmf.Stream
		ost *gmf.Stream
	)

	for i, stream := range srcStreams {
		fmt.Printf("srcStreams[%d] - %s, %d\n", i, stream.CodecCtx().Codec().LongName(), stream.CodecCtx().Width())
	}

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

	for frame := range multiplex(inputs, filter) {
		packets, err := ost.CodecCtx().Encode([]*gmf.Frame{frame}, -1)
		if err != nil {
			log.Fatalf("%s\n", err)
		}

		for _, op := range packets {
			gmf.RescaleTs(op, ist.TimeBase(), ost.TimeBase())
			op.SetStreamIndex(ost.Index())
			op.Dump()
			if err = octx.WritePacket(op); err != nil {
				break
			}

			op.Free()
		}
	}

	octx.WriteTrailer()
}
