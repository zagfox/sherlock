package sherlock

import (
	"sync"
	"net/rpc"
	"sherlock/common"
)

// msg client
//var _ common.SherlockIf = new(SherClient)

type SherClient struct {
	// private fields
	addr string         //address to connect
	srv  *rpc.Client    // rpc client
	lock sync.Mutex
}

func NewSherClient() common.SherlockIf {
	return &SherClient{addr: "localhost:26999"}
}

func (self *SherClient) connect(firsttime bool) error {
	self.lock.Lock()
	defer self.lock.Unlock()
	if (self.srv != nil) && firsttime {
		return nil
	}
	var err error
	for i := 0; i < 3; i++ {
		self.srv, err = rpc.DialHTTP("tcp", self.addr)
		if err == nil {
			return nil
		}
	}
	return err
}

func (self *SherClient) Acquire(lname string, succ *bool) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("SherListener.Acquire", lname, succ)
	for ; e != nil; e = self.srv.Call("SherListener.Acquire", lname, succ) {
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}

func (self *SherClient) Release(lname string, succ *bool) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("SherListener.Release", lname, succ)
	for ; e != nil; e = self.srv.Call("SherListener.Release", lname, succ) {
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}
