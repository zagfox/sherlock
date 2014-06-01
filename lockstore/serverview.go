package lockstore

import (
	"fmt"
	"sync"
)

type ServerView struct {
	Id      int //self id
	mid     int //master server id
	midLock sync.Mutex

	vid   int   // id for view
	view  []int // members in current view
	vlock sync.Mutex

	state     string //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex
}

func NewServerView(Id, mid int, state string) *ServerView {
	return &ServerView{
		Id: Id, mid: mid,
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

	return self.vid, self.view
}

func (self *ServerView) SetView(vid int, view []int) {
	self.vlock.Lock()
	defer self.vlock.Unlock()

	self.vid = vid
	self.view = view
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

func (self *ServerView) UpdateView() error {
	fmt.Println("update view")
	return nil
}
