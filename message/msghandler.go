package message

import (
	"fmt"
	"sherlock/common"
)

var _ common.MessageIf = new(MsgHandler)

type MsgHandler struct {
	ch chan string
}

func NewMsgHandler(ch chan string) *MsgHandler {
	return &MsgHandler{ch: ch}
}

func (self *MsgHandler) Msg(msg string, succ *bool) error {
	fmt.Println("in msg handler")
	return nil
}
