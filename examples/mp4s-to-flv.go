package main

import (
	"flag"
	"fmt"
	"io"
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

func addStream(name string, oc *gmf.FmtCtx, ist *gmf.Stream) (int, int, error) {
	var (
		cc      *gmf.CodecCtx
		ost     *gmf.Stream
		options []gmf.Option
	)

	codec, err := gmf.FindEncoder(name)
	if err != nil {
		return 0, 0, err
	}

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		return 0, 0, fmt.Errorf("unable to create codec context")
	}

	if oc.IsGlobalHeader() {
		cc.SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	}

	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	if cc.Type() == gmf.AVMEDIA_TYPE_AUDIO {
		options = append(
			[]gmf.Option{
				{Key: "time_base", Val: ist.CodecCtx().TimeBase().AVR()},
				{Key: "ar", Val: ist.CodecCtx().SampleRate()},
				{Key: "ac", Val: ist.CodecCtx().Channels()},
				{Key: "channel_layout", Val: cc.SelectChannelLayout()},
			},
		)

		cc.SetSampleFmt(ist.CodecCtx().SampleFmt())
		cc.SelectSampleRate()
	}

	if cc.Type() == gmf.AVMEDIA_TYPE_VIDEO {
		options = append(
			[]gmf.Option{
				{Key: "time_base", Val: gmf.AVR{Num: 1, Den: 25}},
				{Key: "pixel_format", Val: gmf.AV_PIX_FMT_YUV420P},
				// Save original
				{Key: "video_size", Val: ist.CodecCtx().GetVideoSize()},
				{Key: "b", Val: 500000},
			},
		)

		cc.SetProfile(ist.CodecCtx().GetProfile())
	}

	cc.SetOptions(options)

	if err := cc.Open(nil); err != nil {
		return 0, 0, err
	}

	par := gmf.NewCodecParameters()
	if err = par.FromContext(cc); err != nil {
		return 0, 0, fmt.Errorf("error creating codec parameters from context - %s", err)
	}

	if ost = oc.NewStream(codec); ost == nil {
		return 0, 0, fmt.Errorf("unable to create new stream in output context")
	}

	ost.SetCodecParameters(par)
	ost.SetCodecCtx(cc)

	if cc.Type() == gmf.AVMEDIA_TYPE_VIDEO {
		ost.SetTimeBase(gmf.AVR{Num: 1, Den: 25})
		ost.SetRFrameRate(gmf.AVR{Num: 25, Den: 1})
	}

	return ist.Index(), ost.Index(), nil
}

func main() {
	var (
		src       arrayFlags
		dst       string
		streamMap map[int]int = make(map[int]int)
		pts       int64       = 0
	)

	flag.Var(&src, "src", "source files, e.g.: -src=1.mp4 -src=2.mp4")
	flag.StringVar(&dst, "dst", "", "destination file, e.g. -dst=result.flv")
	flag.Parse()

	if len(src) == 0 || dst == "" {
		log.Fatal("at least one source and destination required")
	}

	octx, err := gmf.NewOutputCtx(dst)
	if err != nil {
		log.Fatal(err)
	}

	for _, name := range src {
		log.Printf("Processing file %s\n", name)

		ictx, err := gmf.NewInputCtx(name)
		if err != nil {
			log.Fatal(err)
		}

		if len(streamMap) == 0 {
			srcVideoStream, err := ictx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
			if err != nil {
				log.Fatalf("No video stream found in %s", name)
			} else {
				i, o, err := addStream("libx264", octx, srcVideoStream)
				if err != nil {
					log.Fatal(err)
				}
				streamMap[i] = o
			}

			srcAudioStream, err := ictx.GetBestStream(gmf.AVMEDIA_TYPE_AUDIO)
			if err != nil {
				log.Println("No audio stream found in", name)
			} else {
				i, o, err := addStream("aac", octx, srcAudioStream)
				if err != nil {
					log.Fatal(err)
				}
				streamMap[i] = o
			}

			if err := octx.WriteHeader(); err != nil {
				log.Fatalf("error writing header - %s\n", err)
			}
		}

		var (
			pkt       *gmf.Packet
			ist, ost  *gmf.Stream
			streamIdx int
			flush     int = -1
		)

		for {
			if flush < 0 {
				pkt, err = ictx.GetNextPacket()
				if err != nil && err != io.EOF {
					if pkt != nil {
						pkt.Free()
					}
					log.Fatalf("error getting next packet - %s", err)
				} else if err != nil && pkt == nil {
					log.Printf("=== flushing \n")
					flush++
					break
				}
			}

			if flush == len(streamMap) {
				break
			}

			if flush < 0 {
				streamIdx = pkt.StreamIndex()
			} else {
				streamIdx = flush
				flush++
			}

			if _, ok := streamMap[streamIdx]; !ok {
				if pkt != nil {
					pkt.Free()
				}

				continue
			}

			ist, err = ictx.GetStream(streamIdx)
			if err != nil {
				if pkt != nil {
					pkt.Free()
				}
				log.Fatalf("error getting stream %d - %s", pkt.StreamIndex(), err)
			}

			ost, err = octx.GetStream(streamMap[ist.Index()])
			if err != nil {
				if pkt != nil {
					pkt.Free()
				}
				log.Fatalf("error getting stream %d - %s", pkt.StreamIndex(), err)
			}

			frames, err := ist.CodecCtx().Decode(pkt)
			if err != nil {
				log.Fatalf("error decoding - %s\n", err)
			}

			for _, frame := range frames {
				frame.SetPts(pts)
				pts++
			}

			packets, err := ost.CodecCtx().Encode(frames, flush)

			for _, op := range packets {
				gmf.RescaleTs(op, ost.CodecCtx().TimeBase(), ost.TimeBase())
				op.SetStreamIndex(ost.Index())

				if err = octx.WritePacket(op); err != nil {
					break
				}

				op.Free()
			}

			for _, frame := range frames {
				if frame != nil {
					frame.Free()
				}
			}

			if pkt != nil {
				pkt.Free()
			}
		}

		ictx.Free()
	}

	octx.WriteTrailer()
	octx.Free()

}
