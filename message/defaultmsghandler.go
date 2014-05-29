package message

import (
	"fmt"
	"sherlock/common"
)

type DefaultMsgHandler struct {
}

func NewDefaultMsgHandler() common.MsgHandlerIf {
	return &DefaultMsgHandler{}
}

func (self *DefaultMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	fmt.Println(ctnt)
	return nil
}
