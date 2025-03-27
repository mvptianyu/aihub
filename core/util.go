/*
@Project: aihub
@Module: core
@File : util.go
*/
package core

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

const seatalkHookUrl = "https://openapi.seatalk.io/webhook/group/%s"

type SeaTalkMsg struct {
	Tag    string      `json:"tag"` // 默认text
	Detail SeaTalkText `json:"text"`
}

type SeaTalkText struct {
	Content            string   `json:"content"`              // 文本内容
	MentionedList      []string `json:"mentioned_list"`       // 可选，提醒的人员id列表，例如：["111111"]
	MentionedEmailList []string `json:"mentioned_email_list"` // 可选，提醒的email列表，例如：["xxxx@shopee.com"]
	AtAll              bool     `json:"at_all"`               // 可选，是否群组@all,默认false
}

type SeaTalkImage struct {
	Tag         string `json:"tag"`
	ImageBase64 struct {
		Content string `json:"content"`
	} `json:"image_base64"`
}

// 发送文本消息
func SendSeatalkText(group string, txt SeaTalkText) error {
	// MarkDown替换
	txt.Content = strings.ReplaceAll(txt.Content, "```sql", "```")
	txt.Content = strings.ReplaceAll(txt.Content, "```json", "```")
	txt.Content = strings.ReplaceAll(txt.Content, "```xml", "```")

	url := fmt.Sprintf(seatalkHookUrl, group)
	msgData := SeaTalkMsg{
		Tag:    "text",
		Detail: txt,
	}
	byData, err := json.Marshal(msgData)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(byData))
	return err
}

// 发送图片消息
func SendSeatalkImage(group string, buf []byte) error {
	url := fmt.Sprintf(seatalkHookUrl, group)
	msgData := SeaTalkImage{
		Tag: "image",
	}
	msgData.ImageBase64.Content = base64.StdEncoding.EncodeToString(buf)

	byData, err := json.Marshal(msgData)
	if err != nil {
		return err
	}

	_, err = http.Post(url, "application/json", bytes.NewBuffer(byData))
	return err
}

// 定义用于匹配 Markdown 语法的正则表达式
var markdownRegexps = []*regexp.Regexp{
	regexp.MustCompile(`^#{1,6}\s`),      // 标题（# 到 ######）
	regexp.MustCompile(`^- \[ \] `),      // 任务列表
	regexp.MustCompile(`^- `),            // 无序列表
	regexp.MustCompile(`^\d+\. `),        // 有序列表
	regexp.MustCompile("```"),            // 代码块
	regexp.MustCompile(`!\[.*\]\(.*\)`),  // 图片
	regexp.MustCompile(`\[(.*)\]\(.*\)`), // 链接
	regexp.MustCompile(`^\|\s.*\s\|`),    // 表格
	regexp.MustCompile(`\*\*.*\*\*`),     // 加粗
	regexp.MustCompile(`\*.*\*`),         // 斜体
	regexp.MustCompile(`~~.*~~`),         // 删除线
}

// HasMarkdownSyntax 函数用于检查输入的字符串是否包含 Markdown 语法
func HasMarkdownSyntax(s string) bool {
	// 遍历每个正则表达式模式
	for _, re := range markdownRegexps {
		// 检查输入字符串是否匹配当前模式
		if re.MatchString(s) {
			return true
		}
	}
	return false
}
