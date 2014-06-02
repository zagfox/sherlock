package paxos

import (
	"fmt"
	"sync"
	"math"
	"sherlock/common"
)

type PaxosManager struct {
	Id  int //self id
	Num int // total lock server number

	vid   int   // id for view
	view  []int // members in current view, 0 means not in, 1 means in view
	vlock sync.Mutex

	srvs []common.MessageIf //interface to talk to other server


	//paxos related variable
	//n_a, id_a, v_a int  // last accepted proposal n, v
	np_a      common.ProposalNumPair
	v_a       int  // last accepted proposal n, v
	//n_h, id_h      int  // highest n seen in progress
	np_h      common.ProposalNumPair  // highest n seen in progress
	//my_n      int
	my_np	  common.ProposalNumPair
}

func NewPaxosManager(Id, Num int, srvs []common.MessageIf) *PaxosManager {
	view := make([]int, Num)
	for i := 0; i < Num; i++ {
		view[i] = 1
	}
	return &PaxosManager{
		Id: Id, Num: Num,

		//view related
		vid: 0, view: view,
		srvs: srvs,
		np_a:  common.ProposalNumPair{-1, -1}, v_a: 0,
		np_h: common.ProposalNumPair{-1, -1},
		my_np: common.ProposalNumPair{-1, -1},
	}
}

func (self *PaxosManager) GetView() (int, []int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	ret := make([]int, 0)
	for i, v := range self.view {
		if v == 1 {
			ret = append(ret, i)
		}
	}

	return self.vid, ret
}

// Set vid and view number
func (self *PaxosManager) SetView(vid int, view []int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	self.vid = vid
	for i, _ := range self.view {
		self.view[i] = 0
	}
	for _, v := range view {
		self.view[v] = 1
	}
}

func (self *PaxosManager) NodeInView(nid int) bool {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	if nid >= len(self.view) || nid < 0{
		return false
	}

	if self.view[nid] == 1 {
		return true
	} else {
		return false
	}
}
func (self *PaxosManager) AddNode(nid int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	// check if node exist
	if self.view[nid] == 0 {
		self.vid++
		self.view[nid] = 1
	}
}

func (self *PaxosManager) DelNode(nid int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	// check if node exist
	if self.view[nid] == 1 {
		self.vid++
		self.view[nid] = 0
	}
}

// function to get accepted value
func (self *PaxosManager) GetAcceptedValue() (common.ProposalNumPair, int) {
	return self.np_a, self.v_a
}

func (self *PaxosManager) SetAcceptedValue(np_a common.ProposalNumPair, v_a int) {
	self.np_a = np_a
	self.v_a = v_a
}

// function to get highest value
func (self *PaxosManager) GetHighestNumPair() (common.ProposalNumPair) {
    return self.np_h
}

func (self *PaxosManager) SetHighestNumPair(np_h common.ProposalNumPair) {
	self.np_h = np_h
}

/*
 * send phase1
 * send phase2
 * send phase3
 * return mid, info
 */
func (self *PaxosManager) updateView() (int, string) {
	var mid  int
	var info string

	// prepare phase
	info = self.phasePrepare()
	if info != common.PaxosSuccess {
		return -1, info
	}
	fmt.Println("paxos", self.Id, "->prepare success")

	// accept phase
	info = self.phaseAccept()
	if info != common.PaxosSuccess {
		return -1, info
	}
	fmt.Println("paxos", self.Id, "->accept success")

	// decide phase
	mid, info = self.phaseDecide()
	if info != common.PaxosSuccess {
		return -1, info
	}
	fmt.Print("paxos", self.Id, "->decide success! ")
	fmt.Print("mid = ", mid)
	vid, view := self.GetView()
	fmt.Println(" view = ", vid, view)

	return mid, common.PaxosSuccess
}

// phase1, paxos prepare
// 1.my_n = max(n_h, my_n) + 1, send prepare_request(my_n, vid+1) to nodes in ?
// 2. if receive prepare ok from majority, phase1 success, else handle it...
// return prepare_state
func (self *PaxosManager) phasePrepare() string {
	var ctnt, reply common.Content

	// create variable to send prepare request
	vid, view := self.GetView()
	ctnt.Head = "paxos"

	// generate my_n
	ProNum := int(math.Max(float64(self.my_np.ProposalNum), float64(vid))) + 1
	self.my_np = common.ProposalNumPair{ProNum, self.Id}
	ctnt.Body = PaxosToString(common.PaxosBody{
		Phase: "prepare", Action: "request",
		ProNumPair:      self.my_np,
		ProValue:        -1,
		VID:             vid + 1, View: nil})

	// record returned value
	num_tmp := 0 //the response get
	np_tmp := common.ProposalNumPair{-1, -1}  //temp highest n
	v_tmp := -1          // value corresponding to np_tmp

	//send pprepare to everyone
	for _, v := range view {
		e := self.srvs[v].Msg(ctnt, &reply)
		if e != nil {
			// error during paxos
			// Plan: return failure
			return common.PaxosFailure
		}
		reply_pb := StringToPaxos(reply.Body)

		if reply_pb.Action == "oldview" {
			// believe it and restart paxos
			self.SetView(reply_pb.VID, reply_pb.View)
			return common.PaxosRestart

		} else if reply_pb.Action == "reject" {
			//go on

		} else if reply_pb.Action == "ok" {
			num_tmp++
			if reply_pb.ProNumPair.BiggerEqualThan(np_tmp) {
				//update np_tmp and v_tmp
				// also send to itself, so must have value
				np_tmp = reply_pb.ProNumPair
				v_tmp = reply_pb.ProValue
			}

		} else {
			// don't know what's gotten
			return common.PaxosFailure
		}
	}

	//if majority reply, prepare success
	if num_tmp > len(view)/2 {
		// set value and return success
		// should be right, though a little different from lecture nodes
		// because it would always send to self
		if v_tmp == -1 {
			//if no one reply for value, select self to be master
			v_tmp = self.Id
		}
		self.SetAcceptedValue(np_tmp, v_tmp)
		return common.PaxosSuccess
	}

	return common.PaxosFailure
}

// phase2, paxos accept
// 1. send (my_n, vid+1, value) to all in view
// 2. if receive accept ok from majority, phase1 success, else handle it...
// return prepare_state
func (self *PaxosManager) phaseAccept() string {
	var ctnt, reply common.Content

	// get current view
	vid, view := self.GetView()

	// generate message (my_n, vid+1, value)
	ctnt.Head = "paxos"
	ctnt.Body = PaxosToString(common.PaxosBody{
		Phase: "accept", Action: "request",
		ProNumPair:      self.my_np,
		ProValue:        self.v_a,
		VID:             vid + 1, View: nil})

	// record accept ok num
	num_tmp := 0 //the response get

	//send pprepare to everyone
	for _, v := range view {
		e := self.srvs[v].Msg(ctnt, &reply)
		if e != nil {
			// error during paxos
			// Plan: return failure
			return common.PaxosFailure
		}
		reply_pb := StringToPaxos(reply.Body)

		if reply_pb.Action == "oldview" {
			// believe it and restart paxos
			//fmt.Println("paxosManager : accept-> old view")
			self.SetView(reply_pb.VID, reply_pb.View)
			return common.PaxosRestart

		} else if reply_pb.Action == "reject" {
			// go on

		} else if reply_pb.Action == "ok" {
			// record reply num
			num_tmp++
		} else {
			// don't know what's gotten
			return common.PaxosFailure
		}
	}

	//fmt.Println(num_tmp)
	//fmt.Println(len(view)/2)
	//if majority reply, prepare success
	if num_tmp > len(view)/2 {
		return common.PaxosSuccess
	}

	//fmt.Println("paxosManager : accept-> not enough reply")
	return common.PaxosFailure
}



// phase3, paxos decide
// 1. send (vid+1, view, value) to all in view
// return value, info
func (self *PaxosManager) phaseDecide() (int, string) {
	var ctnt, reply common.Content

	// get current view
	vid, view := self.GetView()

	// generate message (my_n, vid+1, value)
	ctnt.Head = "paxos"
	ctnt.Body = PaxosToString(common.PaxosBody{
		Phase: "decide", Action: "request",
		ProNumPair:      common.ProposalNumPair{},
		ProValue:        self.v_a,
		VID:             vid + 1, View: view})

	//send pprepare to everyone
	for _, v := range view {
		e := self.srvs[v].Msg(ctnt, &reply)
		if e != nil {
			// error during paxos
			// Plan: return failure
			return -1, common.PaxosFailure
		}
		reply_pb := StringToPaxos(reply.Body)

		if reply_pb.Action == "oldview" {
			// believe it and restart paxos
			self.SetView(reply_pb.VID, reply_pb.View)
			return -1, common.PaxosRestart

		} else if reply_pb.Action == "reject" {
			// go on

		} else if reply_pb.Action == "ok" {
			// nothing special here

		} else {
			// don't know what's gotten
			return -1, common.PaxosFailure
		}
	}

	// decide phase always success
	 return self.v_a, common.PaxosSuccess
}

