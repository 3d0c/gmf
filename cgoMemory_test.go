package gmf

import (
	//	"log"
	"testing"
)

type SubData struct {
	CgoMemoryManage
	Id   int32
	Name [16]byte
}

type Data struct {
	SubData
	abc int
}

var dataIsFree bool = false
var subDataIsFree bool = false

func (d *Data) Free() {
	//	log.Printf(" Data Free(%p) retainCount=%d",d,d.RetainCount())
	dataIsFree = true
}

func (d *SubData) Free() {
	//	log.Printf(" SubData Free(%p) retainCount=%d",d,d.RetainCount())
	subDataIsFree = true
}

func TestCgoMemory(t *testing.T) {

	var abc *Data

	abc = new(Data)
	sss := SubData{Id: 12}

	Retain(abc)

	if 2 != abc.RetainCount() {
		t.Fatal("has error.")
	}

	Release(abc)

	if 1 != abc.RetainCount() {
		t.Fatal("has error.")
	}

	Release(abc)

	if 0 != abc.RetainCount() {
		t.Fatal("has error.")
	}

	if !dataIsFree {
		t.Fatal("Data not run Free.")
	}

	Release(&sss)

	if 0 != sss.RetainCount() {
		t.Fatal("has error.")
	}

	if !subDataIsFree {
		t.Fatal("subData not run Free.")
	}
}
