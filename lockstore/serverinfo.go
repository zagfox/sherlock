package lockstore

import (
	"sync"
)

type ServerInfo struct {
	Id      int           //self id
	mid     int           //master server id
	midLock sync.Mutex

	state     string      //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex
}

func NewServerInfo(Id, mid int, state string) *ServerInfo {
	return &ServerInfo{
		Id: Id, mid: mid,
		state: state,
	}
}

// master mid modify interface
func (self *ServerInfo) SetMasterId(mid int) {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	self.mid = mid
}

func (self *ServerInfo) GetMasterId() int {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	return self.mid
}

// function to set lockserver state
func (self *ServerInfo) GetState() string {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	return self.state
}

func (self *ServerInfo) SetState(state string) {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	self.state = state
}
