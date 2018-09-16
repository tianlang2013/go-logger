# go-logger
log for golang  on logger,support level and persional
在原生logger包的基础上改变而来
支持级别控制
默认基本为INFO，高级别自动显示低级别的日志，但是低级别的日志是不显示高级别的日志
修改日志在没有另外实例化的情况下，是影响设置后的代码显示，建议只在项目开始处进行配置
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

···
error	2018/09/15 23:10:22 logger_test.go:8: test [why {11}]
warn	2018/09/15 23:10:22 logger_test.go:9: warn [why {11}]
info	2018/09/15 23:10:22 logger_test.go:10: info [why {11}]
debug	2018/09/15 23:10:22 logger_test.go:12: debug [why {11}]
trace	2018/09/15 23:10:22 logger_test.go:13: trace [why {11}]
···
as you see that didn`t print debug and trace if you didn`t set the level as trace 
of cause that didn`t print trace if you set level as debug
