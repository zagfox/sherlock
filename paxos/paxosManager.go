package paxos

import (
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
	n_a, v_a int
	n_h      int
	my_n     int
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
		n_a:  0, v_a: 0,
		n_h: 0, my_n: 0,
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

/*
 * send phase1
 * send phase2
 * send phase3
 *return info
 */
func (self *PaxosManager) updateView() (int, string) {
	var mid  int
	var info string

	info = self.prepare()
	if info != common.PaxosSuccess {
		return -1, info
	}
	return mid, common.PaxosSuccess
}

// phase1, paxos prepare
// my_n = max(n_h, my_n) + 1
// send prepare_request(my_n, vid+1) to nodes in ?
// return prepare_state
func (self *PaxosManager) prepare() string {
	var ctnt, reply common.Content

	vid, view := self.GetView()
	ctnt.Head = "paxos"

	ctnt.Body = PaxosToString(common.PaxosBody{
		Phase: "prepare", Action: "request",
		ProposerId:    self.Id,
		ProposalNum:   int(math.Max(float64(self.my_n), float64(vid))) + 1,
		ProposalValue: -1,
		VID:           vid + 1, View: nil})
	for _, v := range view {
		e := self.srvs[v].Msg(ctnt, &reply)
		if e != nil {
			// error during paxos
			// Plan to restart
		}
		reply_pb := StringToPaxos(reply.Body)
		if reply_pb.Action == "oldview" {
			// believe it and restart paxos
			self.SetView(reply_pb.VID, reply_pb.View)
			return common.PaxosRestart
		} else if reply_pb.Action == "reject" {
			// restart paxos
			return common.PaxosRestart
		}

	}

	return "success"
}
