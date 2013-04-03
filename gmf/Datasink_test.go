package gmf


import "testing"


func TestDatasinkConnect(t *testing.T) {
	println("start func TestDatasourceConnect(t*testing.T){")
	//loc:=MediaLocator{Filename:"/media/video/ChocolateFactory.ts"}
	loc := MediaLocator{Filename: "testsink.flv"}
	source := DataSink{Locator: loc}
	if source.Connect() != nil {
		t.Errorf("cold not open file : %s", loc.Filename)
	}

	source.Disconnect()
	println("end func TestDatasourceConnect(t*testing.T){")

}
