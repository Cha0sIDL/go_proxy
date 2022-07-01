package main

import (
	"bytes"
	"context"
	"github.com/aliyun/fc-runtime-go-sdk/fc"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func HandleHttpRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	cli := &http.Client{}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "text/plain")
		w.Write([]byte(err.Error()))
		return nil
	}
	// 转发的URL
	reqURL := req.Header.Get("proxy")
	reqProxy, err := http.NewRequest(req.Method, reqURL, strings.NewReader(string(body)))
	if err != nil {
		log.Println("创建转发请求发生错误")
		// 响应状态码
		w.WriteHeader(http.StatusServiceUnavailable)
		return nil
	}
	// 转发请求的 Header
	for k, v := range req.Header {
		reqProxy.Header.Set(k, v[0])
	}
	// 发起请求
	responseProxy, err := cli.Do(reqProxy)
	defer responseProxy.Body.Close()
	// 转发响应的 Header
	for k, v := range responseProxy.Header {
		w.Header().Set(k, v[0])
	}
	// 转发响应的Body数据
	var data []byte
	// 读取转发响应的Body
	data, err = ioutil.ReadAll(responseProxy.Body)
	// 转发响应的输出数据
	//var dataOutput []byte
	//dataOutput = data
	// response的Body不能多次读取，
	// 上面已经被读取过一次，需要重新生成可读取的Body数据。
	resProxyBody := ioutil.NopCloser(bytes.NewBuffer(data))
	defer resProxyBody.Close() // 延时关闭
	// 响应状态码
	w.WriteHeader(responseProxy.StatusCode)
	// 复制转发的响应Body到响应Body
	io.Copy(w, resProxyBody)
	return nil
}

func main() {
	fc.StartHttp(HandleHttpRequest)
}
