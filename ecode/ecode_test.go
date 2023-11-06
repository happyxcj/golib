package ecode

import (
	"errors"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	_ECodeServerErr = 3000
	// 参数错误
	_ECodeParamErr = 4000
)

var _cms = map[ECode]string{
	_ECodeParamErr:  "param error",
	_ECodeServerErr: "server error, try again later",
}

func TestECode(t *testing.T) {
	Convey("TestECode", t, func() {
		RegisterCMs(_cms)
		eCode := ECode(_ECodeParamErr)
		So(_ECodeParamErr, ShouldEqual, eCode.Code())
		So(_cms[_ECodeParamErr], ShouldEqual, eCode.Msg())

		eCode = ECode(1)
		So(1, ShouldEqual, eCode.Code())
		So("-", ShouldEqual, eCode.Msg())
	})
}

func TestErr(t *testing.T) {
	Convey("TestErr", t, func() {
		RegisterCMs(_cms)
		err := New(_ECodeParamErr)
		So(_ECodeParamErr, ShouldEqual, err.Code())
		So(_cms[_ECodeParamErr], ShouldEqual, err.Msg())
		So(fmt.Sprintf("code: %v, msg: %v", _ECodeParamErr, err.Msg()), ShouldEqual, err.Error())

		err = NewEM(_ECodeParamErr, "param 'name' is illegal")
		So(_ECodeParamErr, ShouldEqual, err.Code())
		So("param 'name' is illegal", ShouldEqual, err.Msg())
		So(fmt.Sprintf("code: %v, msg: %v", _ECodeParamErr, "param 'name' is illegal"), ShouldEqual, err.Error())

		err = NewEFArgs(_ECodeParamErr, "param 'name' length exceeds the maximum value '%d'", 100)
		So(_ECodeParamErr, ShouldEqual, err.Code())
		So("param 'name' length exceeds the maximum value '100'", ShouldEqual, err.Msg())
		So(fmt.Sprintf("code: %v, msg: %v", _ECodeParamErr, "param 'name' length exceeds the maximum value '100'"), ShouldEqual, err.Error())

		err = New(_ECodeServerErr).WithCause(errors.New("mysql error"))
		So(_ECodeServerErr, ShouldEqual, err.Code())
		So(_cms[_ECodeServerErr], ShouldEqual, err.Msg())
		So(fmt.Sprintf("code: %v, msg: %v, cause: %v", _ECodeServerErr, _cms[_ECodeServerErr], "mysql error"), ShouldEqual, err.Error())
	})
}
