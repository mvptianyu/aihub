/*
@Project: aihub
@Module: core
@File : util.go
*/
package aihub

import (
	"regexp"
)

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
