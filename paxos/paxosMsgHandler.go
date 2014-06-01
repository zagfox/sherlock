package paxos

import (
	"fmt"
	"sherlock/common"
)

var _ common.MsgHandlerIf = new(PaxosMsgHandler)

type PaxosMsgHandler struct {
	srvView *ServerView
}

func NewPaxosMsgHandler(srvView *ServerView) common.MsgHandlerIf {
	return &PaxosMsgHandler{srvView: srvView}
}

// Handle paxos message, ctnt.head is "paxos" already
func (self *PaxosMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	fmt.Println("in paxosMsgHandler")
	pb := StringToPaxos(ctnt.Body)
	fmt.Println(pb)
	switch pb.Phase {
	case "prepare":
		return self.HandlePrepare(pb, reply)
	case "accept":
		return self.HandleAccept(pb, reply)
	case "decide":
		return self.HandleDecide(pb, reply)
	}
	return nil
}

// handle phase1
func (self *PaxosMsgHandler) HandlePrepare(pb common.PaxosBody, reply *common.Content) error {
	_, view := self.srvView.GetView()
	//if pb.VID <= vid {
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "prepare", Action: "oldview",
			ProposerId:    -1,
			ProposalNum:   -1,
			ProposalValue: -1,
			VID: -1, View: view})
		return nil
	//}
	//return nil
}

// handle phase2
func (self *PaxosMsgHandler) HandleAccept(pb common.PaxosBody, reply *common.Content) error {
	return nil
}

// handle phase3
func (self *PaxosMsgHandler) HandleDecide(pb common.PaxosBody, reply *common.Content) error {
	return nil
}
