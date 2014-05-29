//Implementation of msg server func

package message

import (
	//"fmt"
	"sherlock/common"
)

var _ common.MessageIf = new(MsgListener)

type MsgListener struct {
	handler common.MsgHandlerIf
}

func NewMsgListener(handler common.MsgHandlerIf) *MsgListener {
	return &MsgListener{handler:handler}
}

func (self *MsgListener) Msg(msg common.Content, reply *common.Content) error {
	//self.ch<- msg
	//reply.Head = "success"
	return self.handler.Handle(msg, reply)
	//return nil
}
