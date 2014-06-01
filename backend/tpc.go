package backend

import (
	"sherlock/common"
)

var _ common.MsgHandlerIf = new(TpcMsgHandler)

type TpcMsgHandler struct {
}

func NewTpcMsgHandler() common.MsgHandlerIf {
	return &TpcMsgHandler{}
}

func (self *TpcMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	return nil
}
