//wrapped client that could send rpc call,
// Wait for event, used directly by user
package frontend

import (
	"fmt"
	"sherlock/common"
	"sherlock/message"
)

// struct that used by user
type lockclient struct {
	saddr string             //addr of its server
	clt   common.LockStoreIf //client to call lcok rpc

	ch chan string        //chan for listen
	laddr string			//addr for client listening
	lpid int                //pid for listen thread

	acqOk chan string       //Chan to send acquire ok
}

func NewLockClient(saddr, laddr string) common.LockStoreIf {
	ch := make(chan string, 1000) //make a channel
	clt := NewClient(saddr)
	acqOk := make(chan string, 1000)
	lc := lockclient{saddr:saddr, clt:clt, laddr:laddr, ch:ch, acqOk:acqOk}

	lc.startMsgListener()
	go lc.startMsgHandler()
	return &lc
}

// Start msg listener, it is an rpc server
func (self *lockclient) startMsgListener() {
	msglistener := message.NewMsgListener(self.ch)
	msgconfig := common.MsgConfig{
		Addr:        self.laddr,
		MsgListener: msglistener,
		Ready:       nil,
	}
	fmt.Println("start msg listener", self.laddr)
	//no error handling here
	go message.ServeBack(&msgconfig)
}

// Start msg handler, it reads message from channel
func (self *lockclient) startMsgHandler() {
	for {
		// Read event string from channel
		event := <-self.ch
		fmt.Println(event)
		//first write in this way
		self.acqOk<- event
	}
}

func (self *lockclient) Acquire(lu common.LUpair, succ *bool) error {
	lu.Username = self.laddr
	err := self.clt.Acquire(lu, succ)
	if err != nil {
		return err
	}

	if *succ == true {
		return nil
	}
	//block and wait for set free
	<-self.acqOk
	*succ = true
	return nil
}

func (self *lockclient) Release(lu common.LUpair, succ *bool) error {
	lu.Username = self.laddr
	return self.clt.Release(lu, succ)
}

func (self *lockclient) ListQueue(lname string, cList *common.List) error {
	return self.clt.ListQueue(lname, cList)
}

var _ common.LockStoreIf = new(lockclient)
