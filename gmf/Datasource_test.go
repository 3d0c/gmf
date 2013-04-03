package gmf


import "testing"


func TestDatasourceConnect(t *testing.T) {
	println("start func TestDatasourceConnect(t*testing.T){")
	//loc:=MediaLocator{Filename:"/media/video/ChocolateFactory.ts"}
	loc := MediaLocator{Filename: "../../../target/dependency/fixtures/testfile.flv"}
	source := DataSource{Locator: loc}
	if source.Connect() != nil {
		t.Errorf("cold not open file : %s", loc.Filename)
	}

	source.Disconnect()
	println("end func TestDatasourceConnect(t*testing.T){")

}
