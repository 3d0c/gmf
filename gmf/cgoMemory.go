package gmf

import (
	"log"
	"sync/atomic"
)

type CgoMemoryManage struct {
	retainCount int32
}

type CgoMemoryManager interface {
	Retain()
	RetainCount() int32
	Release()
	Free()
}

func Retain(i CgoMemoryManager) CgoMemoryManager {
	//func Retain(i CgoMemoryManager) interface {} {
	i.Retain()
	return i
}

func Release(i CgoMemoryManager) {
	if nil == i {
		return
	}
	i.Release()
	if 0 >= i.RetainCount() {
		i.Free()
	}
}

func debugLogf(functionname string, c *CgoMemoryManage) {
	if false {
		log.Printf("CgoMemoeryMangaer "+functionname+"(%p) retainCount=%d", c, c.RetainCount())
	}
}

func (c *CgoMemoryManage) Retain() {
	atomic.AddInt32(&c.retainCount, 1)
	debugLogf("Retain", c)
}

func (c *CgoMemoryManage) RetainCount() int32 {
	return c.retainCount + 1
}
func (c *CgoMemoryManage) Release() {
	atomic.AddInt32(&c.retainCount, -1)
	debugLogf("Release", c)
}

func (c *CgoMemoryManage) Free() {
	debugLogf("Free", c)
}
