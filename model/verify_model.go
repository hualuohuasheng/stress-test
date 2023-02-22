package model

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
)

// 校验函数
var (
	// verifyMapHTTP http 校验函数
	verifyMapHTTP = make(map[string]VerifyHTTP)
	// verifyMapHTTPMutex http 并发锁
	verifyMapHTTPMutex sync.RWMutex
)

type VerifyHTTP func(request *Request, response *http.Response) (verifyCode string, isSucceed bool, respBody string)

func RegisterVerifyHTTP(verify string, verifyFunc VerifyHTTP) {
	verifyMapHTTPMutex.Lock()
	// 注册完成后解锁
	defer verifyMapHTTPMutex.Unlock()
	key := fmt.Sprintf("http.%s", verify)
	verifyMapHTTP[key] = verifyFunc
}

type ResponseJSON struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func VerifyHttpJson(request *Request, response *http.Response) (verifyCode string, isSucceed bool, respBody string) {
	defer func() {
		_ = response.Body.Close()
	}()
	if response.StatusCode == http.StatusOK {
		body, err := getData(response)
		respBody = string(body)
		if err != nil {
			verifyCode = strconv.Itoa(response.StatusCode)
			fmt.Printf("请求结果 ioutil.ReadAll err:%v\n", err)
		} else {
			responseJSON := &ResponseJSON{}
			err = json.Unmarshal(body, responseJSON)
			if err != nil {
				verifyCode = strconv.Itoa(response.StatusCode)
				fmt.Printf("请求结果 json.Unmarshal err:%v\n", err)
			} else {
				// body 中code返回0为返回数据成功
				verifyCode = responseJSON.Code
				if verifyCode == "0" {
					isSucceed = true
				}
			}
		}
		// 开启调试模式
		if request.Debug {
			fmt.Printf("请求结果 httpCode:%d verifyCode:%s response body:%s err:%v \n", response.StatusCode, verifyCode, respBody, err)
		}
	}
	_, err := io.Copy(ioutil.Discard, response.Body)
	if err != nil {
		fmt.Printf("io copy err:%v", err)
	}
	return
}

func VerifyHttpStatusCode(request *Request, response *http.Response) (verifyCode string, isSucceed bool, respBody string) {
	defer func() {
		_ = response.Body.Close()
	}()
	code := strconv.Itoa(response.StatusCode)
	if code == request.Verify {
		isSucceed = true
	}
	// 开启调试模式
	if request.Debug {
		body, err := getData(response)
		respBody = string(body)
		fmt.Printf("请求结果 httpCode:%d body:%s err:%v \n", response.StatusCode, respBody, err)
	}
	// 必须将http.Response的Body读取完毕并且关闭后，才会重用底层的TCP连接
	// https://gocn.vip/topics/10626
	_, err := io.Copy(ioutil.Discard, response.Body)
	if err != nil {
		fmt.Printf("io copy err:%v", err)
	}
	return
}
