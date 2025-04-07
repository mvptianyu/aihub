/*
@Project: aihub
@Module: main
@File : toolkits.go
*/
package tools

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
)

func GetWeather(ctx context.Context, input *aihub.ToolInputBase, output *aihub.Message) (err error) {
	fmt.Printf("===> GetWeather input: %v\n", input)
	/*
		fmt.Println("即将调用工具：GetWeather，参数为：" + input.GetRawInput() + "，输入 'OK' 继续:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		userInput := scanner.Text()

		if userInput != "OK" {
			return "用户取消了退出", errors.New("用户取消了退出")
		}
	*/

	switch input.GetRawInput() {
	case "深圳":
		output.Content = "30度,天晴"
		return
	case "香港":
		output.Content = "20度,多云"
		return
	case "北京":
		output.Content = "7度,暴雨"
		return
	default:
		output.Content = "40度,暴晒"
		return
	}
}

type SongReq struct {
	aihub.ToolInputBase
	Temperature int `json:"temperature"`
}

func GetSong(ctx context.Context, input *SongReq, output *aihub.Message) (err error) {
	fmt.Printf("===> GetSong input: %v\n", input)

	if input.Temperature <= 10 {
		output.Content = "雨爱"
		return
	} else if input.Temperature <= 20 {
		output.Content = "云层记忆"
		return
	} else if input.Temperature <= 30 {
		output.Content = "晴天"
		return
	}

	output.Content = "晒死了，还听啥歌"
	return
}
