package backend

import (
	"sherlock/common"
)

var _ common.MsgHandlerIf = new(PaxosMsgHandler)

type PaxosMsgHandler struct {
}

func NewPaxosMsgHandler() common.MsgHandlerIf {
    return &PaxosMsgHandler{}
	}

func (self *PaxosMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	return nil
}
