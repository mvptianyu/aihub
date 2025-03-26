/*
@Project: aihub
@Module: main
@File : tools.go
*/
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

func Dispath(ctx context.Context, name string, in string) (out interface{}, err error) {
	switch name {
	case "GetWeather":
		return GetWeather(ctx, in)
	case "GetSong":
		return GetSong(ctx, in)
	}
	return
}

func GetWeather(ctx context.Context, in string) (out interface{}, err error) {
	fmt.Println("即将调用工具：GetWeather，参数为：" + in + "，输入 'OK' 继续:")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	userInput := scanner.Text()

	if userInput != "OK" {
		return "用户取消了退出", errors.New("用户取消了退出")
	}

	switch in {
	case "深圳":
		return "30度,天晴", nil
	case "香港":
		return "20度,多云", nil
	case "北京":
		return "7度,暴雨", nil
	default:
		return "40度,暴晒", nil
	}
}

type SongReq struct {
	Temperature int `json:"temperature"`
}

func GetSong(ctx context.Context, in string) (out interface{}, err error) {
	fmt.Println("===> GetSong in: ", in)
	req := &SongReq{}
	json.Unmarshal([]byte(in), req)

	if req.Temperature <= 10 {
		return "雨爱", nil
	} else if req.Temperature <= 20 {
		return "云层记忆", nil
	} else if req.Temperature <= 30 {
		return "晴天", nil
	}

	return "晒死了，还听啥歌", nil
}
