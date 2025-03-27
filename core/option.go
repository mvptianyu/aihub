/*
@Project: aihub
@Module: core
@File : option.go
*/
package core

type RunOptions struct {
	StopWords string // 结束退出词
	Debug     bool   // debug标志，开启则输出具体工具调用过程信息
	Claim     string // 宣称文案，例如：本次返回由xxx提供
	Context   string // 上下文提示词相关，例如在systemprompt中插入/替换该内容
}

type RunOptionFunc func(*RunOptions)

func WithStopWords(StopWords string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.StopWords = StopWords
	}
}

func WithDebug(Debug bool) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Debug = Debug
	}
}

func WithClaim(Claim string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Claim = Claim
	}
}

func WithContext(Context string) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Context = Context
	}
}
