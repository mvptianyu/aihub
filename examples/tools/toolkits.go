/*
@Project: aihub
@Module: main
@File : toolkits.go
*/
package tools

import (
	"context"
	"errors"
	"fmt"
	"github.com/mvptianyu/aihub"
	"io"
	"net/http"
	"strings"
)

type Toolkits struct {
}

func (d *Toolkits) GetWeather(ctx context.Context, input *aihub.ToolInputBase, output *aihub.Message) (err error) {
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

func (d *Toolkits) GetSong(ctx context.Context, input *SongReq, output *aihub.Message) (err error) {
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

func (d *Toolkits) QueryClickHouse(ctx context.Context, input *aihub.ToolInputBase, output *aihub.Message) (err error) {
	// curl --location 'https://clickhouse-k8s-sg-prod.data-infra.shopee.io/?max_result_rows=10000&max_execution_time=60' \
	// --header 'authorization: Basic c2hvcGVlX21tY19tbXMtY2x1c3Rlcl9tcHBfU2hvcGVlTU1DX2RhdGFTZXJ2aWNlX29ubGluZTpzaG9wZWVfbW1jX21tc18yMDIz' \
	// --header 'Content-Type: text/plain' \
	// --data 'SELECT grass_date, scene_id, SUM(play_cnt) AS total_play_cnt, AVG(play_succ_rate) AS avg_play_succ_rate FROM mmc_dp.mmc_mart_dws_vod_daily_play_metrics_1d_all WHERE grass_date >= toDate(now()) - 5 AND scene_id = '\''12401'\'' AND grass_region = '\''ID'\'' GROUP BY grass_date, scene_id ORDER BY grass_date ASC FORMAT JSON'

	surl := "https://clickhouse-k8s-sg-prod.data-infra.shopee.io/?max_result_rows=10000&max_execution_time=60"
	header := &http.Header{}
	header.Set("authorization", "Basic c2hvcGVlX21tY19tbXMtY2x1c3Rlcl9tcHBfU2hvcGVlTU1DX2RhdGFTZXJ2aWNlX29ubGluZTpzaG9wZWVfbW1jX21tc18yMDIz")
	header.Set("Content-Type", "text/plain")
	sql := strings.TrimRight(input.GetRawInput(), ";") + " FORMAT CSV"
	fmt.Println("====> sql: ", sql)

	rsp, err := aihub.HTTPCall(surl, http.MethodPost, sql, header)
	if err != nil {
		output.Content = err.Error()
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		output.Content = "Clickhouse查询失败"
		return errors.New(output.Content)
	}

	defer rsp.Body.Close()
	bs, err := io.ReadAll(rsp.Body)
	output.Content = string(bs)
	return err
}
