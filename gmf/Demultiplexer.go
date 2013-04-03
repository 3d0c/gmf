package gmf

import "unsafe"

//import "runtime"

type Demultiplexer struct {
	ds     DataSource
	tracks []Track
}

//Gets the duration of this media stream when played at the default rate.
//Note that each track can have a different duration and a different start time. This method returns the total duration from when the first track starts and the last track ends.
func (dpx *Demultiplexer) GetDuration() Duration {
	return Duration{Timestamp{int64(dpx.ds.ctx.ctx.duration), Rational{1, 1000000}}}
}

//Gets the current media time. This is the stream position that the next readFrame will read.
func (dpx *Demultiplexer) GetTimestamp() Timestamp {
	return Timestamp{int64(dpx.ds.ctx.ctx.timestamp), Rational{1, 1000000}}
}

//Retrieves the individual tracks that the media stream contains.
//A stream can contain multiple media tracks, such as separate tracks for audio and video data.
//The Track interface also provides methods for enabling or disabling a track.
// returns An array of Track objects. The length of the array is equal to the number of tracks in the stream.
func (dpx *Demultiplexer) GetTracks() []Track {
	scount := dpx.ds.ctx.ctx.nb_streams
	var result []Track = make([]Track, int(scount))
	for i := 0; i < int(scount); i++ {
		//var streams []Stream =dpx.ds.ctx.ctx.streams[]
		result[i] = Track{av_get_stream(dpx.ds.ctx.ctx, i), make(chan Packet), 0}
	}
	dpx.tracks = result
	return result
}

//Signals that data is going to start being read from the Demultiplexer.
//The start method is called before any calls are made to readFrame.
func (dpx *Demultiplexer) Start() {
	avpacket := new(avPacket)
	av_init_packet2(avpacket)
	for {
		if av_read_frame(dpx.ds.ctx, avpacket) < 0 {
			println("end of file reached, closing channels")
			for i := 0; i < len(dpx.tracks); i++ {
				print("closing channel ")
				println(i)
				close((dpx.tracks)[i].stream)
			}
			break
		}

		track := (dpx.tracks)[avpacket.stream_index]

		packet := new(Packet)

		packet.Size = int(avpacket.size)
		// packet.Pts = Timestamp{int64(avpacket.pts), Rational{int(track.time_base.num), int(track.time_base.den)}}
		// packet.Dts = Timestamp{int64(avpacket.dts), Rational{int(track.time_base.num), int(track.time_base.den)}}
		// packet.Duration = Timestamp{int64(avpacket.duration), Rational{int(track.time_base.num), int(track.time_base.den)}}
		packet.Data = make([]byte, packet.Size+8)

		data := (*(*[1 << 30]byte)(unsafe.Pointer(avpacket.data)))[0:packet.Size]
		for i := 0; i < packet.Size; i++ {
			packet.Data[i] = data[i]
		}
		av_free_packet2(avpacket)
		track.stream <- *packet
	}
}
func (dpx *Demultiplexer) Stop() {
	//NOOP
}

//Creates a new Demutliplexer from a DataSource
func NewDemultiplexer(ds *DataSource) *Demultiplexer {
	return &Demultiplexer{ds: *ds}
}
