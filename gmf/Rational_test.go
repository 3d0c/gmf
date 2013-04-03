package gmf


import "testing"

func TestRational(t *testing.T) {
	var R1 Rational = Rational{1, 2}
	var R2 Rational = Rational{2, 4}
	var R3 Rational = Rational{2, 3}
	var R4 Rational = NewRational()
	var R5 Rational = Rational{1, 3}
	var R6 Rational = Rational{1, 4}
	var R7 Rational = Rational{3, 6}
	var R8 Rational = Rational{2, 6}
	if !R1.Equals(R2) {
		t.Error("fail !R1.Equals(R2)")
	}

	if !R1.Lower(R3) {
		t.Error("fail R1.Lower(R3)")
	}
	if R1.Lower(R5) {
		t.Error("fail R1.Lower(R5)")
	}
	if !R5.Lower(R1) {
		t.Error("fail R5.Lower(R1)")
	}
	if R5.Greater(R1) {
		t.Error("fail R5.Greater(R1)")
	}

	if R3.Lower(R1) {
		t.Error("fail R3.Lower(R1)")
	}

	if R1.Greater(R3) {
		t.Error("fail R1.Greater(R3)")
	}

	if !R1.Greater(R4) {
		t.Error("fail !R1.Greater(R4)")
	}

	if !R1.Greater(R6) {
		t.Error("fail !R1.Greater(R6)")
	}

	if !R7.Greater(R8) {
		t.Error("fail !R7.Greater(R8)")
	}
}
