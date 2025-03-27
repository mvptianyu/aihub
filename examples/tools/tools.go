/*
@Project: aihub
@Module: main
@File : tools.go
*/
package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mvptianyu/aihub/core"
	"io"
	"net/http"
	"strings"
)

func Dispath(ctx context.Context, name string, in string) (out interface{}, err error) {
	switch name {
	case "GetWeather":
		return GetWeather(ctx, in)
	case "GetSong":
		return GetSong(ctx, in)
	case "QueryClickHouse":
		return QueryClickHouse(ctx, in)
	}
	return
}

func GetWeather(ctx context.Context, in string) (out interface{}, err error) {
	fmt.Println("===> GetWeather in: ", in)
	/*
		fmt.Println("即将调用工具：GetWeather，参数为：" + in + "，输入 'OK' 继续:")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		userInput := scanner.Text()

		if userInput != "OK" {
			return "用户取消了退出", errors.New("用户取消了退出")
		}
	*/

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

func QueryClickHouse(ctx context.Context, in string) (out interface{}, err error) {
	// curl --location 'https://clickhouse-k8s-sg-prod.data-infra.shopee.io/?max_result_rows=10000&max_execution_time=60' \
	// --header 'authorization: Basic c2hvcGVlX21tY19tbXMtY2x1c3Rlcl9tcHBfU2hvcGVlTU1DX2RhdGFTZXJ2aWNlX29ubGluZTpzaG9wZWVfbW1jX21tc18yMDIz' \
	// --header 'Content-Type: text/plain' \
	// --data 'SELECT grass_date, scene_id, SUM(play_cnt) AS total_play_cnt, AVG(play_succ_rate) AS avg_play_succ_rate FROM mmc_dp.mmc_mart_dws_vod_daily_play_metrics_1d_all WHERE grass_date >= toDate(now()) - 5 AND scene_id = '\''12401'\'' AND grass_region = '\''ID'\'' GROUP BY grass_date, scene_id ORDER BY grass_date ASC FORMAT JSON'

	surl := "https://clickhouse-k8s-sg-prod.data-infra.shopee.io/?max_result_rows=10000&max_execution_time=60"
	header := &http.Header{}
	header.Set("authorization", "Basic c2hvcGVlX21tY19tbXMtY2x1c3Rlcl9tcHBfU2hvcGVlTU1DX2RhdGFTZXJ2aWNlX29ubGluZTpzaG9wZWVfbW1jX21tc18yMDIz")
	header.Set("Content-Type", "text/plain")
	sql := strings.TrimRight(in, ";") + " FORMAT JSON"
	fmt.Println("====> sql: ", sql)

	rsp, err := core.HTTPCall(surl, http.MethodPost, sql, header)
	if err != nil {
		return err.Error(), err
	}
	if rsp.StatusCode != http.StatusOK {
		return "Clickhouse查询失败", errors.New("Clickhouse查询失败")
	}

	defer rsp.Body.Close()
	bs, _ := io.ReadAll(rsp.Body)
	return string(bs), nil
}
