package gmf


import "testing"


func TestAvMalloc(t *testing.T) {
	buffer := av_malloc(10)
	println(buffer)
	av_free(buffer)
}
