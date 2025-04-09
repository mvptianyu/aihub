package aihub

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

var defaultHTTPClient = &http.Client{}

// HTTPCall 统一发送http请求方法入口
//
// Examples:
//
//	httpcli.Call( ctx, "http://www.xxx.com/", "GET", "aa=2", HTTPWithTimeOut(3) )
func HTTPCall(surl string, method string, req interface{}, reqHeader *http.Header, options ...HTTPOption) (rsp *http.Response, err error) {
	var httpReq *http.Request
	var httpResp *http.Response

	// 选项初始化
	opts := newHTTPOptions(reqHeader)
	for _, opt := range options {
		opt(opts)
	}

	var err1 error
	attempt := 0
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	defer func() {
		if httpResp != nil && httpResp.Body != nil {
			httpResp.Body.Close()
		}

		if err1 != nil {
			log.Printf("HTTPCall err => [surl]:%s, [err]:%v, [req]:%v\n", surl, err1, req)
		}
	}()

	_, err1 = url.Parse(surl)
	if err1 != nil {
		err = ErrHTTPRequestURLInvalid
		return
	}

	var bodyReq []byte
	if method == http.MethodPost || method == http.MethodPut {
		if reflect.TypeOf(req).Kind() == reflect.String {
			// 字符串的话不需要转json
			bodyReq = []byte(req.(string))
		} else {
			bodyReq, err1 = json.Marshal(req)
			if err1 != nil {
				err = ErrHTTPRequestBodyInvalid
				return
			}
		}
	}

	// 超时控制
	ctx, cancel := context.WithTimeout(context.Background(), opts.TimeOut)
	defer cancel()

	// 重试执行
	for attempt = 0; attempt <= opts.Retry; attempt++ {
		// 每次http.Client.Do后http.Request会失效，重试前需要new一个
		httpReq = httpCallCreateRequest(method, surl, bodyReq, &opts.Header)
		quitCh := make(chan bool, 1)

		go func() {
			bQuit := false
			defer func() {
				quitCh <- bQuit
			}()
			rsp, err1 = defaultHTTPClient.Do(httpReq)
			if rsp != nil {
				// 只重试 5xx and 429，其余正常处理和跳过
				if rsp.StatusCode <= http.StatusInternalServerError && rsp.StatusCode != http.StatusTooManyRequests {
					bQuit = true
					return
				}
			}
		}()

		// 超时判断
		select {
		case <-ctx.Done():
			if err1 == nil {
				err1 = ctx.Err()
			}
			err = ErrHTTPRequestTimeout
			return
		case bQuit := <-quitCh:
			if err1 != nil && err == nil {
				err = ErrUnknown
			}

			if bQuit {
				// 跳出不重试
				return
			}

			time.Sleep(opts.RetryWait)
		}
	}

	return
}

// 构建请求
func httpCallCreateRequest(method string, surl string, body []byte, headers *http.Header) *http.Request {
	var bodybuf = bytes.NewBuffer(body)
	httpReq, _ := http.NewRequest(
		method,
		surl,
		bodybuf,
	)

	if headers != nil {
		httpReq.Header = *headers
	}

	return httpReq
}

// --------------------------

const (
	defaultHTTPOptionsTimeOut   = 15 * time.Second
	defaultHTTPOptionsRetry     = 0
	defaultHTTPOptionsRetryWait = 30 * time.Millisecond
)

// HTTPOptions http请求选项设置
type HTTPOptions struct {
	Header    http.Header   // header头设定
	TimeOut   time.Duration // 超时设定
	Retry     int           // 重试次数设定
	RetryWait time.Duration // 重试间隔设定，退火策略
}

func newHTTPOptions(headers *http.Header) *HTTPOptions {
	opts := &HTTPOptions{
		Header:    make(http.Header),
		TimeOut:   defaultHTTPOptionsTimeOut,
		Retry:     defaultHTTPOptionsRetry,
		RetryWait: defaultHTTPOptionsRetryWait,
	}
	if headers != nil {
		opts.Header = *headers
	}
	return opts
}

type HTTPOption func(c *HTTPOptions)

func HTTPWithTimeOut(timeout int64) HTTPOption {
	return func(c *HTTPOptions) {
		if timeout <= 0 || timeout >= 60 {
			return
		}
		c.TimeOut = time.Duration(timeout) * time.Second
	}
}

func HTTPWithRetry(retry int) HTTPOption {
	return func(c *HTTPOptions) {
		c.Retry = retry
	}
}

func HTTPWithRetryWait(retrywait time.Duration) HTTPOption {
	return func(c *HTTPOptions) {
		c.RetryWait = retrywait
	}
}
