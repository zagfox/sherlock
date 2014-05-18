//Implementation of msg server func

package message

import (
	//"fmt"
	"sherlock/common"
)

var _ common.MessageIf = new(MsgListener)

type MsgListener struct {
	ch chan string
}

func NewMsgListener(ch chan string) *MsgListener {
	return &MsgListener{ch: ch}
}

func (self *MsgListener) Msg(msg string, succ *bool) error {
	self.ch<- msg
	*succ = true
	return nil
}
