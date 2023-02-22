package client

import (
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"time"
)

var (
	LangHttpClient *http.Client
	once           sync.Once
)

func CreateLongHttpClient(maxConns int) {
	once.Do(func() {
		tr := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        0,                // 最大连接数,默认0无穷大
			MaxIdleConnsPerHost: maxConns,         // 对每个host维持的最大连接数量(MaxIdleConnsPerHost<=MaxIdleConns)
			IdleConnTimeout:     90 * time.Second, // 多长时间未使用自动关闭连接池
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
		LangHttpClient = &http.Client{
			Transport: tr,
		}
	})
}
