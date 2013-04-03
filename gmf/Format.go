package gmf


type Format interface{}


//Encapsulates format information for video data. 
//The attributes of a VideoFormat include the encoding type, frame size, frame rate, and the data type. 
type VideoFormat struct {
	FrameRate Rational //Real base framerate of the stream. .
	TimeBase  Rational //This is the fundamental unit of time (in seconds) in terms of which frame timestamps are represented.
	Width     int      //A Dimension that specifies the frame width.
	Height    int      //A Dimension that specifies the frame height.
}


//Encapsulates format information for audio data. 
//The attributes of an AudioFormat include the sample rate, bits per sample, and number of channels. 
type AudioFormat struct {
	Channels   int //The number of channels as an integer.
	FrameSize  int //The frame size of this AudioFormat in bits.
	SampleRate int //The sample rate.
	SampleSize int //The sample size in bits.
}
