package gmf

type Buffer struct {
	//the internal data object that holds the media chunk contained in this Buffer.
	Data []byte
	//the length of the valid data in this Buffer if the data is held in an array.
	Length int
	//the duration of this Buffer.
	Duration Duration
	//the mask of the flags set for this Buffer.
	Flags int
	//the Format of the data in this Buffer.
	Format Format
	//the sequence number of this Buffer.
	SequenceNumber int64
	//Checks whether or not this Buffer is to be discarded.
	Discard bool
	//Checks whether or not this Buffer marks the end of the media stream.
	EOM bool
	//the time stamp of this Buffer.
	TimeStamp Timestamp

	isFinished bool
}
