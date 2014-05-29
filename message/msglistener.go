//Implementation of msg server func

package message

import (
	//"fmt"
	"sherlock/common"
)

var _ common.MessageIf = new(MsgListener)

type MsgListener struct {
	ch chan common.Content
	handler common.MsgHandlerIf
}

func NewMsgListener(ch chan common.Content, handler common.MsgHandlerIf) *MsgListener {
	return &MsgListener{ch: ch, handler:handler}
}

func (self *MsgListener) Msg(msg common.Content, reply *common.Content) error {
	//self.ch<- msg
	//reply.Head = "success"
	return self.handler.Handle(msg, reply)
	//return nil
}
