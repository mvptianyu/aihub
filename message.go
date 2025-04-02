/*
@Project: aihub
@Module: core
@File : message.go
*/
package aihub

import (
	"encoding/json"
)

type Message struct {
	Content      string                `json:"-"`
	MultiContent []*MessageContentPart `json:"-"`
	Role         MessageRoleType       `json:"role"`
	Name         string                `json:"name,omitempty"`
	ToolCallID   string                `json:"tool_call_id,omitempty"` // Role=tool发出请求时携带之前由Role=assistant返回的ToolCallID
	ToolCalls    []*MessageToolCall    `json:"tool_calls,omitempty"`   // Role=assistant返回的Message所带的ToolCalls
	Refusal      string                `json:"refusal,omitempty"`

	CreateTime int64 `json:"-"`
}

type messageSingle struct {
	Content    string             `json:"content,omitempty"`
	Role       MessageRoleType    `json:"role"`
	Name       string             `json:"name,omitempty"`
	ToolCallID string             `json:"tool_call_id,omitempty"` // Role=tool发出请求时携带之前由Role=assistant返回的ToolCallID
	ToolCalls  []*MessageToolCall `json:"tool_calls,omitempty"`   // Role=assistant返回的Message所带的ToolCalls
	Refusal    string             `json:"refusal,omitempty"`
}

type messageMulti struct {
	Role         MessageRoleType       `json:"role"`
	Name         string                `json:"name,omitempty"`
	ToolCallID   string                `json:"tool_call_id,omitempty"` // Role=tool发出请求时携带之前由Role=assistant返回的ToolCallID
	ToolCalls    []*MessageToolCall    `json:"tool_calls,omitempty"`   // Role=assistant返回的Message所带的ToolCalls
	Refusal      string                `json:"refusal,omitempty"`
	MultiContent []*MessageContentPart `json:"content,omitempty"`
}

func (m *Message) MarshalJSON() ([]byte, error) {
	if m.MultiContent != nil && len(m.MultiContent) > 0 {
		msg := messageMulti{
			MultiContent: m.MultiContent,
			Role:         m.Role,
			Name:         m.Name,
			ToolCallID:   m.ToolCallID,
			ToolCalls:    m.ToolCalls,
			Refusal:      m.Refusal,
		}
		return json.Marshal(msg)
	}

	msg := messageSingle{
		Content:    m.Content,
		Role:       m.Role,
		Name:       m.Name,
		ToolCallID: m.ToolCallID,
		ToolCalls:  m.ToolCalls,
		Refusal:    m.Refusal,
	}
	return json.Marshal(msg)
}

func (m *Message) UnmarshalJSON(bs []byte) error {
	msg2 := &messageMulti{}
	if err := json.Unmarshal(bs, &msg2); err == nil {
		m.Role = msg2.Role
		m.Name = msg2.Name
		m.ToolCallID = msg2.ToolCallID
		m.ToolCalls = msg2.ToolCalls
		m.Refusal = msg2.Refusal
		m.MultiContent = msg2.MultiContent

		if m.MultiContent != nil && len(m.MultiContent) > 0 {
			return nil
		}
	}

	msg1 := &messageSingle{}
	if err := json.Unmarshal(bs, &msg1); err == nil {
		m.Role = msg1.Role
		m.Name = msg1.Name
		m.ToolCallID = msg1.ToolCallID
		m.ToolCalls = msg1.ToolCalls
		m.Refusal = msg1.Refusal
		m.Content = msg1.Content
		return nil
	}

	return ErrMessageContentFieldsMisused
}

func (m *Message) Copy() *Message {
	tmp := &Message{}
	*tmp = *m
	return tmp
}

type MessageToolCall struct {
	Id       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type MessageRoleType string

const (
	MessageRoleUser      MessageRoleType = "user"
	MessageRoleAssistant MessageRoleType = "assistant"
	MessageRoleSystem    MessageRoleType = "system"
	MessageRoleTool      MessageRoleType = "tool"
)

type MessageContentType string

const (
	MessageContentTypeText  MessageContentType = "text"
	MessageContentTypeImage MessageContentType = "image_url"
	MessageContentTypeAudio MessageContentType = "input_audio"
	MessageContentTypeFile  MessageContentType = "file"
)

type MessageContentPart struct {
	Type       MessageContentType   `json:"type"`
	Text       string               `json:"text,omitempty"`
	ImageUrl   *MessageContentImage `json:"image_url,omitempty"`
	InputAudio *MessageContentAudio `json:"input_audio,omitempty"`
	File       *MessageContentFile  `json:"file,omitempty"`
}

type MessageContentImage struct {
	URL string `json:"url"` // 图片url或base64数据
}
type MessageContentAudioFormat string

const (
	MessageContentAudioFormatMP3 MessageContentAudioFormat = "mp3"
	MessageContentAudioFormatWAV MessageContentAudioFormat = "wav"
)

type MessageContentAudio struct {
	Data   string                    `json:"data"`
	Format MessageContentAudioFormat `json:"format"` // mp3|wav
}

type MessageContentFile struct {
	FileData string `json:"file_data"` // base64数据
	FileName string `json:"format"`    // 文件名
}
