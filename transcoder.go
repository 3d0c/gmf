package main

import (
	"flag"
	"fmt"
	"gmf/gmf"
	"log"
	"os"
	"time"
)

func process_track(track *gmf.Track, multiplexer *gmf.Multiplexer) {
	var encoder *gmf.Encoder
	var resizer *gmf.Resizer
	var resampler *gmf.Resampler
	var rate_converter *gmf.FrameRateConverter
	var deinterlacer *gmf.Deinterlacer

	decoder := track.GetDecoder()
	decoder.SetParameter("request_channels", "2")
	decoder.SetParameter("request_channel_layout", "2")
	decoder.Open()
	defer decoder.Close()
	if decoder.GetCodecType() == gmf.CODEC_TYPE_VIDEO {
		encoder = new(gmf.Encoder)

		encoder.SetParameter("codecid", "mpeg4")
		encoder.SetParameter("time_base", "1/25")
		encoder.SetParameter("b", "512000")
		encoder.SetParameter("width", "320")
		encoder.SetParameter("height", "240")

		encoder.SetParameter("flags", "global_header")

		encoder.Open()
		defer encoder.Close()

		resizer = new(gmf.Resizer)
		resizer.Init(decoder, encoder)
		defer resizer.Close()

		rate_converter = new(gmf.FrameRateConverter)
		rate_converter.Init(decoder.GetFrameRate(), encoder.GetFrameRate())
		rate_converter.Init(decoder.GetFrameRate(), decoder.GetFrameRate())

		fmt.Printf("decoder FR: %v, encoder FR: %v\n", decoder.GetFrameRate(), encoder.GetFrameRate())

		deinterlacer = new(gmf.Deinterlacer)
		deinterlacer.Init(decoder)

		multiplexer.AddTrack(encoder)
	}
	if decoder.GetCodecType() == gmf.CODEC_TYPE_AUDIO {

		encoder = new(gmf.Encoder)
		encoder.SetParameter("codecid", "mp2")
		encoder.SetParameter("ab", "128000")
		encoder.SetParameter("ar", "44100")
		encoder.SetParameter("ac", "2")
		encoder.SetParameter("flags", "global_header")
		encoder.Open()
		defer encoder.Close()
		resampler = new(gmf.Resampler)
		resampler.Init(decoder, encoder)
		defer resampler.Close()
		multiplexer.AddTrack(encoder)

		return
	}
	p := new(gmf.Packet)
	for true {
		//stream end reached break this loop, no more processing is needed
		if !track.ReadPacket(p) {
			break
		}
		var frame *gmf.Frame
		frame = decoder.Decode(p)
		if frame != nil && frame.IsFinished() && encoder != nil {
			typ := decoder.GetCodecType()
			switch typ {
			case gmf.CODEC_TYPE_VIDEO:
				frame = deinterlacer.Deinterlace(frame)
				frame = resizer.Resize(frame)
				frame = rate_converter.Convert(frame)
			case gmf.CODEC_TYPE_AUDIO:
				frame = resampler.Resample(frame)
			}
			encoder.Encode(frame)
		}
	}
}

var inputfile *string = flag.String("i", "", "input file name")
var outfile *string = flag.String("o", "", "output file name")
var outformat *string = flag.String("f", "", "output file format")

func main() {
	flag.Parse()

	if inputfile == nil {
		fmt.Printf("no inputfile given")
		os.Exit(1)
	}
	if outfile == nil {
		fmt.Printf("no outputfile given")
		os.Exit(1)
	}

	//prepare the input file
	source := gmf.NewDataSource(gmf.MediaLocator{Filename: *inputfile})
	err := source.Connect()

	if err != nil {
		log.Printf("%s : %s\n", err, source.Locator.Filename)
		os.Exit(1)
	}
	demultiplexer := gmf.NewDemultiplexer(source)

	//prepare the output file
	var sink = gmf.NewDatasink(gmf.MediaLocator{Filename: *outfile, Format: *outformat})
	err = sink.Connect()

	if err != nil {
		log.Printf("%s : %s\n", err, sink.Locator.Filename)
		os.Exit(1)
	}
	multiplexer := gmf.NewMultiplexer(sink)

	//prepare processing each track
	tracks := demultiplexer.GetTracks()
	fmt.Printf("tracks: %v\n", tracks)

	for i := 0; i < len(tracks); i++ {
		go process_track(&tracks[i], multiplexer)
	}

	time.Sleep(2000000000)

	multiplexer.Start()
	demultiplexer.Start()

	time.Sleep(1000000000)
	multiplexer.Stop()
	demultiplexer.Stop()

	source.Disconnect()
	sink.Disconnect()
}
