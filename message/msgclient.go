package message

import (
	//"fmt"
	"net/rpc"
	"sherlock/common"
	"sync"
)

var _ common.MessageIf = new(MsgClient)

type MsgClient struct {
	// private fields
	addr string
	srv  *rpc.Client
	lock sync.Mutex
}

func NewMsgClient(addr string) common.MessageIf {
	return &MsgClient{addr: addr}
}

func (self *MsgClient) connect(firsttime bool) error {
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

func (self *MsgClient) Msg(msg string, succ *bool) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("MsgListener.Msg", msg, succ)
	for ; e != nil; e = self.srv.Call("MsgListener.Msg", msg, succ) {
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}
