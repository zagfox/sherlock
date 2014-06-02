package paxos

import (
	"fmt"
	"sync"
	"time"
	"errors"
	"sherlock/common"
)

// ServerView, used to maintain view group and info
type ServerView struct {
	Id      int           //self id
	Num  int           // total lock server number

	mid     int           //master server id
	midLock sync.Mutex

	cntReq int            // number of request for uptate
	cntLock sync.Mutex

	state     string      //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex

	srvs     []common.MessageIf  //interface to talk to other server

	//paxos related variable
	paxosMgr *PaxosManager
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
		mid: mid,
		state: state,
		srvs: srvs,
		paxosMgr: paxosMgr,
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
	return self.paxosMgr.GetView()
}

// Set vid and view number
func (self *ServerView) SetView(vid int, view []int) {
	self.paxosMgr.SetView(vid, view)
}

func (self *ServerView) AddNode(nid int) {
	self.paxosMgr.AddNode(nid)
}

func (self *ServerView) DelNode(nid int) {
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

// function to get accepted value
func (self *ServerView) GetAcceptedValue() (common.ProposalNumPair, int) {
	return self.paxosMgr.GetAcceptedValue()
}

// function to get highest value
func (self *ServerView) GetHighestNumPair() (common.ProposalNumPair) {
	return self.paxosMgr.GetHighestNumPair()
}

func (self *ServerView) SetHighestNumPair(np_h common.ProposalNumPair) {
    self.paxosMgr.SetHighestNumPair(np_h)
}

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

	// updateview
	_, info := self.paxosMgr.updateView()
	for ;info == common.PaxosRestart;  {
		// info is restart, then do again
		fmt.Println("get something, restart paxos in 1s")
		time.Sleep(1000*time.Millisecond)
		_, info = self.paxosMgr.updateView()
	}
	if info != common.PaxosSuccess {
		fmt.Println("serverview update fail")
	}

	//self.SetMasterId(mid)
	self.SetState(common.SrvReady)
	return nil
}

