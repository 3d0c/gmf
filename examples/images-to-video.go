package main

/* Valgrind report summary

==6002== LEAK SUMMARY:
==6002==    definitely lost: 0 bytes in 0 blocks
==6002==    indirectly lost: 0 bytes in 0 blocks
==6002==      possibly lost: 1,152 bytes in 4 blocks
==6002==    still reachable: 0 bytes in 0 blocks
==6002==         suppressed: 0 bytes in 0 blocks

*/

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/3d0c/gmf"
)

var pts int64 = 0

func initOst(name string, oc *gmf.FmtCtx, ist *gmf.Stream) (*gmf.Stream, error) {
	var (
		cc      *gmf.CodecCtx
		ost     *gmf.Stream
		options []gmf.Option
	)

	codec, err := gmf.FindEncoder(name)
	if err != nil {
		return nil, err
	}

	if cc = gmf.NewCodecCtx(codec); cc == nil {
		return nil, fmt.Errorf("unable to create codec context")
	}

	if oc.IsGlobalHeader() {
		cc.SetFlag(gmf.CODEC_FLAG_GLOBAL_HEADER)
	}

	if codec.IsExperimental() {
		cc.SetStrictCompliance(gmf.FF_COMPLIANCE_EXPERIMENTAL)
	}

	options = append(
		[]gmf.Option{
			{Key: "time_base", Val: gmf.AVR{Num: 1, Den: 25}},
		},
	)

	options = append(
		[]gmf.Option{
			{Key: "pixel_format", Val: gmf.AV_PIX_FMT_YUV420P},
			{Key: "video_size", Val: ist.CodecCtx().GetVideoSize()},
		},
		options...,
	)

	cc.SetProfile(gmf.FF_PROFILE_H264_MAIN)
	cc.SetOptions(options)

	if err := cc.Open(nil); err != nil {
		return nil, err
	}

	par := gmf.NewCodecParameters()
	if err = par.FromContext(cc); err != nil {
		return nil, fmt.Errorf("error creating codec parameters from context - %s", err)
	}
	defer par.Free()

	if ost = oc.NewStream(codec); ost == nil {
		return nil, fmt.Errorf("unable to create new stream in output context")
	}

	ost.CopyCodecPar(par)
	ost.SetCodecCtx(cc)
	ost.SetTimeBase(gmf.AVR{Num: 1, Den: 25})
	ost.SetRFrameRate(gmf.AVR{Num: 25, Den: 1})

	return ost, nil
}

func main() {
	var (
		src     string
		dst     string
		ost     *gmf.Stream
		pkt     *gmf.Packet
		frame   *gmf.Frame
		swsCtx  *gmf.SwsCtx
		ret     int
		sources []string = make([]string, 0)
	)

	flag.StringVar(&src, "src", "./tmp", "source images folder")
	flag.StringVar(&dst, "dst", "result.mp4", "destination file")

	flag.Parse()

	fis, err := ioutil.ReadDir(src)
	if err != nil {
		log.Fatalf("Error reding '%s' - %s\n", src, err)
	}

	for _, fi := range fis {
		ext := filepath.Ext(fi.Name())
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			log.Printf("skipping %s, ext: '%s'\n", fi.Name(), ext)
			continue
		}

		sources = append(sources, filepath.Join(src, fi.Name()))
	}

	if len(sources) == 0 {
		log.Fatalf("Not enough source files\n")
	}

	octx, err := gmf.NewOutputCtx(dst)
	if err != nil {
		log.Fatalf("Error creating output context - %s\n", err)
	}
	defer octx.Free()

	for _, source := range sources {
		log.Printf("Loading %s\n", source)

		ictx, err := gmf.NewInputCtx(source)
		if err != nil {
			log.Fatalf("Error creating input context - %s\n", err)
		}

		ist, err := ictx.GetBestStream(gmf.AVMEDIA_TYPE_VIDEO)
		if err != nil {
			log.Fatalf("Error getting source stream - %s\n", err)
		}

		if ost == nil {
			if ost, err = initOst("libx264", octx, ist); err != nil {
				log.Fatalf("Error init output stream - %s\n", err)
			}
			if err = octx.WriteHeader(); err != nil {
				log.Fatalf("%s\n", err)
			}
		}

		if swsCtx == nil {
			icc := ist.CodecCtx()
			occ := ost.CodecCtx()
			if swsCtx, err = gmf.NewSwsCtx(icc.Width(), icc.Height(), icc.PixFmt(), occ.Width(), occ.Height(), occ.PixFmt(), gmf.SWS_BICUBIC); err != nil {
				panic(err)
			}
			defer swsCtx.Free()
		}

		if pkt, err = ictx.GetNextPacket(); err != nil {
			log.Fatalf("Error getting packet - %s", err)
		}

		frame, ret = ist.CodecCtx().Decode2(pkt)
		if ret < 0 {
			log.Fatalf("Unexpected error - %s\n", gmf.AvError(ret))
		}

		dstFrames, err := gmf.DefaultRescaler(swsCtx, []*gmf.Frame{frame})
		if err != nil {
			log.Fatalf("Error scaling - %s\n", err)
		}

		encode(octx, ost, dstFrames[0], -1)

		pkt.Free()
		frame.Free()
		dstFrames[0].Free()

		ist.CodecCtx().Free()
		ist.Free()
		ictx.Free()
	}

	encode(octx, ost, nil, 1)

	ost.CodecCtx().Free()
	ost.Free()

	octx.WriteTrailer()
}

func encode(octx *gmf.FmtCtx, ost *gmf.Stream, frame *gmf.Frame, drain int) {
	if frame != nil {
		frame.SetPts(pts)
	}

	pts += 1

	packets, err := ost.CodecCtx().Encode([]*gmf.Frame{frame}, drain)
	if err != nil {
		log.Fatalf("Error encoding - %s\n", err)
	}
	if len(packets) == 0 {
		return
	}

	for _, packet := range packets {
		packet.SetPts(gmf.RescaleQ(packet.Pts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))
		packet.SetDts(gmf.RescaleQ(packet.Dts(), ost.CodecCtx().TimeBase(), ost.TimeBase()))

		if err = octx.WritePacket(packet); err != nil {
			log.Fatalf("Error writing packet - %s\n", err)
		}

		packet.Free()

		if drain > 0 {
			pts += 1
		}
	}

	return
}
