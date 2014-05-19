//Lock service frontend client

package frontend

import (
	"fmt"
	"net/rpc"
	"sync"

	"sherlock/common"
	//"sherlock/message"
)

type client struct {
	addr string
	srv  *rpc.Client
	lock sync.Mutex
}

func NewClient(addr string) common.LockStoreIf {
	//return client struct
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

	e := self.srv.Call("LockStore.Acquire", lu, succ)
	for ; e != nil; e = self.srv.Call("LockStore.Acquire", lu, succ) {
		fmt.Println(e)
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}

func (self *client) Release(lu common.LUpair, succ *bool) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("LockStore.Release", lu, succ)
	for ; e != nil; e = self.srv.Call("LockStore.Release", lu, succ) {
		fmt.Println(e)
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}

/*
func (self *client) ListEntry(lname string, uname *string) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("LockStore.ListEntry", lname, uname)
	for ; e != nil; e = self.srv.Call("LockStore.ListEntry", lname, uname) {
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}
*/

func (self *client) ListQueue(lname string, cList *common.List) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("LockStore.ListQueue", lname, cList)
	for ; e != nil; e = self.srv.Call("LockStore.ListQueue", lname, cList) {
		fmt.Println(e)
		if e = self.connect(false); e != nil {
			return e
		}
	}

	return nil
}
