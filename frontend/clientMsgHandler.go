package frontend

import (
	"fmt"
	"encoding/json"
	"sherlock/common"
)

type ClientMsgHandler struct {
	laddr string
	acqOk chan string //channel wait for acquire ok
}

func NewClientMsgHandler(laddr string, acqOk chan string) common.MsgHandlerIf {
	return &ClientMsgHandler{laddr:laddr, acqOk:acqOk}
}

func (self *ClientMsgHandler) Handle(ctnt common.Content, reply *common.Content) error  {
	// Examine the content
	switch ctnt.Head {
	case "acqOk":
		var lu common.LUpair
		json.Unmarshal([]byte(ctnt.Body), &lu)
		if lu.Username == self.laddr {
			fmt.Println("client msg handler", ctnt)
			self.acqOk <- "true"
			reply.Head = "success"
		}
	}
	return nil
}
