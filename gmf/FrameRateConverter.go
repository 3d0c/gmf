package gmf

import "log"

type FrameRateConverter struct {
	src             Rational
	trg             Rational
	frame_list      []*Frame
	inframes        int
	outframes       int
	compensate_base float32
}


func (self *FrameRateConverter) Init(src, trg Rational) {
	self.src = src
	self.trg = trg
	self.frame_list = make([]*Frame, 100)
	log.Printf("FrameRateConverter src = %s trg = %s", self.src, self.trg)
}

func (self *FrameRateConverter) Convert(in *Frame) *Frame {
	if self.src.Equals(self.trg) {
		return in
	}
	result := in
	result.frame_count = 0
	self.inframes++
	var cts float32 = ((((float32(self.inframes) / float32(self.src.Den)) * float32(self.src.Num)) / float32(self.trg.Num)) * float32(self.trg.Den))
	cts += self.compensate_base
	src := Timestamp{1, self.src}
	trg := Timestamp{1, self.trg}
	if src.Lower(trg) {
		if int(cts)-self.outframes > 0 {
			result.frame_count = 1
			self.outframes++
		}
	} else {
		if src.Greater(trg) {
			result.frame_count = 1
			if cts-float32(self.outframes) > 2 {
				result.frame_count = int(cts) - self.outframes
			}
		}
		self.outframes += result.frame_count
	}
	//log.Printf("CTS:%f inframes %d outframes %d frame_count %d frame_diff %f",cts,self.inframes, self.outframes, result.frame_count,cts-float32(self.outframes))
	//print(cts)
	return result
}
