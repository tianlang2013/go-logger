package log

import "testing"
type data struct {
	A int
}
func TestLogger(t *testing.T){
	Error( "test" ,"why" , data{11})
	Warn( "warn" ,"why" , data{11})
	Info( "info" ,"why" , data{11})
	SetLevel(LvlTrace)
	Debug( "debug" ,"why" , data{11})
	Trace( "trace" ,"why" , data{11})
}
