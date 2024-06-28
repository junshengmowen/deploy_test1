/*
Copyright (year) Bytedance Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package service

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Request struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type Response struct {
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers,omitempty"`
}

var client = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     false,
		MaxIdleConns:          2000,
		MaxIdleConnsPerHost:   1000,
		MaxConnsPerHost:       2000,
		IdleConnTimeout:       10 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

func RunOpenApi(ctx *gin.Context) {
	req := &Request{}
	resp := &Response{
		Headers: make(map[string]string),
	}

	err := ctx.BindJSON(req)
	if err != nil {
		resp.Body = fmt.Sprintf("request错误，err: %+v", err)
		fmt.Println(resp.Body)
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	httpReq, err := http.NewRequest(strings.ToUpper(req.Method), req.URL, strings.NewReader(req.Body))
	if err != nil {
		resp.Body = fmt.Sprintf("创建http request失败，err: %+v", err)
		fmt.Println(resp.Body)
		ctx.JSON(http.StatusBadRequest, resp)
		return
	}
	for k, v := range req.Headers {
		httpReq.Header.Add(k, v)
	}

	httpResp, err := client.Do(httpReq)
	if err != nil {
		resp.Body = fmt.Sprintf("请求失败，err: %+v", err)
		fmt.Println(resp.Body)
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}

	defer httpResp.Body.Close()
	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		resp.Body = fmt.Sprintf("请求Body解析失败，err: %+v", err)
		fmt.Println(resp.Body)
		ctx.JSON(http.StatusInternalServerError, resp)
		return
	}
	resp.Body = string(body)
	for k, v := range httpResp.Header {
		resp.Headers[k] = strings.Join(v, ",")
	}
	ctx.JSON(http.StatusOK, resp)
	return
}
