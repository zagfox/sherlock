package backend

import (
	"encoding/json"
	"sherlock/common"
	"sherlock/lockstore"
	"sherlock/paxos"
)

var _ common.MsgHandlerIf = new(TransferMsgHandler)

type TransferMsgHandler struct {
	view *paxos.ServerView
	ds common.DataStoreIf
	lg *lockstore.LogPlayer
}

func NewTransferMsgHandler(view *paxos.ServerView, ds common.DataStoreIf, lg *lockstore.LogPlayer) common.MsgHandlerIf {
	return &TransferMsgHandler{view:view, ds:ds, lg:lg}
}

//Handles the message involved with data transfer between the servers
func (self *TransferMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	// unmarshal from string
	var sw common.StoreWraper
	json.Unmarshal([]byte(ctnt.Body), &sw)

	// do transfer
	self.ds.ApplyWraper(sw)
	self.lg.ApplyWraper(sw)
	return nil
}
