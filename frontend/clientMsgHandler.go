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
		/*if lu.Username == self.laddr {
			fmt.Println("client msg handler", ctnt)
			self.acqOk <- "true"
			reply.Head = "success"
		}*/
		ch, ok := self.mAcqChan[lu]
		if !ok {
			reply.Head = "failure"
			reply.Body = "lu pair not waiting"
		} else {
			ch <- "acquired"
			reply.Head = "success"
		}
	case "LockLeaseCheck":
	}
	return nil
}
