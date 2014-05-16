//Lock service frontend client

package frontend

import (
	//"fmt"
	"net/rpc"
	"sync"

	"sherlock/common"
)

type client struct {
	addr string
	srv  *rpc.Client
	lock sync.Mutex
}

func NewClient(addr string) common.LockStoreIf {
	return &client{addr: addr}
}

func (self *client) connect(firsttime bool) error {
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

func (self *client) Acquire(lu common.LUpair, succ *bool) error {
	if e := self.connect(true); e != nil {
		return e
	}
	for e := self.srv.Call("LockStore.Acquire", lu, succ); e != nil; {
		//fmt.Println(e)
		if e := self.connect(false); e != nil {
			return e
		}
	}
	return nil

}

func (self *client) Release(lu common.LUpair, succ *bool) error {
	if e := self.connect(true); e != nil {
		return e
	}
	for e := self.srv.Call("LockStore.Release", lu, succ); e != nil; {
		if e := self.connect(false); e != nil {
			return e
		}
	}
	return nil
}
