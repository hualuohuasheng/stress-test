package client

import (
	"fmt"
	"github.com/hualuohuasheng/stress-test/model"
	"log"
	"net/http"
	"time"
)

func NewHttpRequest(request *model.Request, logger *log.Logger) (resp *http.Response, startTime, endTime time.Time, err error) {
	method := request.Method
	url := request.URL
	body := request.GetBody()
	timeout := request.Timeout
	headers := request.Headers
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	var client *http.Client
	if request.KeepAlive {
		client = LangHttpClient
	} else {
		req.Close = true
		tr := &http.Transport{}
		client = &http.Client{
			Transport: tr,
			Timeout:   timeout,
		}
	}

	startTime = time.Now()
	resp, err = client.Do(req)
	endTime = time.Now()
	if err != nil {
		logger.Println("请求失败:", err)
		return
	}
	return
}

func SendHttpRequest(request *model.Request, logger *log.Logger) (bool, string, int64, time.Time, time.Time) {
	var (
		isSucceed     = false
		errCode       = "200"
		respBody      string
		contentLength = int64(0)
		err           error
		resp          *http.Response
		sTime         time.Time
		eTime         time.Time
		//requestTime   uint64
	)
	resp, sTime, eTime, err = NewHttpRequest(request, logger)

	if err != nil {
		errCode = fmt.Sprintf("%s", err.Error())
	} else {
		contentLength += resp.ContentLength
		errCode, isSucceed, respBody = request.GetVerifyHTTP()(request, resp)
		if !isSucceed {
			logger.Printf("请求校验失败: %s, response body:%s, response headers:%s \n", respBody, request.Body, request.Headers)
		}
	}
	//logger.Println(isSucceed, errCode, requestTime, contentLength)
	return isSucceed, errCode, contentLength, sTime, eTime
}

func PressHttp(chanID uint64, request *model.Request, ch chan<- *model.RequestResults, logger *log.Logger) {
	isSucceed, errCode, contentLength, startTime, endTime := SendHttpRequest(request, logger)
	requestResults := &model.RequestResults{
		RequestTime:   startTime,
		ResponseTime:  endTime,
		IsSucceed:     isSucceed,
		ErrCode:       errCode,
		ReceivedBytes: contentLength,
	}
	requestResults.SetID(chanID, chanID)
	ch <- requestResults
}
