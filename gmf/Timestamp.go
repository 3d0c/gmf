package gmf


import "fmt"

type Timestamp struct {
	Time     int64
	Timebase Rational
}


func (time *Timestamp) RescaleTo(base Rational) Timestamp {
	result := Timestamp{0, base}
	result.Time = av_rescale_q(time.Time, time.Timebase, base)
	return result
}


func (time *Timestamp) Greater(ts Timestamp) bool {
	return av_compare_ts(time.Time, time.Timebase, ts.Time, ts.Timebase) == 1
}

func (time *Timestamp) Lower(ts Timestamp) bool {
	return av_compare_ts(time.Time, time.Timebase, ts.Time, ts.Timebase) == -1
}

func (time *Timestamp) Equals(ts Timestamp) bool {
	return av_compare_ts(time.Time, time.Timebase, ts.Time, ts.Timebase) == 0
}

func (time Timestamp) String() string {
	return fmt.Sprintf("%d:%d/%d", time.Time, time.Timebase.Num, time.Timebase.Den)
}

type Duration struct {
	Timestamp
}


func (dur *Duration) RescaleTo(base Rational) Timestamp {
	return dur.Timestamp.RescaleTo(base)
}


func (dur *Duration) Greater(ts Duration) bool {
	return dur.Timestamp.Greater(ts.Timestamp)
}

func (dur *Duration) Lower(ts Duration) bool {
	return dur.Timestamp.Lower(ts.Timestamp)
}

func (dur *Duration) Equals(ts Duration) bool {
	return dur.Timestamp.Equals(ts.Timestamp)
}

func (dur Duration) String() string {
	return dur.Timestamp.String()
}
