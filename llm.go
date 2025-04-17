package aihub

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mvptianyu/aihub/ssestream"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"net/url"
)

const chatCompletionsAPI = "/chat/completions"

// LLM提供商
type llm struct {
	cfg *LLMConfig

	client  *http.Client
	limiter *rate.Limiter
}

func newLLM(cfg *LLMConfig) (ILLM, error) {
	if err := cfg.AutoFix(); err != nil {
		return nil, err
	}

	ins := &llm{
		cfg:    cfg,
		client: &http.Client{},
	}
	ins.limiter = rate.NewLimiter(rate.Limit(cfg.RateLimit), cfg.RateLimit)
	return ins, nil
}

func (p *llm) checkRateLimit() (err error) {
	if p.limiter != nil && !p.limiter.Allow() {
		return ErrProviderRateLimit
	}
	return nil
}

func (p *llm) GetBriefInfo() BriefInfo {
	return p.cfg.BriefInfo
}

func (p *llm) CreateChatCompletion(ctx context.Context, request *CreateChatCompletionReq) (response *CreateChatCompletionRsp, err error) {
	if request.Stream {
		request.Stream = false
	}
	if request.Model == "" {
		request.Model = p.cfg.Name
	}

	if err = p.checkRateLimit(); err != nil {
		return
	}

	surl, _ := url.JoinPath(p.cfg.BaseURL, p.cfg.Version, chatCompletionsAPI)
	headers := &http.Header{
		"Content-Type": {"application/json"},
	}
	if p.cfg.APIKey != "" {
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey))
	}

	rsp, err1 := HTTPCall(surl, http.MethodPost, request, headers, HTTPWithTimeOut(30))
	if err1 != nil {
		err = err1
		return
	}

	bs, _ := io.ReadAll(rsp.Body)
	defer rsp.Body.Close()
	tmp := &CreateChatCompletionRsp{}
	if err = json.Unmarshal(bs, tmp); err != nil {
		return
	}

	response = tmp
	return
}

func (p *llm) CreateChatCompletionStream(ctx context.Context, request *CreateChatCompletionReq) (stream *ssestream.StreamReader[CreateChatCompletionRsp]) {
	if request.Stream == false {
		request.Stream = true
	}
	if request.Model == "" {
		request.Model = p.cfg.Name
	}

	if err := p.checkRateLimit(); err != nil {
		return ssestream.NewStreamReader[CreateChatCompletionRsp](nil, err)
	}

	surl, _ := url.JoinPath(p.cfg.BaseURL, p.cfg.Version, chatCompletionsAPI)
	headers := &http.Header{
		"Content-Type": {"application/json"},
	}
	if p.cfg.APIKey != "" {
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", p.cfg.APIKey))
	}

	rsp, err := HTTPCall(surl, http.MethodPost, request, headers, HTTPWithTimeOut(60))
	if err != nil {
		return ssestream.NewStreamReader[CreateChatCompletionRsp](nil, err)
	}

	return ssestream.NewStreamReader[CreateChatCompletionRsp](ssestream.NewDecoder(rsp.Body), err)
}
