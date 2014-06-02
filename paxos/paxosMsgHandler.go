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
	// first check if it is old view
	vid, view := self.srvView.GetView()
	if pb.VID <= vid {
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "prepare", Action: "oldview",
			//ProposerId:    -1,
			//ProposalNum:   -1,
			ProNumPair:    common.ProposalNumPair{-1, -1},
			ProValue:      -1,
			VID:           -1, View: view})
		return nil
	}

	// then check if n is bigger than n_h
	np_a, v_a := self.srvView.GetAcceptedValue()
	np_h := self.srvView.GetHighestNumPair()
	if pb.ProNumPair.BiggerThan(np_h) {
		// set self state to be updating
		self.srvView.SetState(common.SrvUpdating)

		// set n_h
		self.srvView.SetHighestNumPair(pb.ProNumPair)

		// reply it with prepare ok
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "prepare", Action: "ok",
			//ProposerId:    id_a,
			//ProposalNum:   n_a,
			ProNumPair:     np_a,
			ProValue:       v_a,
			VID:            -1, View: nil})
		return nil
	} else {
		// reject
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "prepare", Action: "reject",
			//ProposerId:    -1,
			//ProposalNum:   -1,
			ProNumPair:    common.ProposalNumPair{-1, -1},
			ProValue:      -1,
			VID:           -1, View: nil})
		return nil
	}
}

// handle phase2
func (self *PaxosMsgHandler) HandleAccept(pb common.PaxosBody, reply *common.Content) error {
	return nil
}

// handle phase3
func (self *PaxosMsgHandler) HandleDecide(pb common.PaxosBody, reply *common.Content) error {
	return nil
}
