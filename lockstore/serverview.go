package lockstore

import (
	"fmt"
	"sync"
	"errors"
)

// ServerView, used to maintain view group and info
type ServerView struct {
	Id      int           //self id
	Num  int           // total lock server number

	mid     int           //master server id
	midLock sync.Mutex

	vid    int            // id for view
	view   []int          // members in current view, 0 means not in, 1 means in view
	vlock sync.Mutex

	cntReq int            // number of request for uptate
	cntLock sync.Mutex

	state     string      //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex
}

func NewServerView(Id, Num, mid int, state string) *ServerView {
	// suppose the first view has all the member
	view := make([]int, Num)
	for i := 0; i < Num; i++ {
		view[i] = 1
	}
	return &ServerView{
		Id: Id, Num: Num,
		mid: mid,
		vid: 0, view: view,
		state: state,
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
	self.vlock.Lock()
	defer self.vlock.Unlock()

	ret := make([]int, 0)
	for i, v := range(self.view) {
		if v == 1 {
			ret = append(ret, i)
		}
	}

	return self.vid, ret
}

// Set vid and view number
func (self *ServerView) SetView(vid int, view []int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	self.vid = vid
	for i, _ := range(self.view) {
		self.view[i] = 0
	}
	for _, v := range(view) {
		self.view[v] = 1
	}
}

func (self *ServerView) AddNode(nid int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	// check if node exist
	if self.view[nid] == 0 {
		self.vid++
		self.view[nid] = 1
	}
}

func (self *ServerView) DelNode(nid int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	// check if node exist
	if self.view[nid] == 1 {
		self.vid++
		self.view[nid] = 0
	}
}


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

	cntReq := self.getCntReq()
	if self.cntReq == 0 {
		self.setCntReq(cntReq+1)
		return self.updateView()
	} else {
		return errors.New("already updating")
	}
}

// update the view group member
func (self *ServerView) updateView() error {
	fmt.Println("updating view")
	return nil
}

