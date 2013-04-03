package gmf

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"
)

func decoder_test(track *Track, finish chan bool) {
	decoder := track.GetDecoder()
	decoder.Open()
	var p *Packet = new(Packet)
	for true {
		if !track.ReadPacket(p) {
			println("stream end reached")
			break
		}
		frame := decoder.Decode(p)

		//fmt.Printf("frame:%d codecid:%d\n",frame,decoder.Ctx.ctx.codec_id)
		if frame != nil && frame.isFinished {
			if decoder.Ctx.ctx.codec_type == CODEC_TYPE_VIDEO {
				//ppsWriter(frame)
			}
		}
	}
	finish <- true
}

func TestDecoderTracks(t *testing.T) {
	println("starting decoder test")
	//loc:=MediaLocator{Filename:"/media/video/ChocolateFactory.ts"}
	loc := MediaLocator{Filename: "../../../target/dependency/fixtures/testfile.flv"}

	//loc:=MediaLocator{Filename:"/media/TREKSTOR/videos/20070401 0140 - PREMIERE 3 - Ein Duke kommt selten allein (The Dukes of Hazzard).ts"}
	source := DataSource{Locator: loc}
	result := source.Connect()
	println(result)
	if result != nil {
		t.Fatalf("cold not open file : %s", loc.Filename)

	}
	plex := NewDemultiplexer(&source)
	//tracks:=plex.GetTracks()
	tracks := plex.GetTracks()
	finish := make(chan bool)
	for i := 0; i < len(tracks); i++ {
		//dec:=tracks[i].GetDecoder()
		//dec.Open()
		go decoder_test(&tracks[i], finish)
	}
	plex.Start()
	for i := 0; i < len(tracks); i++ {
		/*waiting for all tracks to be finished*/
		<-finish
	}
	//println(len(tracks))
	//println(plexer.GetTimestamp().String())
	source.Disconnect()
	println(" decoder test finished")

}

func TestReadParameters(t *testing.T) {
	println("read parameter decoder test start")
	loc := MediaLocator{Filename: "../../../target/dependency/fixtures/testfile.flv"}

	source := DataSource{Locator: loc}
	result := source.Connect()
	if result != nil {
		t.Fatalf("cold not open file : %s", loc.Filename)

	}
	plex := NewDemultiplexer(&source)
	//tracks:=plex.GetTracks()
	tracks := plex.GetTracks()
	//    finish:=make(chan bool)
	for i := 0; i < len(tracks); i++ {
		dec := tracks[i].GetDecoder()
		dec.Open()
		fmt.Printf("%s", dec.GetParameters())
		//go decoder_test(&tracks[i], finish)
	}
	//plex.Start()
	//println(len(tracks))
	//println(plexer.GetTimestamp().String())
	//source.Disconnect()
	println("read parameter decoder test finished")
}

func TestSerializeDecoder(t *testing.T) {
	println("serialize decoder test start")
	loc := MediaLocator{Filename: "../../../target/dependency/fixtures/testfile.flv"}

	source := DataSource{Locator: loc}
	result := source.Connect()
	if result != nil {
		t.Fatalf("cold not open file : %s", loc.Filename)

	}
	plex := NewDemultiplexer(&source)
	//tracks:=plex.GetTracks()
	tracks := plex.GetTracks()
	for i := 0; i < len(tracks); i++ {
		dec := tracks[i].GetDecoder()
		//dec.Open()
		params := dec.GetParameters()
		dec2 := Decoder{}
		for k, v := range params {
			dec2.SetParameter(k, v)
		}
		dec2.Open()
		gob.RegisterName("NewTypeObject", Packet{})
		b := new(bytes.Buffer)
		enc := gob.NewEncoder(b)
		err := enc.Encode(dec)
		if err != nil {
			fmt.Printf("%s\n", err)
			//return err
			break
		}
		bdec := gob.NewDecoder(b)
		err = bdec.Decode(dec)
		if err != nil {
			fmt.Printf("%s\n", err)
			break
			//return err
		}

	}
	println("serialize decoder test finished")

}
