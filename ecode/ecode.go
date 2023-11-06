package ecode

import "fmt"

var globalCMs = make(map[ECode]string)

// 注册全局通用错误码/错误信息集合
// 一般在进程初始化时调用
func RegisterCMs(cms map[ECode]string) {
	globalCMs = cms
}

type ECode uint32

// Msg 返回错误码对应的错误信息，匹配不到时返回"-"
func (e ECode) Msg() string {
	msg, ok := globalCMs[e]
	if !ok {
		return "-"
	}
	return msg
}

// Code 返回错误码
func (e ECode) Code() uint32 {
	return uint32(e)
}

type Err struct {
	// 错误码
	code uint32

	// 错误信息
	msg string

	// 造成错误的内在原因，不关注时可为nil
	cause error
}

// New 根据给定错误码eCode构建Err
func New(eCode ECode) *Err {
	return NewEM(eCode.Code(), eCode.Msg())
}

// NewEM 根据给定错误码code和错误信息msg构建Err
func NewEM(code uint32, msg string) *Err {
	err := &Err{
		code: code,
		msg:  msg,
	}
	return err
}

// NewEFArgs 根据给定错误码code和错误信息（format+args格式化后信息）构建Err
func NewEFArgs(code uint32, format string, args ...interface{}) *Err {
	return NewEM(code, fmt.Sprintf(format, args...))
}

// WithCause 设置造成错误的内在原因cause
func (e *Err) WithCause(cause error) *Err {
	e.cause = cause
	return e
}

// Code 返回错误码
func (e *Err) Code() uint32 {
	return e.code
}

// Msg 返回错误信息
func (e *Err) Msg() string {
	return e.msg
}

func (e *Err) Error() string {
	if e.cause == nil {
		return fmt.Sprintf("code: %v, msg: %v", e.code, e.msg)
	}
	return fmt.Sprintf("code: %v, msg: %v, cause: %v", e.code, e.msg, e.cause)
}
