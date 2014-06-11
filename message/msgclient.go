package message

import (
	//"fmt"
	"net/rpc"
	"sherlock/common"
	"sync"
)

type MsgClientFactory struct {
	lock sync.Mutex
	mclients map[string]common.MessageIf
}

func NewMsgClientFactory() *MsgClientFactory {
	return &MsgClientFactory {
		mclients: make(map[string]common.MessageIf),
	}
}

func (self *MsgClientFactory) GetMsgClient(name string) common.MessageIf {
	self.lock.Lock()
	defer self.lock.Unlock()

	if _, ok := self.mclients[name]; !ok {
		self.mclients[name] = NewMsgClient(name)
	}

	return self.mclients[name]
}




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

func (self *MsgClient) Msg(msg common.Content, reply *common.Content) error {
	if e := self.connect(true); e != nil {
		return e
	}

	e := self.srv.Call("MsgListener.Msg", msg, reply)
	for ; e != nil; e = self.srv.Call("MsgListener.Msg", msg, reply) {
		if e = self.connect(false); e != nil {
			return e
		}
	}
	return nil
}
