package xhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// 请求编码方法
type EncodeReqFn func(reqMsg interface{}) (io.Reader, error)

// 响应解码方法
type DecodeRespFn func(reader io.Reader, respMsg interface{}) error

// EncodeJsonReq 编码指定的json格式请求消息reqMsg
func EncodeJsonReq(reqMsg interface{}) (io.Reader, error) {
	reqData, err := json.Marshal(reqMsg)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(reqData), nil
}

// DecodeJsonResp 解码指定的reader到json格式响应消息respMsg
func DecodeJsonResp(reader io.Reader, respMsg interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	_, err := io.Copy(buf, reader)
	if err != nil {
		return errors.New("io copy error: " + err.Error())
	}
	return json.Unmarshal(buf.Bytes(), respMsg)
}

// DecodeJsonResp 解码指定的reader到*string格式响应消息respMsg
func DecodeStrResp(reader io.Reader, respMsg interface{}) error {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	_, err := io.Copy(buf, reader)
	if err != nil {
		return errors.New("io copy error: " + err.Error())
	}
	tmp, ok := respMsg.(*string)
	if !ok {
		return errors.New("respMsg is not string pointer")
	}
	*tmp = string(buf.Bytes())
	return nil
}
