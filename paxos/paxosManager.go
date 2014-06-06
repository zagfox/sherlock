package paxos

import (
	"fmt"
	"sync"
	"math"
	//"strconv"
	"sherlock/common"
)

type PaxosManager struct {
	Id  int //self id
	Num int // total lock server number

	vid   int   // id for view
	view  []int // members in current view, 0 means not in, 1 means in view
	vlock sync.Mutex

	next_view []int     // tmp place to place changed view
	nvlock sync.Mutex
	responders []int

	srvs []common.MessageIf //interface to talk to other server

	//paxos related variable
	np_a      common.ProposalNumPair
	v_a       int  // last accepted proposal n, v

	np_h      common.ProposalNumPair  // highest n seen in progress
	my_np	  common.ProposalNumPair

	// logging info
	logging bool
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
		np_a:  common.ProposalNumPair{-1, -1}, v_a: -1,
		np_h: common.ProposalNumPair{-1, -1},
		my_np: common.ProposalNumPair{-1, -1},

		// logging
		logging: true,
	}
}

// reset paxos state for restart
func (self *PaxosManager) Reset() {
	view := make([]int, 0)
	for i := 0; i < self.Num; i++ {
		view = append(view, i)
	}

	//self.SetView(0, view)  //view remain, cause it may learn from others
	self.SetAcceptedValue(common.ProposalNumPair{-1, -1}, -1)
	self.SetHighestNumPair(common.ProposalNumPair{-1, -1})
}

func (self *PaxosManager) Logln(str string) {
	if self.logging == true {
		fmt.Println("PaxosManager", self.Id, ">", str);
	}
}

// Get/Set vid and view number
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

// check if a node is in view
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


// check how many nodes are in view
func (self *PaxosManager) NodesInView(nodes []int) int {
	num := 0

	for i := 0; i < len(nodes); i++ {
		if self.NodeInView(nodes[i]) {
			num++
		}
	}

	return num
}

// Deprecated
// add/delete node in next view
func (self *PaxosManager) RequestAddNode(nid int) {
	self.nvlock.Lock()
	defer self.nvlock.Unlock()

/*
	// check if node exist
	if self.next_view[nid] == 0 {
		self.next_view[nid] = 1
	}
	*/
}

func (self *PaxosManager) RequestDelNode(nid int) {
	self.nvlock.Lock()
	defer self.nvlock.Unlock()

	/*
	//self.Logln("requestDelNode" + "nid=" + strconv.FormatInt(int64(nid), 10))
	// check if node exist
	if self.next_view[nid] == 1 {
		self.next_view[nid] = 0
	}
	*/
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
 * send phase1: prepare
 * send phase2: accept
 * send phase3: decide
 * return mid, info
 */
func (self *PaxosManager) updateView() (int, string) {
	var mid  int
	var info string

	self.Logln("->begin update")
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
	var ctnt common.Content
	reply := make([]common.Content, self.Num)

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
	np_tmp := common.ProposalNumPair{-1, -1}  //temp highest n
	v_tmp := -1          // value corresponding to np_tmp
	responders := make([]int, 0)    // keep record to responders

	self.Logln("phasePrepare, sending all messeges")
	//send pprepare to everyone
	ch_finish := make(chan string, self.Num)
	for v := 0; v < self.Num; v++ {
		e := self.srvs[v].Msg(ctnt, &reply[v])
		if e != nil {
			// network error during paxos
			// Plan: go on
		}
		// fmt.Println(reply[v])
		ch_finish <- "finish"
	}

	// wait for all response
	for v := 0; v < self.Num; v++ {
		<-ch_finish
	}

	// get response from every one
	for v := 0; v < self.Num; v++ {
		reply_pb := StringToPaxos(common.Content(reply[v]).Body)

		if reply_pb.Action == "oldview" {
			// believe it and restart paxos
			self.SetView(reply_pb.VID, reply_pb.View)
			return common.PaxosRestart

		} else if reply_pb.Action == "reject" {
			//go on

		} else if reply_pb.Action == "ok" {
			responders = append(responders, v)
			if reply_pb.ProNumPair.BiggerEqualThan(np_tmp) {
				//update np_tmp and v_tmp
				// also send to itself, so must have value
				np_tmp = reply_pb.ProNumPair
				v_tmp = reply_pb.ProValue
			}

		} else {
			// should be empty, or don't know what's gotten
			//self.Logln("phasePrepare, don't know what's gotten back")
			//return common.PaxosFailure
		}
	}

	//if majority reply, prepare success
	if self.NodesInView(responders) > len(view)/2 {
		// set value and return success
		// should be right, though a little different from lecture nodes
		// because it would always send to self
		/*if v_tmp == -1 || !self.NodeInView(v_tmp) {  // this is incorrect, node may be down at this time
			//if no one has valid reply for value, select self to be master
			np_tmp = self.my_np
			v_tmp = self.Id
		}*/
		// use a brute method, always listen to me!
		np_tmp = self.my_np
		v_tmp = self.Id

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
	var ctnt common.Content
	reply := make([]common.Content, self.Num)

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
	responders := make([]int, 0)    // keep record to responders

	//send pprepare to everyone
	ch_finish := make(chan string, self.Num)
	for v := 0; v < self.Num; v++ {
		e := self.srvs[v].Msg(ctnt, &reply[v])
		if e != nil {
			// error during paxos
			// Plan: go on
		}
		ch_finish <- "finish"
	}

	// wait for all response
	for v := 0; v < self.Num; v++ {
		<-ch_finish
	}

	// get response from every one
	for v := 0; v < self.Num; v++ {
		reply_pb := StringToPaxos(reply[v].Body)

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
			responders = append(responders, v)
		} else {
			// don't know what's gotten
			//return common.PaxosFailure
		}
	}

	//if majority reply, accept success
	// set self view to be responders
	if self.NodesInView(responders) > len(view)/2 {
		self.responders = responders
		return common.PaxosSuccess
	}

	//fmt.Println("paxosManager : accept-> not enough reply")
	return common.PaxosFailure
}



// phase3, paxos decide
// 1. send (vid+1, view, value) to all in view
// return value, info
func (self *PaxosManager) phaseDecide() (int, string) {
	var ctnt common.Content
	reply := make([]common.Content, self.Num)

	// the view has been updated to accept responders
	// get current view
	vid, view := self.GetView()
	fmt.Println("PaxosManager", self.Id, "> phaseDecide", vid, view)

	// generate message (my_n, vid+1, value)
	ctnt.Head = "paxos"
	ctnt.Body = PaxosToString(common.PaxosBody{
		Phase: "decide", Action: "request",
		ProNumPair:      common.ProposalNumPair{},
		ProValue:        self.v_a,
		// set next_view, to be phase2 responders
		VID:             vid + 1, View: self.responders})

	//send pprepare to everyone
	ch_finish := make(chan string, self.Num) //len(view))
	for v := 0; v < self.Num; v++ {
		e := self.srvs[v].Msg(ctnt, &reply[v])
		if e != nil {
			// error during paxos
			// Plan: go on
		}
		ch_finish <- "finish"
	}

	// wait for all response
	for v := 0; v < self.Num; v++ {
		<-ch_finish
	}

	// get response from every one
	for v := 0; v < self.Num; v++ {
		reply_pb := StringToPaxos(reply[v].Body)

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
			//return -1, common.PaxosFailure
		}
	}

	// decide phase always success
	 return self.v_a, common.PaxosSuccess
}

