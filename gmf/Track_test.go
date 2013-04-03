package gmf


import "testing"

//import "runtime"
func read(track *Track) {
	var p Packet
	for true {
		if !track.ReadPacket(&p) {
			println("stream end reached")
			return
		} else {
			//println("packet readed")
			p.Free()
		}
	}
}

func TestReadTracks(t *testing.T) {
	println("start func TestReadTracks(t*testing.T){")
	//loc:=MediaLocator{Filename:"/media/video/ChocolateFactory.ts"}
	loc := MediaLocator{Filename: "../../../target/dependency/fixtures/testfile.flv"}
	//loc:=MediaLocator{"/media/TREKSTOR/videos/20070401 0140 - PREMIERE 3 - Ein Duke kommt selten allein (The Dukes of Hazzard).ts"}
	source := DataSource{Locator: loc}
	if source.Connect() != nil {
		t.Fatalf("cold not open file : %s", loc.Filename)
	}
	plex := NewDemultiplexer(&source)
	//tracks:=plex.GetTracks()
	tracks := plex.GetTracks()
	for i := 0; i < len(tracks); i++ {
		//dec:=tracks[i].GetDecoder()
		//dec.Open()
		go read(&tracks[i])
	}
	plex.Start()
	//println(len(tracks))
	//println(plexer.GetTimestamp().String())
	source.Disconnect()
	println("end func TestReadTracks(t*testing.T){")
	//println("successful reader test")
}
/*
var number int=0
func ppsWriter(frame * Frame){
   header:=fmt.Sprintf("P5\n%d %d\n255\n",frame.width,frame.height)

   file, err := os.Open("test.ppm", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)

   if(err!=nil){
   println("error while openning file")
   }
   file.WriteString(header)
   linesize:=int(frame.avframe.linesize[0])
   //println(linesize)
   data:=make([]byte, frame.width)
   tmpdata:=(*(*[1<<30]byte)(unsafe.Pointer(frame.avframe.data[0])))[0:frame.size]
   //    for i:= 0; i < 1; i++ {
   //      data[i] = tmpdata[i];
   //    }

   for i := 0; i < int(frame.height); i++{
   for a:= 0; a < int(frame.width); a++ {
   data[a] = tmpdata[(i*linesize)+a];
   }
   file.Write(data);
   }

   file.Close()
    if(number>1000){
    os.Exit(1)
    }
    number++
}
*/
