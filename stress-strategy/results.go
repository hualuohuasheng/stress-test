package stress_strategy

import (
	"fmt"
	"github.com/hualuohuasheng/stress-test/model"
	"github.com/hualuohuasheng/stress-test/tools"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"log"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// 输出统计数据的时间
	exportStatisticsTime = 1 * time.Second
	p                    = message.NewPrinter(language.English)
	requestTimeList      []uint64 // 所有请求响应时间
	startTimeList        []uint64 // 所有请求的发起时间
	endTimeList          []uint64 // 所有请求的响应时间
)

//func ReceivingResults(concurrent uint64, ch <-chan *model.RequestResults, wg *sync.WaitGroup) {
//	defer func() {
//		wg.Done()
//	}()
//	var stopChan = make(chan bool)
//	// 时间
//	var (
//		processingTime uint64 // 处理总时间
//		requestTime    uint64 // 请求总时间
//		maxTime        uint64 // 最大时长
//		minTime        uint64 // 最小时长
//		successNum     uint64 // 成功处理数，code为0
//		failureNum     uint64 // 处理失败数，code不为0
//		chanIDLen      int    // 并发数
//		chanIDs        = make(map[uint64]bool)
//		receivedBytes  int64
//		mutex          = sync.RWMutex{}
//	)
//	statTime := uint64(time.Now().UnixNano())
//
//	// 错误码/错误个数
//	var errCode = &sync.Map{}
//	// 定时输出一次计算结果
//	ticker := time.NewTicker(exportStatisticsTime)
//	go func() {
//		for {
//			select {
//			case <-ticker.C:
//				endTime := uint64(time.Now().UnixNano())
//				mutex.Lock()
//				go calculateData(concurrent, processingTime, endTime-statTime, maxTime, minTime, successNum, failureNum, chanIDLen, errCode, receivedBytes)
//				mutex.Unlock()
//			case <-stopChan:
//				// 处理完成
//				return
//			}
//		}
//	}()
//	header()
//	for data := range ch {
//		mutex.Lock()
//		//fmt.Println("处理一条数据", data.ID, data.Time, data.IsSucceed, data.ErrCode, len(requestTimeList))
//		processingTime = processingTime + data.Time
//		if maxTime <= data.Time {
//			maxTime = data.Time
//		}
//		if minTime == 0 {
//			minTime = data.Time
//		} else if minTime > data.Time {
//			minTime = data.Time
//		}
//		// 是否请求成功
//		if data.IsSucceed == true {
//			successNum = successNum + 1
//		} else {
//			failureNum = failureNum + 1
//		}
//		// 统计错误码
//		if value, ok := errCode.Load(data.ErrCode); ok {
//			valueInt, _ := value.(int)
//			errCode.Store(data.ErrCode, valueInt+1)
//		} else {
//			errCode.Store(data.ErrCode, 1)
//		}
//		receivedBytes += data.ReceivedBytes
//		if _, ok := chanIDs[data.ChanID]; !ok {
//			chanIDs[data.ChanID] = true
//			chanIDLen = len(chanIDs)
//		}
//		requestTimeList = append(requestTimeList, data.Time)
//		mutex.Unlock()
//	}
//	// 数据全部接受完成，停止定时输出统计数据
//	stopChan <- true
//	endTime := uint64(time.Now().UnixNano())
//	requestTime = endTime - statTime
//	calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum, chanIDLen, errCode,
//		receivedBytes)
//
//	fmt.Printf("\n\n")
//	fmt.Println("*************************  结果 stat  ****************************")
//	fmt.Println("处理协程数量:", concurrent)
//	// fmt.Println("处理协程数量:", concurrent, "程序处理总时长:", fmt.Sprintf("%.3f", float64(processingTime/concurrent)/1e9), "秒")
//	fmt.Println("请求总数（并发数*请求数 -c * -n）:", successNum+failureNum, "总请求时间:",
//		fmt.Sprintf("%.3f", float64(requestTime)/1e9),
//		"秒", "successNum:", successNum, "failureNum:", failureNum)
//	printTop(requestTimeList)
//	fmt.Println("*************************  结果 end   ****************************")
//	fmt.Printf("\n\n")
//}

func ReceivingResultsShowRealTimeTPS(concurrent, works uint64, ch <-chan *model.RequestResults, wg *sync.WaitGroup, resLogger *log.Logger) {
	defer func() {
		wg.Done()
	}()
	var stopChan = make(chan bool)
	// 时间
	var (
		tprocessingTime uint64 // 处理总时间
		requestTime     uint64 // 请求总时间
		//maxTime        uint64 // 最大时长
		//minTime        uint64 // 最小时长
		tsuccessNum = uint64(0) // 成功处理数，code为0
		tfailureNum = uint64(0) // 处理失败数，code不为0
		//chanIDLen      int    // 并发数
		//chanIDs        = make(map[uint64]bool)
		receivedBytes int64
		mutex         = sync.RWMutex{}
	)
	statTime := uint64(time.Now().UnixNano())

	// 错误码/错误个数
	var errCode = &sync.Map{}
	// 定时输出一次计算结果
	ticker := time.NewTicker(exportStatisticsTime)
	go func() {
		for {
			select {
			case <-ticker.C:
				endTime := uint64(time.Now().UnixNano())
				mutex.Lock()
				var (
					processingTime uint64     // 处理总时间
					maxTime        uint64     // 最大时长
					minTime        uint64     // 最小时长
					successNum     uint64 = 0 // 成功处理数，code为0
					failureNum     uint64 = 0 // 处理失败数，code不为0
					chanIDLen      int        // 并发数
					chanIDs        = make(map[uint64]bool)
				)
				curChNum := len(ch)
				for i := 0; i < curChNum; i++ {
					data := <-ch
					startTimeList = append(startTimeList, uint64(data.RequestTime.UnixNano()))
					endTimeList = append(endTimeList, uint64(data.ResponseTime.UnixNano()))
					repTime := uint64(data.ResponseTime.Sub(data.RequestTime).Nanoseconds())
					processingTime = processingTime + repTime
					tprocessingTime = tprocessingTime + repTime
					if maxTime <= repTime {
						maxTime = repTime
					}
					if minTime == 0 {
						minTime = repTime
					} else if minTime > repTime {
						minTime = repTime
					}
					// 是否请求成功
					if data.IsSucceed == true {
						atomic.AddUint64(&successNum, 1)
						atomic.AddUint64(&tsuccessNum, 1)
					} else {
						atomic.AddUint64(&failureNum, 1)
						atomic.AddUint64(&tfailureNum, 1)
						//failureNum = failureNum + 1
						//tfailureNum = tfailureNum + 1
					}
					// 统计错误码
					if value, ok := errCode.Load(data.ErrCode); ok {
						valueInt, _ := value.(int)
						errCode.Store(data.ErrCode, valueInt+1)
					} else {
						errCode.Store(data.ErrCode, 1)
					}
					receivedBytes += data.ReceivedBytes
					if _, ok := chanIDs[data.ChanID]; !ok {
						chanIDs[data.ChanID] = true
						chanIDLen = len(chanIDs)
					}
					requestTimeList = append(requestTimeList, repTime)
					works -= 1

				}

				go calculateData(concurrent, processingTime, endTime-statTime, maxTime, minTime, successNum, failureNum, chanIDLen, errCode, receivedBytes, tsuccessNum+tfailureNum, resLogger)
				mutex.Unlock()
			case <-stopChan:
				// 处理完成
				return
			}
		}
	}()
	header(resLogger)

	// 数据全部接受完成，停止定时输出统计数据
	for {
		if works <= uint64(0) {
			stopChan <- true
			break
		}
	}
	time.Sleep(time.Millisecond)
	requestTime = findRequestTime(startTimeList, endTimeList)
	//fmt.Printf("总耗时: %.fns\n", requestTime)
	//calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum, chanIDLen, errCode,
	//	receivedBytes)
	//
	resLogger.Println("")
	resLogger.Println("*************************  结果 stat  ****************************")
	resLogger.Println("处理协程数量:", concurrent)
	resLogger.Println("处理协程数量:", concurrent, "程序处理总时长:", fmt.Sprintf("%.3f", float64(tprocessingTime/concurrent)/1e9), "秒")
	resLogger.Println("请求总数:", tsuccessNum+tfailureNum, "总请求时间:",
		fmt.Sprintf("%.4f", float64(requestTime)/1e9),
		"秒", "successNum:", tsuccessNum, "failureNum:", tfailureNum, "平均tps: ", float64(tsuccessNum)/(float64(requestTime)/1e9))
	resLogger.Println("")
	printTop(requestTimeList, resLogger)
	resLogger.Println("*************************  结果 end   ****************************")
}

func calculateData(concurrent, processingTime, requestTime, maxTime, minTime, successNum, failureNum uint64,
	chanIDLen int, errCode *sync.Map, receivedBytes int64, totalRequest uint64, resLogger *log.Logger) {
	if processingTime == 0 {
		processingTime = 1
	}
	var (
		qps              float64
		averageTime      float64
		maxTimeFloat     float64
		minTimeFloat     float64
		requestTimeFloat float64
	)
	//fmt.Println("每秒打印: ", processingTime, requestTime)
	// 平均 每个协程成功数*总协程数据/总耗时 (每秒)
	if processingTime != 0 {
		qps = float64(successNum*1e9*concurrent) / float64(processingTime)
	}
	// 平均时长 总耗时/总请求数/并发数 纳秒=>毫秒
	if successNum != 0 && concurrent != 0 {
		averageTime = float64(processingTime) / float64(successNum*1e6)
	}
	// 纳秒=>毫秒
	maxTimeFloat = float64(maxTime) / 1e6
	minTimeFloat = float64(minTime) / 1e6
	requestTimeFloat = float64(requestTime) / 1e9
	// 打印的时长都为毫秒
	table(totalRequest, successNum, failureNum, errCode, qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat, chanIDLen,
		receivedBytes, resLogger)
}

// header 打印表头信息
func header(resLogger *log.Logger) {
	//fmt.Printf("\n\n")
	//// 打印的时长都为毫秒 总请数
	//fmt.Println("─────┬────────┬──────────┬──────┬──────┬────────┬────────────┬────────────┬────────────┬────────┬────────┬────────")
	//fmt.Println(" 耗时│总并发数│当前并发数│成功数│失败数│   tps  │最长耗时(ms)│最短耗时(ms)│平均耗时(ms)│下载字节│字节每秒│ 状态码")
	//fmt.Println("─────┼────────┼──────────┼──────┼──────┼────────┼────────────┼────────────┼────────────┼────────┼────────┼────────")
	resLogger.Println("")
	resLogger.Println("─────┬────────┬──────────┬──────┬──────┬────────┬────────────┬────────────┬────────────┬────────┬────────┬────────")
	resLogger.Println(" 耗时│总并发数│当前并发数│成功数│失败数│   tps  │最长耗时(ms)│最短耗时(ms)│平均耗时(ms)│下载字节│字节每秒│ 状态码")
	resLogger.Println("─────┼────────┼──────────┼──────┼──────┼────────┼────────────┼────────────┼────────────┼────────┼────────┼────────")
	return
}

// table 打印表格
func table(totalRequest, successNum, failureNum uint64, errCode *sync.Map,
	qps, averageTime, maxTimeFloat, minTimeFloat, requestTimeFloat float64, chanIDLen int, receivedBytes int64, resLogger *log.Logger) {
	var (
		speed int64
	)
	if requestTimeFloat > 0 {
		speed = int64(float64(receivedBytes) / requestTimeFloat)
	} else {
		speed = 0
	}
	var (
		receivedBytesStr string
		speedStr         string
	)
	// 判断获取下载字节长度是否是未知
	if receivedBytes <= 0 {
		receivedBytesStr = ""
		speedStr = ""
	} else {
		receivedBytesStr = p.Sprintf("%d", receivedBytes)
		speedStr = p.Sprintf("%d", speed)
	}
	// 打印的时长都为毫秒
	result := fmt.Sprintf("%4.0fs│%8d│%10d│%6d│%6d│%8.2f│%12.2f│%12.2f│%12.2f│%8s│%8s│%v",
		requestTimeFloat, totalRequest, chanIDLen, successNum, failureNum, qps, maxTimeFloat, minTimeFloat, averageTime,
		receivedBytesStr, speedStr,
		printMap(errCode))
	//fmt.Println(result)
	resLogger.Println(result)
	return
}

func getNewSortedTimeList(timeList []uint64) tools.MyUint64List {
	stList := tools.MyUint64List{}
	stList = timeList
	sort.Sort(stList)
	return stList
}

func findRequestTime(startTimeList, endTimeList []uint64) uint64 {
	stList := getNewSortedTimeList(startTimeList)
	etList := getNewSortedTimeList(endTimeList)
	//fmt.Println(etList[len(stList)-1] - stList[0])
	return etList[len(stList)-1] - stList[0]
}

func printTop(requestTimeList []uint64, resLogger *log.Logger) {
	if requestTimeList == nil {
		return
	}
	all := tools.MyUint64List{}
	all = requestTimeList
	sort.Sort(all)
	resLogger.Println("tp90响应时间:", fmt.Sprintf("%.3f", float64(all[int(float64(len(all))*0.90)]/1e6)), "ms")
	resLogger.Println("tp95响应时间:", fmt.Sprintf("%.3f", float64(all[int(float64(len(all))*0.95)]/1e6)), "ms")
	resLogger.Println("tp99响应时间:", fmt.Sprintf("%.3f", float64(all[int(float64(len(all))*0.99)]/1e6)), "ms")
}

// printMap 输出错误码、次数 节约字符(终端一行字符大小有限)
func printMap(errCode *sync.Map) (mapStr string) {
	var (
		mapArr []string
	)
	errCode.Range(func(key, value interface{}) bool {
		mapArr = append(mapArr, fmt.Sprintf("%v:%v", key, value))
		return true
	})
	sort.Strings(mapArr)
	mapStr = strings.Join(mapArr, ";")
	return
}
