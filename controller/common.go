package controller

import (
	"fmt"
	"github.com/kamiokk/gosession"
	memsess "github.com/kamiokk/gosession/mem"
	"github.com/gin-gonic/gin"
)

const (
	CODE_FAILED = 0
	CODE_SUCCEED = 1
	CODE_NEED_LOGIN = 10
	CODE_LOGIN_ERR = 11
	CODE_LOGINED = 12
	CODE_REGISTER_DUP_NAME = 20
	CODE_REGISTER_ERR = 21
)

const (
	MSG_SUCCEED = "succeed"
	MSG_ERROR = "error"
	MSG_NEED_LOGIN = "you need to login first"
	MSG_LOGIN_ERR = "account or password wrong"
	MSG_LOGINED = "already logined"
	MSG_REGISTER_DUP_NAME = "account already exist"
	MSG_REGISTER_ERR = "register failed"
)

func startSession(c *gin.Context) (*gosession.Session) {
	smodel := &memsess.Model{}
	session,err := gosession.Start(c.Request,c.Writer,smodel)
	if err != nil {
		panic("Can not start a new session.")
	}
	return session
}

func isLogined(s *gosession.Session) bool {
	uid,err := s.GetUInt("userID")
	if uid > 0 {
		return true
	}
	if err != nil {
		fmt.Printf("check login err : %v\n",err)
	}
	return false
}