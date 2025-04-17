/*
@Project: mvptianyu
@Module: aihub
@File : response.go
*/
package aihub

import (
	"encoding/json"
	"errors"
)

type Response struct {
	rawResponse
	Err error `json:"-"`
}

type rawResponse struct {
	Message *Message `json:"message,omitempty"`
	Session *Session `json:"session,omitempty"`
	Content string   `json:"content"`
	Error   string   `json:"error,omitempty"`
}

func (r *Response) MarshalJSON() ([]byte, error) {
	if r.Err != nil {
		r.Error = r.Err.Error()
	}
	return json.Marshal(r.rawResponse)
}

func (r *Response) UnmarshalJSON(bs []byte) error {
	err1 := json.Unmarshal(bs, &r.rawResponse)
	if err1 != nil {
		return err1
	}

	if r.rawResponse.Error != "" {
		r.Err = errors.New(r.rawResponse.Error)
	}

	return nil
}
