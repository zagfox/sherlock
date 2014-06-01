package backend

import (
	"fmt"
	"sherlock/common"
	"sherlock/paxos"
)

var _ common.MsgHandlerIf = new(ServerMsgHandler)

type ServerMsgHandler struct {
	handle2pc   common.MsgHandlerIf
	handlePaxos common.MsgHandlerIf
}

func NewServerMsgHandler(srvView *paxos.ServerView) common.MsgHandlerIf {
	paxosHandler := paxos.NewPaxosMsgHandler(srvView)
	return &ServerMsgHandler{handlePaxos: paxosHandler}
}

func (self *ServerMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	// Examine the content
	fmt.Println("in serverMsgHandler", ctnt)
	switch ctnt.Head {
	case "2pc":
		return self.handle2pc.Handle(ctnt, reply)
	case "paxos":
		return self.handlePaxos.Handle(ctnt, reply)
	}
	return nil
}
