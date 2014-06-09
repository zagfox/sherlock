package paxos

import (
	"encoding/json"
	"fmt"
	//"time"
	"sherlock/common"
)

var _ common.MsgHandlerIf = new(PaxosMsgHandler)

type PaxosMsgHandler struct {
	srvView *ServerView
	lg      common.LogPlayerIf
}

func NewPaxosMsgHandler(srvView *ServerView, lg common.LogPlayerIf) common.MsgHandlerIf {
	return &PaxosMsgHandler{srvView: srvView, lg: lg}
}

// Handle paxos message, ctnt.head is "paxos" already
func (self *PaxosMsgHandler) Handle(ctnt common.Content, reply *common.Content) error {
	//	fmt.Println("in paxosMsgHandler")
	pb := StringToPaxos(ctnt.Body)
	//	fmt.Println(pb)
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
			ProNumPair: common.ProposalNumPair{-1, -1},
			ProValue:   -1,
			VID:        vid, View: view})
		return nil
	}

	// then check if n is bigger than n_h
	np_a, v_a := self.srvView.GetAcceptedValue()
	np_h := self.srvView.GetHighestNumPair()
	if pb.ProNumPair.BiggerEqualThan(np_h) {
		// set self state to be updating
		self.srvView.SetState(common.SrvUpdating)

		// set np_h
		self.srvView.SetHighestNumPair(pb.ProNumPair)

		// change, set returned value is logid
		logId := self.lg.GetLogID()

		// reply it with prepare ok
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "prepare", Action: "ok",
			ProNumPair:  np_a,
			ProValue:    v_a,
			DecideValue: logId,
			VID:         -1, View: nil})
		return nil
	} else {
		// reject
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "prepare", Action: "reject",
			//ProNumPair: common.ProposalNumPair{-1, -1},
			ProNumPair: np_h,
			ProValue:   -1,
			VID:        -1, View: nil})
		return nil
	}
}

// handle phase2
func (self *PaxosMsgHandler) HandleAccept(pb common.PaxosBody, reply *common.Content) error {
	// first check if it is old view
	vid, view := self.srvView.GetView()
	if pb.VID <= vid {
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "accept", Action: "oldview",
			ProNumPair: common.ProposalNumPair{-1, -1},
			ProValue:   -1,
			VID:        vid, View: view})
		return nil
	}

	// then check if n is bigger than n_h
	//np_a, v_a := self.srvView.GetAcceptedValue()
	np_h := self.srvView.GetHighestNumPair()
	if pb.ProNumPair.BiggerEqualThan(np_h) {
		// accept value, record numpair and value
		self.srvView.SetAcceptedValue(pb.ProNumPair, pb.ProValue)

		// reply it with accept ok
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "accept", Action: "ok",
			ProNumPair: common.ProposalNumPair{},
			ProValue:   -1,
			VID:        -1, View: nil})
		return nil
	} else {
		// reject
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "accept", Action: "reject",
			ProNumPair: np_h,
			//ProNumPair: common.ProposalNumPair{-1, -1},
			ProValue:   -1,
			VID:        -1, View: nil})
		return nil
	}
}

// handle phase3
func (self *PaxosMsgHandler) HandleDecide(pb common.PaxosBody, reply *common.Content) error {
	// first check if it is old view
	vid, view := self.srvView.GetView()
	if pb.VID <= vid {
		reply.Head = "paxos"
		reply.Body = PaxosToString(common.PaxosBody{
			Phase: "decide", Action: "oldview",
			ProNumPair: common.ProposalNumPair{-1, -1},
			ProValue:   -1,
			VID:        vid, View: view})
		return nil
	}

	fmt.Println("paxosHandler", self.srvView.Id, "> receive decide: mid =", pb.ProValue, "view=", pb.VID, pb.View)

	// check if self is master, then transfer data to new node
	if self.srvView.Id == pb.ProValue {
		nodes := self.srvView.NodesNotInView(pb.View)
		if len(nodes) != 0 {

			// prepare send message
			var ctnt, reply common.Content
			bytes, _ := json.Marshal(self.lg.GetStoreWraper())
			ctnt.Head = "transfer"
			ctnt.Body = string(bytes)

			// send message
			for _, v := range nodes {
				err := self.srvView.SendMsg(v, ctnt, &reply)
				if err != nil {
					fmt.Println("paxos msg handler", self.srvView.Id, "handle decide")
				}
			}
		}
	}

	// set self mid, vid and view, state
	self.srvView.SetMasterId(pb.ProValue)
	self.srvView.SetView(pb.VID, pb.View)
	self.srvView.SetState(common.SrvReady)

	// reply it with accept ok
	reply.Head = "paxos"
	reply.Body = PaxosToString(common.PaxosBody{
		Phase: "decide", Action: "ok",
		ProNumPair: common.ProposalNumPair{},
		ProValue:   -1,
		VID:        -1, View: nil})
	return nil

}
