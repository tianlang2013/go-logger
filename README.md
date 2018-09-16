# go-logger
log for golang  on logger,support level and persional

#Default
the info is default info if you didn`t set .
you could set the level as :
```
const (
	LvlCrit int32= iota
	LvlError
	LvlWarn
	LvlInfo
	LvlDebug
	LvlTrace
)

 ```
 
#example
```
type data struct {
	A int
}
func TestLogger(t *testing.T){
	Error( "test" ,"why" , data{11})
	Warn( "warn" ,"why" , data{11})
	Info( "info" ,"why" , data{11})
	SetLevel(LvlTrace)    //set logger Level
	Debug( "debug" ,"why" , data{11})
	Trace( "trace" ,"why" , data{11})
}

```

as you see that didn`t print debug and trace if you didn`t set the level as trace 
of cause that didn`t print trace if you set level as debug
