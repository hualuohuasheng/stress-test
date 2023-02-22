package stress_strategy

import (
	"github.com/hualuohuasheng/stress-test/client"
	"github.com/hualuohuasheng/stress-test/model"
	"github.com/panjf2000/ants/v2"
	"github.com/spf13/viper"
	"log"
	"strconv"
	"sync"
	"time"
)

func init() {
	// 注册HTTP响应的验证函数
	model.RegisterVerifyHTTP("json", model.VerifyHttpJson)
	model.RegisterVerifyHTTP("statusCode", model.VerifyHttpStatusCode)
}

type Task struct {
	id      uint64
	request *model.Request
	logger  *log.Logger
	results chan<- *model.RequestResults
}

func PressHttpServer(i interface{}) {
	t := i.(Task)
	client.PressHttp(t.id, t.request, t.results, t.logger)
}

func Dispose(concurrency, conNumber int, cfg *viper.Viper, debug, keepLongCon bool, maxConns int, logger, resLogger *log.Logger) {
	//eachPoolNum := cfg.GetInt("EachPoolNum")
	//poolNum := cfg.GetInt("PoolNum")
	requestURL := cfg.GetString("Url")
	method := cfg.GetString("Method")
	headers := cfg.GetStringSlice("Headers")
	body := cfg.GetString("Body")
	eachNew := cfg.GetBool("NewRequestEachTime")
	jobDoSleep := cfg.GetInt("DoSleepTime")
	var reqHeaders = make(map[string]string)
	model.GetHeaders(headers, reqHeaders)
	request, err := model.NewRequest(requestURL, method, "json", "0", 0, debug, keepLongCon, reqHeaders, body)
	if err != nil {
		resLogger.Printf("参数不合法 %v \n", err)
		return
	}
	totalNumber := concurrency * conNumber
	resLogger.Println("")
	resLogger.Printf("并发数: %v, 每个并发的请求: %v, 共计%v个请求, 间隔: %.3f ms\n", concurrency, conNumber, totalNumber, float64(jobDoSleep)/1000)
	resLogger.Printf("准备进行压测的url: %v\n", request.URL)
	// 先生成请求参数
	requestList := make([]*model.Request, 0)
	sTm := time.Now()
	if eachNew {
		for i := 0; i < totalNumber; i++ {
			var reqH = make(map[string]string)
			model.GetHeaders(headers, reqH)
			headerTime := strconv.Itoa(int(time.Now().UnixNano() / 1000))
			newBody := model.NewRTSSubmitOrderBody(headerTime, cfg, debug)
			model.NewRTSHeaders(headerTime, newBody, cfg, reqH, debug)
			request, _ = model.NewRequest(requestURL, method, "json", "0", 0, debug, keepLongCon, reqH, newBody)
			requestList = append(requestList, request)
		}
	} else {
		requestList = append(requestList, request)
	}
	eTm := time.Now()
	resLogger.Println("生成请求所需的耗时: ", eTm.Sub(sTm))

	// 接收response响应的通道
	repCh := make(chan *model.RequestResults, totalNumber)
	var wg sync.WaitGroup    // 发送数据
	var wgRec sync.WaitGroup // 接收数据
	wgRec.Add(1)
	go ReceivingResultsShowRealTimeTPS(uint64(concurrency), uint64(totalNumber), repCh, &wgRec, resLogger)
	if keepLongCon {
		client.CreateLongHttpClient(maxConns)
	}
	// 创建指定数量的协程池
	pool, _ := ants.NewPoolWithFunc(concurrency, func(i interface{}) {
		PressHttpServer(i)
		wg.Done()
	})
	defer pool.Release()

	for i := 0; i < totalNumber; i++ {
		wg.Add(1)
		var task Task
		task.id = uint64(i)
		if eachNew {
			task.request = requestList[i]
		} else {
			task.request = requestList[0]
		}
		task.results = repCh
		task.logger = logger
		_ = pool.Invoke(task)
		time.Sleep(time.Duration(jobDoSleep) * time.Microsecond)
	}
	wg.Wait()
	//fmt.Printf("总耗时: %.6f 请求数: %d, tps: %.4f\n", tt, totalNumber, float64(totalNumber)/tt)
	//time.Sleep(3 * time.Millisecond)
	wgRec.Wait()
	close(repCh)
	//wgRec.Wait()
}
