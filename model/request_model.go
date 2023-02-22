package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hualuohuasheng/stress-test/tools"
	"github.com/spf13/viper"
	"io"
	"strings"
	"time"
)

type Request struct {
	URL        string
	Method     string
	Headers    map[string]string
	Body       string
	Verify     string
	Timeout    time.Duration
	Debug      bool
	KeepAlive  bool
	VerifyCode string
	AddData    map[string]string
}

func (r *Request) GetHeaders() map[string]string {
	return r.Headers
}

func (r *Request) GetBody() io.Reader {
	return strings.NewReader(r.Body)
}

func (r *Request) getVerifyKey() (key string) {
	return fmt.Sprintf("http.%s", r.Verify)
}

func (r *Request) GetVerifyHTTP() VerifyHTTP {
	// 获取数据校验的方法
	verify, ok := verifyMapHTTP[r.getVerifyKey()]
	if !ok {
		panic("GetVerifyHTTP 验证方法不存在:" + r.Verify)
	}
	return verify
}

func (r *Request) Print() {
	if r == nil {
		return
	}
	result := fmt.Sprintf("request:\n url:%s \n method:%s \n headers:%v \n", r.URL, r.Method, r.Headers)
	result = fmt.Sprintf("%s data:%v \n", result, r.Body)
	result = fmt.Sprintf("%s verify:%s \n timeout:%s \n debug:%v \n", result, r.Verify, r.Timeout, r.Debug)
	//result = fmt.Sprintf("%s http2.0：%v \n keepalive：%v \n maxCon:%v ", result, r.HTTP2, r.Keepalive, r.MaxCon)
	fmt.Println(result)
	return
}

func getHeaderValue(v string, headers map[string]string) {
	index := strings.Index(v, ":")
	if index < 0 {
		return
	}
	vIndex := index + 1
	if len(v) >= vIndex {
		value := strings.TrimPrefix(v[vIndex:], " ")
		if _, ok := headers[v[:index]]; ok {
			headers[v[:index]] = fmt.Sprintf("%s; %s", headers[v[:index]], value)
		} else {
			headers[v[:index]] = value
		}
	}
}

func GetHeaders(reqHeaders []string, headers map[string]string) {
	for _, v := range reqHeaders {
		getHeaderValue(v, headers)
	}
}

func NewRequest(url, method, verify, code string, timeout time.Duration, debug bool, KeepAlive bool, headers map[string]string, reqBody string) (request *Request, err error) {
	var body string
	if reqBody != "" {
		method = "POST"
		body = reqBody
	}

	if _, ok := headers["Content-Type"]; !ok {
		headers["Content-Type"] = "application/json; charset=utf-8"
	}
	form := ""
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		form = "http"
	}
	if form == "" {
		err = fmt.Errorf("url:%s 不合法,必须是完整http、webSocket连接", url)
		return
	}
	var ok bool
	if verify == "" {
		verify = "statusCode"
	}
	key := fmt.Sprintf("%s.%s", form, verify)
	_, ok = verifyMapHTTP[key]
	if !ok {
		err = errors.New("验证器不存在:" + key)
		return
	}
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	request = &Request{
		URL:        url,
		Method:     strings.ToUpper(method),
		Headers:    headers,
		Body:       body,
		Verify:     verify,
		Timeout:    timeout,
		Debug:      debug,
		KeepAlive:  KeepAlive,
		VerifyCode: code,
	}
	return
}

// 生成RTS下单的body数据
func NewRTSSubmitOrderBody(headerTime string, cfg *viper.Viper, debug bool) (body string) {
	receiveInfo := make(map[string]string, 0)
	receiveInfo["recipientCountry"] = "GB"
	receiveInfo["receiverFirstName"] = "Li XX"
	receiveInfo["routingNumber"] = "111000025"
	receiveInfo["accountNumber"] = "000123123123"
	receiveInfo["accountType"] = "individual"
	receiveInfo["receiverCurrency"] = "USD"
	receiveInfo["payOutMethod"] = "BANK"

	instructionId := "stress_test_" + headerTime
	// 定义rts下单的原始请求参数
	params := make(map[string]interface{})
	params["instructionId"] = instructionId
	params["paymentId"] = instructionId
	params["originalId"] = ""
	params["sendNodeCode"] = "fape1meh4bsz"
	//params["sendCountry"] = "US"
	params["sendCurrency"] = "USD"
	params["sendAmount"] = "30"
	params["receiveNodeCode"] = "fape1meh4bsz"
	//params["receiveCountry"] = "GB"
	params["receiveCurrency"] = "USD"
	params["receiveInfo"] = receiveInfo
	oBody, _ := json.Marshal(params)
	bodyStr := string(oBody)
	nonce := tools.RandStrByLetters(16)
	addData := tools.RandStrByLetters(16)
	secretKey := cfg.GetString("SecretKey")
	enData := tools.AESEncryptData(secretKey, bodyStr, nonce, addData)
	body = "{\"resource\":{\"algorithm\":\"AES_256_GCM\",\"ciphertext\":\"" + enData + "\",\"associatedData\":\"" + addData + "\",\"nonce\":\"" + nonce + "\"}}"
	//body = bodyStr
	if debug {
		fmt.Printf("未加密前的body: %s\n,nonce: %s\n,addData: %s\n,secretKey: %s\n", bodyStr, nonce, addData, secretKey)
		fmt.Printf("得到新的请求body: %s\n", body)
	}
	return
}

func NewRTSHeaders(headerTime, enData string, cfg *viper.Viper, headers map[string]string, debug bool) {
	keyFile := cfg.GetString("RsaKeyFile")
	priKey := tools.ReadKeyFromPermFile(keyFile)
	plainText := fmt.Sprintf("%s::%s", headerTime, enData)
	sign := tools.RsaEncryptUsePrivateKey(priKey, plainText)
	headers["sign"] = sign
	headers["timestamp"] = headerTime
	if debug {
		fmt.Printf("待签名的数据: %s\n, 签名后的数据: %s\n", plainText, sign)
	}
}
