// wrapped client that could send rpc call,
// Wait for event, used directly by user
package frontend

import (
	"fmt"
	"sync"
	"time"
	//"encoding/json"
	"sherlock/common"
	"sherlock/message"
)

// struct that used by user
type lockclient struct {
	saddrs []string            //addr of its server
	mid    int		           //id of master server
	midLock sync.Mutex          //lock for sid
	clts   []common.LockStoreIf //client to call lock rpc

	laddr string			//addr for client listening
	lpid int                //pid for listen thread

	//TODO, give channel to every lock
	acqOk chan string       //Chan to send acquire ok
}

func NewLockClient(saddrs []string, laddr string) common.LockStoreIf {
	// create clt that connects to all servers
	clts := make([]common.LockStoreIf, len(saddrs))
	for i, saddr := range(saddrs) {
		clts[i] = NewClient(saddr)
	}

	// acqOK channel, waiting for event of lock release
	acqOk := make(chan string, common.ChSize)

	// Create lockclient
	lc := lockclient{saddrs:saddrs, mid:0, clts:clts, laddr:laddr, acqOk:acqOk}

	//Start msg listener and handler
	lc.startMsgListener()
	//go lc.startMsgHandler()

	return &lc
}

func (self *lockclient) startMsgListener() {
	// Start msg listener, it is an rpc server
	msghandler := NewClientMsgHandler(self.laddr, self.acqOk)
	msglistener := message.NewMsgListener(msghandler)

	msgconfig := common.MsgConfig{
		Addr:        self.laddr,
		MsgListener: msglistener,
		Ready:       nil,
	}
	fmt.Println("start msg listener", self.laddr)

	//no error handling here
	go message.ServeBack(&msgconfig)
}

// Lock related function
func (self *lockclient) getMid() int {
	self.midLock.Lock()
	defer self.midLock.Unlock()
	return self.mid
}

func (self *lockclient) setMid(mid int) {
	self.midLock.Lock()
	defer self.midLock.Unlock()
	self.mid = mid % len(self.saddrs)
}

// Acquire and Release
func (self *lockclient) Acquire(lu common.LUpair, reply *common.Content) error {
	//set lu username
	lu.Username = self.laddr

	// find a machine that could be connected 
	mid := self.getMid()
	err := self.clts[mid].Acquire(lu, reply)
	for ; err != nil;  {
		mid = self.getMid()
		err = self.clts[mid].Acquire(lu, reply)

		if err != nil{
			self.setMid(mid+1)
			continue
		}

		if reply.Head == "NotReady" {
			fmt.Println("NotReady")
			time.Sleep(time.Second)
			continue
		}
		break
	}

	// handle reply Head
	switch reply.Head {
	// acquire success
	case "LockAcquired":

	//block and wait for set free
	case "LockQueuing":
		<-self.acqOk
		reply.Head = "LockAcquiredByEvent"
	default:
	}
	return nil
}

func (self *lockclient) Release(lu common.LUpair, reply *common.Content) error {
	//set lu username
	lu.Username = self.laddr

	// find a machine that could be connected 
	mid := self.getMid()
	err := self.clts[mid].Release(lu, reply)
	for ; err != nil; self.setMid(mid+1) {
		mid = self.getMid()
		err = self.clts[mid].Release(lu, reply)
		if err == nil {break}
	}

	// handle reply Head
	switch reply.Head {
	default:
	}

	return nil
}

func (self *lockclient) ListQueue(lname string, cList *common.List) error {
	mid := self.getMid()
	return self.clts[mid].ListQueue(lname, cList)
}

var _ common.LockStoreIf = new(lockclient)
