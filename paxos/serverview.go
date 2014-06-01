package paxos

import (
	"fmt"
	"sync"
	"errors"
	"sherlock/common"
)

// ServerView, used to maintain view group and info
type ServerView struct {
	Id      int           //self id
	Num  int           // total lock server number

	mid     int           //master server id
	midLock sync.Mutex

	/*vid    int            // id for view
	view   []int          // members in current view, 0 means not in, 1 means in view
	vlock sync.Mutex
	*/

	cntReq int            // number of request for uptate
	cntLock sync.Mutex

	state     string      //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex

	srvs     []common.MessageIf  //interface to talk to other server

	//paxos related variable
	paxosMgr *PaxosManager
	/*n_a, v_a int
	n_h      int
	my_n     int
	*/
}

func NewServerView(Id, Num, mid int, state string, srvs []common.MessageIf) *ServerView {
	// suppose the first view has all the member
	view := make([]int, Num)
	for i := 0; i < Num; i++ {
		view[i] = 1
	}

	paxosMgr := NewPaxosManager(Id, Num, srvs)
	return &ServerView {
		Id: Id, Num: Num,
		//mid: mid,
		//vid: 0, view: view,
		state: state,
		srvs: srvs,
		paxosMgr: paxosMgr,
		//n_a: 0, v_a: 0,
		//n_h: 0, my_n: 0,
	}
}

// master mid modify interface
func (self *ServerView) SetMasterId(mid int) {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	self.mid = mid
}

func (self *ServerView) GetMasterId() int {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	return self.mid
}

// Get vid and view member
func (self *ServerView) GetView() (int, []int) {
	/*self.vlock.Lock()
	defer self.vlock.Unlock()

	ret := make([]int, 0)
	for i, v := range(self.view) {
		if v == 1 {
			ret = append(ret, i)
		}
	}

	return self.vid, ret
	*/
	return self.paxosMgr.GetView()
}

// Set vid and view number
func (self *ServerView) SetView(vid int, view []int) {
	/*self.vlock.Lock()
	defer self.vlock.Unlock()

	self.vid = vid
	for i, _ := range(self.view) {
		self.view[i] = 0
	}
	for _, v := range(view) {
		self.view[v] = 1
	}*/
	self.paxosMgr.SetView(vid, view)
}

func (self *ServerView) AddNode(nid int) {
	/*
	self.vlock.Lock()
	defer self.vlock.Unlock()

	// check if node exist
	if self.view[nid] == 0 {
		self.vid++
		self.view[nid] = 1
	}*/
	self.paxosMgr.AddNode(nid)
}

func (self *ServerView) DelNode(nid int) {
	/*
	self.vlock.Lock()
	defer self.vlock.Unlock()

	// check if node exist
	if self.view[nid] == 1 {
		self.vid++
		self.view[nid] = 0
	}*/
	self.paxosMgr.DelNode(nid)
}

/*
// function to operate on cntReq
func (self *ServerView) getCntReq() int {
	self.cntLock.Lock()
	defer self.cntLock.Unlock()

	return self.cntReq
}

func (self *ServerView) setCntReq(cntReq int) {
	self.cntLock.Lock()
	defer self.cntLock.Unlock()

	self.cntReq = cntReq
}
*/

// function to set lockserver state
func (self *ServerView) GetState() string {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	return self.state
}

func (self *ServerView) SetState(state string) {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	self.state = state
}

// request to update view
func (self *ServerView) RequestUpdateView() error {
	self.cntLock.Lock()
	defer self.cntLock.Unlock()

	cntReq := self.cntReq
	if cntReq == 0 {
		self.cntReq++
		return self.updateView()
	} else {
		return errors.New("already updating")
	}
}

// update the view group member
/*
 * set self state
 * do update view using paxos
 * do migration here?
 * set self state back
 */
func (self *ServerView) updateView() error {
	fmt.Println("updating view")

	self.SetState(common.SrvUpdating)

	self.paxosMgr.updateView()
	//self.paxosPrepare()

	self.SetState(common.SrvReady)
	return nil
}

/*
// phase1, paxos prepare
// my_n = max(n_h, my_n) + 1
// send prepare_request(my_n, vid+1) to nodes in ?
func (self *ServerView) paxosPrepare() {
	var ctnt, reply common.Content

	vid, view := self.GetView()
	ctnt.Head = "paxos"

	ctnt.Body = PaxosToString(common.PaxosBody{
		Phase: "prepare", Action:"request",
		ProposerId: self.Id,
		ProposalNum: int(math.Max(float64(self.my_n), float64(vid)))+1,
		ProposalValue: -1,
		VID: vid+1,View: nil})
	for _, v := range(view) {
		e := self.srvs[v].Msg(ctnt, &reply)
		if e != nil {
			// error during paxos
			// Plan to restart
		}
		reply_pb := StringToPaxos(reply.Body)
		if reply_pb.Action == "oldview" {
			self.SetView(reply_pb.VID, reply_pb.View)
		}
	}

}*/
