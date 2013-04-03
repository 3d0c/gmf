package gmf


import "testing"

func TestTimestamp(t *testing.T) {
	t2 := Timestamp{100, Rational{1, 2}}
	//var d Duration=Duration{t2}
	t2.Greater(t2)
	//if(d.Greater(d)){

	//}
}
