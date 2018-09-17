package log

import (
	"testing"
	"os"
)
type data struct {
	A int
}
func TestLogger(t *testing.T){
	Error( "test" ,"why" , data{11})
	Warn( "warn" ,"why" , data{11})
	Info( "info" ,"why" , data{11})
	SetLevel(LvlDebug)
	Debug( "debug" ,"why" , data{11})
	Trace( "trace" ,"why" , data{11})

	log := New( os.Stderr , Lshortfile |  LstdFlags  , "test	")

	log.Error( "test" ,"why" , data{11})
	log.Warn( "warn" ,"why" , data{11})
	log.Info( "info" ,"why" , data{11})
	log.Debug( "debug" ,"why" , data{11})
	log.Trace( "trace" ,"why" , data{11})
	log.Debug( "debug" ,"why" , data{11})
	log.Trace( "trace" ,"why" , data{11})

}
