package main

import (
	"flag"
	stress_strategy "github.com/hualuohuasheng/stress-test/stress-strategy"
	"github.com/hualuohuasheng/stress-test/tools"
	"log"
	"runtime"
	"strings"
)

//// array 自定义数组参数
//type array []string
//
//// String string
//func (a *array) String() string {
//	return fmt.Sprint(*a)
//}
//
//// Set set
//func (a *array) Set(s string) error {
//	*a = append(*a, s)
//
//	return nil
//}

var (
	concurrency     = 200     // 并发数
	conNumber       = 30      // 单个并发的请求数(单个并发/协程)
	app             = "rts"   // 运行的压测程序名称, 配置文件的顶级节点
	debugStr        = "false" // 是否调试
	cfgFileName     = "rts21" // 运行的压测配置文件
	keepLongConnect = "true"  // 是否保持长链接
	//cfgFilePath = "."  // 压测配置文件路径
	infoLogFile = "log_info" // 日志文件路径
	maxCon      = 500        // 对host的最大连接数量

	infoLogger   *log.Logger
	resultLogger *log.Logger
)

func init() {
	flag.IntVar(&concurrency, "c", concurrency, "并发数, 启动的协程数量")
	flag.IntVar(&conNumber, "n", conNumber, "单个并发的请求数")
	flag.StringVar(&app, "app", app, "压测程序对象, 配置文件的顶级节点")
	flag.StringVar(&debugStr, "d", debugStr, "是否调试，默认false")
	flag.StringVar(&cfgFileName, "cf", cfgFileName, "运行的配置文件名称")
	flag.StringVar(&keepLongConnect, "keep", keepLongConnect, "是否保持长连接")
	flag.StringVar(&infoLogFile, "logName", infoLogFile, "日志文件路径, 默认为当前目录")
	flag.IntVar(&maxCon, "m", maxCon, "host的最大连接数量")

	// 解析参数
	flag.Parse()

	//infoName := fmt.Sprintf("%s_%s", infoLogFile, time.Now().Format("2006-01-02T15:04:05"))
	infoLogger = tools.CreateLogger(".", infoLogFile, "INFO ", false)
	resultLogger = tools.CreateLogger(".", "stress_test_result", "INFO ", true)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	cfg := tools.LoadConfiguration(cfgFileName, ".", app)
	debug := strings.ToLower(debugStr) == "true"
	keep := strings.ToLower(keepLongConnect) == "true"
	//fmt.Println(keep)
	stress_strategy.Dispose(concurrency, conNumber, cfg, debug, keep, maxCon, infoLogger, resultLogger)
}
