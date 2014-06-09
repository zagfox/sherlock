package backend

import (
//	"fmt"
	"sherlock/common"
	"sherlock/lockstore"
	"sherlock/paxos"
)

var _ common.MsgHandlerIf = new(ServerMsgHandler)

type ServerMsgHandler struct {
	handle2pc       common.MsgHandlerIf
	handlePaxos     common.MsgHandlerIf
	handleTransfer  common.MsgHandlerIf
}

func NewServerMsgHandler(srvView *paxos.ServerView, ds common.DataStoreIf, lg *lockstore.LogPlayer) common.MsgHandlerIf {
	paxosHandler := paxos.NewPaxosMsgHandler(srvView, lg)
	tpcHandler := NewTpcMsgHandler(lg, srvView)
	handleTransfer := NewTransferMsgHandler(srvView, ds, lg)
	return &ServerMsgHandler{handlePaxos: paxosHandler, handle2pc: tpcHandler, handleTransfer: handleTransfer}
}

func (self *ServerMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
//	fmt.Println(ctnt.Head)
//	fmt.Println(ctnt.Body)
	// Examine the content
	//fmt.Println("in serverMsgHandler", ctnt)
	switch ctnt.Head {
	case "2pc":
		return self.handle2pc.Handle(ctnt, reply)
	case "paxos":
		return self.handlePaxos.Handle(ctnt, reply)
	case "transfer":
		return self.handleTransfer.Handle(ctnt, reply)
	}
	return nil
}
