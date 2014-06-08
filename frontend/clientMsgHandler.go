package frontend

import (
	//"fmt"
	"encoding/json"
	"sherlock/common"
)

type ClientMsgHandler struct {
	laddr string
	mAcqChan map[common.LUpair]chan string //channel wait for acquire ok
}

func NewClientMsgHandler(laddr string, mAcqChan map[common.LUpair]chan string) common.MsgHandlerIf {
	return &ClientMsgHandler{laddr:laddr, mAcqChan:mAcqChan}
}

func (self *ClientMsgHandler) Handle(ctnt common.Content, reply *common.Content) error  {
	// Examine the content
	switch ctnt.Head {
	case "LockAcquired":
		var lu common.LUpair
		json.Unmarshal([]byte(ctnt.Body), &lu)

		// the channel must be there when doing rpc call
		ch, ok := self.mAcqChan[lu]
		if !ok {
			reply.Head = "failure"
			return nil
		}
		ch <- "acquired"
		reply.Head = "success"
	case "LockLeaseCheck":
	}
	return nil
}
