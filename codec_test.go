package gmf

import (
	"log"
	"testing"
)

func _TestCodec(t *testing.T) {
	type test struct {
		In       interface{}
		Expected interface{}
	}

	tests := []test{
		{In: 13, Expected: "mpeg4"},
		{In: "mpeg4", Expected: "mpeg4"},
		{In: "mp2", Expected: "mp2"},
	}

	for _, i := range tests {
		c, err := NewEncoder(i.In)
		if err != nil || c.Name() != i.Expected {
			t.Fatalf("Unexpected error occured. In: %v, Error: %v\n", i.In, err)
		} else {
			log.Printf("Expected codec '%v:%v' found by value '%v'\n", c.Name(), c.LongName(), i.In)
		}
	}
}

func _TestNewPacket(t *testing.T) {
	expectedPts := -9223372036854775808
	p := NewPacket()
	if p.Pts() != expectedPts {
		t.Fatalf("Expected pts %d, %d got.\n", expectedPts, p.Pts())
	} else {
		log.Printf("Expected pts %d found, packet initialized.\n", expectedPts)
	}
}
