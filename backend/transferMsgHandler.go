package backend

import (
	"sherlock/common"
	"sherlock/lockstore"
	"sherlock/paxos"
)

var _ common.MsgHandlerIf = new(TransferMsgHandler)

type TransferMsgHandler struct {
	view *paxos.ServerView
	lg *lockstore.LogPlayer
}

func NewTransferMsgHandler(view *paxos.ServerView, lg *lockstore.LogPlayer) common.MsgHandlerIf {
	return &TransferMsgHandler{view:view, lg:lg}
}

//Handles the message involved with data transfer between the servers
func (self *TransferMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	//TODO
	return nil
}
