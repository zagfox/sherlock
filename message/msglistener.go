//Implementation of msg server func

package message

import (
	//"fmt"
	"sherlock/common"
)

var _ common.MessageIf = new(MsgListener)

type MsgListener struct {
	ch chan common.Content
}

func NewMsgListener(ch chan common.Content) *MsgListener {
	return &MsgListener{ch: ch}
}

func (self *MsgListener) Msg(msg common.Content, reply *common.Content) error {
	self.ch<- msg
	reply.Head = "success"
	return nil
}
