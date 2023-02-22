package model

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type RequestResults struct {
	ID           string    // 消息ID
	ChanID       uint64    // 消息ID
	RequestTime  time.Time // 发出请求的时间
	ResponseTime time.Time // 收到响应的时间
	//Time          uint64 // 请求时间 纳秒
	IsSucceed     bool   // 是否请求成功
	ErrCode       string // 错误码
	ReceivedBytes int64
}

// SetID 设置请求唯一ID
func (r *RequestResults) SetID(chanID uint64, number uint64) {
	id := fmt.Sprintf("%d_%d", chanID, number)
	r.ID = id
	r.ChanID = chanID
}

func getData(response *http.Response) (body []byte, err error) {
	var reader io.ReadCloser
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		// gzip格式数据需要解压缩
		reader, err = gzip.NewReader(response.Body)
		defer func() {
			_ = reader.Close()
		}()
	default:
		reader = response.Body
	}
	// 读取response body数据
	body, err = ioutil.ReadAll(reader)
	if err != nil {
		if strings.Contains(err.Error(), "EOF") && len(body) != 0 {
			fmt.Printf("when read response: %s, will parse: ", err.Error())
			r_b := make([]byte, 0, len(body))
			_, err = reader.Read(r_b)
			body = r_b
			fmt.Println(body)
		}
	}
	response.Body = ioutil.NopCloser(bytes.NewReader(body))
	return
}
