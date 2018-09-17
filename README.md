# go-logger
log for golang  on logger,support level and persional
在原生logger包的基础上改变而来
支持级别控制
默认基本为INFO，高级别自动显示低级别的日志，但是低级别的日志是不显示高级别的日志
修改日志在没有另外实例化的情况下，是影响设置后的代码显示，建议只在项目开始处进行配置

自带行号和日志发生时的文件

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

```

```
base	error	2018/09/17 15:28:25 logger_test.go:11: test [why {11}]
base	warn	2018/09/17 15:28:25 logger_test.go:12: warn [why {11}]
base	info	2018/09/17 15:28:25 logger_test.go:13: info [why {11}]
base	debug	2018/09/17 15:28:25 logger_test.go:15: debug [why {11}]
test	error	2018/09/17 15:28:25 logger_test.go:20: test [why {11}]
test	warn	2018/09/17 15:28:25 logger_test.go:21: warn [why {11}]
test	info	2018/09/17 15:28:25 logger_test.go:22: info [why {11}]
test	debug	2018/09/17 15:28:25 logger_test.go:23: debug [why {11}]
test	debug	2018/09/17 15:28:25 logger_test.go:25: debug [why {11}]
```
as you see that didn`t print debug and trace if you didn`t set the level as trace 
of cause that didn`t print trace if you set level as debug
