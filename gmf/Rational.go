package gmf

import "fmt"

type Rational struct {
	Num int
	Den int
}

func NewRational() Rational {
	var result Rational = Rational{1, 1000000}
	return result
}

func (r *Rational) String() string {
	return fmt.Sprintf("%d/%d", r.Num, r.Den)
}

func (r *Rational) Equals(r2 Rational) bool {
	return av_cmp_q(*r, r2) == 0
}

func (r *Rational) Greater(r2 Rational) bool {
	return av_cmp_q(*r, r2) > 0
}

func (r *Rational) Lower(r2 Rational) bool {
	return av_cmp_q(*r, r2) < 0
}
