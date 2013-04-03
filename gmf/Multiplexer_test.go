package gmf

import _ "net/http/pprof"
import "time"

//import  "http"

import "testing"

func multiplex_encoder_test(track *Track, multiplexer *Multiplexer) {
	var encoder Encoder
	decoder := track.GetDecoder()
	decoder.Open()
	var outvideoTrack *Track
	var resizer *Resizer
	if decoder.Ctx.ctx.codec_type == CODEC_TYPE_VIDEO {
		println("Create encoder")
		encoder.SetParameter("codecid", "22")
		encoder.SetParameter("time_base", "1/25")
		encoder.SetParameter("width", "320")
		encoder.SetParameter("height", "240")
		encoder.SetParameter("bf", "0")
		encoder.SetParameter("b", "500000")
		encoder.SetParameter("g", "250")
		encoder.SetParameter("qmin", "2")
		encoder.SetParameter("qmax", "51")
		encoder.SetParameter("qdiff", "4")
		encoder.SetParameter("flags", "+global_header")
		encoder.Open()
		outvideoTrack = multiplexer.AddTrack(&encoder)
		print(outvideoTrack)
		resizer = new(Resizer)
		resizer.Init(decoder, &encoder)
	}

	var p Packet
	for true {
		if !track.ReadPacket(&p) {
			println("stream end reached")
			break
		}
		frame := decoder.Decode(&p)
		p.Free()
		//fmt.Printf("frame:%d codecid:%d\n",frame,decoder.Ctx.ctx.codec_id)
		if frame != nil && frame.isFinished {
			//println("frame finished")
			//ppsWriter(frame)
			if decoder.Ctx.ctx.codec_type == CODEC_TYPE_VIDEO {
				frame.avframe.pict_type = 0
				frame.avframe.key_frame = 1
				of := resizer.Resize(frame)
				encoder.Encode(of)
				//outvideoTrack.WritePacket(op)
				//op.destroy()
			}
		}
	}
}

func TestMultiplexer(t *testing.T) {

	println("starting func TestMultiplexer(t*testing.T){")
	/*
		    go func (){
		    err := http.ListenAndServe(":6060", nil)
		    if err != nil {
			panic("ListenAndServe: " + err.String())


			}
			println("listen")
		    }()*/
	//loc:=MediaLocator{Filename:"/media/video/ChocolateFactory.ts"}
	//loc:=MediaLocator{Filename:"/media/TREKSTOR/videos/20070401 0140 - PREMIERE 3 - Ein Duke kommt selten allein (The Dukes of Hazzard).ts"}
	loc := MediaLocator{Filename: "../../../test.dvd"}
	//loc:=MediaLocator{Filename:"/Users/jholscher/Movies/39,90.avi.divx"}
	source := DataSource{Locator: loc}
	if source.Connect() != nil {
		t.Fatalf("cold not open file : %s", loc.Filename)
	}

	var mloc = MediaLocator{Filename: "testmultiplexer.flv", Format: "flv"}
	var sink = DataSink{Locator: mloc}
	sink.Connect()
	var multiplexer = Multiplexer{Ds: sink}

	plex := NewDemultiplexer(&source)
	//tracks:=plex.GetTracks()
	tracks := plex.GetTracks()
	for i := 0; i < len(tracks); i++ {
		//dec:=tracks[i].GetDecoder()
		//dec.Open()
		go multiplex_encoder_test(&tracks[i], &multiplexer)
	}
	time.Sleep(10000000)
	go multiplexer.Start()
	plex.Start()
	time.Sleep(10000000)
	multiplexer.Stop()

	//plex.Stop()
	//println(len(tracks))
	//println(plexer.GetTimestamp().String())
	source.Disconnect()
	println(" encoder func TestMultiplexer(t*testing.T){")
}
