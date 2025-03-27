package main

import (
	"context"
	"fmt"
	"github.com/mvptianyu/aihub"
	"github.com/mvptianyu/aihub/core"
	"github.com/mvptianyu/aihub/examples/tools"
)

func main() {
	ctx := context.Background()
	wiki := `
表名称:'mmc_dp.mmc_mart_dws_vod_daily_play_metrics_1d_all'

字段列表:
'grass_date', 统计日期，Date 类型
'grass_region', 统计区域，字符串
'scene_id', 场景 ID, 字符串
'cdn', 内容分发网络，字符串
'dau', 日活跃用户数，Int64 类型
'pcu', 最高同时在线用户数，Int64 类型
'qps', 每秒查询率，Int64 类型
'play_cnt', 播放次数，Int64 类型
'play_succ_rate', 播放成功率，Float64 类型
'v270p_play_cnt',270p 分辨率播放次数，Int64 类型
'v360p_play_cnt',360p 分辨率播放次数，Int64 类型
'v480p_play_cnt',480p 分辨率播放次数，Int64 类型
'v540p_play_cnt',540p 分辨率播放次数，Int64 类型
'v640p_play_cnt',640p 分辨率播放次数，Int64 类型
'v720p_play_cnt',720p 分辨率播放次数，Int64 类型
'v1080p_play_cnt',1080p 分辨率播放次数，Int64 类型
'sdk_error_rate',SDK 错误率，Float64 类型
'first_frame_event_cnt', 首帧事件次数，Int64 类型
'avg_first_frame_cost', 平均首帧耗时，Float64 类型
'p95_first_frame_cost',95% 分位首帧耗时，Float64 类型
'rebuff_rate', 卡顿率，Float64 类型
'rebuff_event_cnt', 卡顿事件次数，Int64 类型
'avg_rebuff_duration', 平均卡顿时长，Float64 类型
'p95_rebuff_duration',95% 分位卡顿时长，Float64 类型
'end_play_event_cnt', 播放结束事件次数，Int64 类型
'avg_play_duration', 平均播放时长，Float64 类型
'p95_play_duration',95% 分位播放时长，Float64 类型
'avg_play_bitrate', 平均播放比特率，Float64 类型
'p95_play_bitrate',95% 分位播放比特率，Float64 类型
'h264_play_cnt',H264 编码播放次数，Int64 类型
'h265_play_cnt',H265 编码播放次数，Int64 类型
'avg_first_loading_cost', 平均首次加载耗时，Float64 类型
'p95_first_loading_cost',95% 分位首次加载耗时，Float64 类型
'avg_first_request_cost', 平均首次请求耗时，Float64 类型
'p95_first_request_cost',95% 分位首次请求耗时，Float64 类型
'cancel_play_rate', 取消播放率，Float64 类型
's0_first_frame_rate',S0 首帧率，Float64 类型
'soft_decode_rate', 软解码率，Float64 类型
'os', 操作系统，字符串
'client_version', 客户端版本，字符串
'ori_play_cnt', 原始播放次数，Int64 类型
'avg_startplay_cost', 平均开始播放耗时，Float64 类型
'have_cache_rate', 有缓存率，Float64 类型
'back_source_rate', 回源率，Float64 类型
'avg_download_file_size', 平均下载文件大小，Float64 类型
'avg_process_cpu_payload', 平均进程 CPU 负载，Float64 类型
'avg_process_memory', 平均进程内存，Float64 类型
'avg_cpu_temperature', 平均 CPU 温度，Float64 类型
'avg_play_file_cost', 平均播放文件成本，Float64 类型
'avg_play_cost', 平均播放成本，Float64 类型
'avg_vv_cnt', 平均视频播放量，Int64 类型
's100_rebuff_duration',100% 分位卡顿时长，Int64 类型
'avg_vv_cnt_double', 平均视频播放量（双精度）,Float64 类型
`

	// Create a new agent
	myAgent := aihub.NewAgentWithYamlFile("sql.yaml", tools.Dispath)

	_, txt, err := myAgent.Run(
		ctx,
		"查询播放核心归档最近5天12401场景在id地区的总播放量和平均播放成功率，按日期、场景、地区分组和升序排序",
		core.WithClaim("本结果由MMS AI Agent自动生成"),
		core.WithDebug(true),
		core.WithContext(wiki),
	)
	fmt.Println(err)
	fmt.Println("=======================")
	fmt.Println(txt)

	seatalkGroup := "LMWNqAYCQVGLGi2fGYfvHw"

	core.SendSeatalkText(seatalkGroup, core.SeaTalkText{
		Content: txt,
	})
}
