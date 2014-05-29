//wrapped client that could send rpc call,
// Wait for event, used directly by user
package frontend

import (
	"fmt"
	"sync"
	"time"
	"encoding/json"
	"sherlock/common"
	"sherlock/message"
)

// struct that used by user
type lockclient struct {
	saddrs []string            //addr of its server
	sid    int		           //id of master server
	sidLock sync.Mutex          //lock for sid
	clts   []common.LockStoreIf //client to call lock rpc

	ch chan common.Content      //chan for listen
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

	// channel for listening all message
	ch := make(chan common.Content, 1000)

	// acqOK channel, waiting for event of lock release
	acqOk := make(chan string, 1000)

	// Create lockclient
	lc := lockclient{saddrs:saddrs, sid:0, clts:clts, laddr:laddr, ch:ch, acqOk:acqOk}

	//Start msg listener and handler
	lc.startMsgListener()
	go lc.startMsgHandler()

	return &lc
}

func (self *lockclient) startMsgListener() {
	// Start msg listener, it is an rpc server
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
		ctnt := <-self.ch
		fmt.Println(ctnt)

		// Examine the content
		switch ctnt.Head {
		case "acqOk":
			var lu common.LUpair
			json.Unmarshal([]byte(ctnt.Body), &lu)
			if lu.Username == self.laddr {
				self.acqOk<- "true"
			}
		}
		/*
		//unmarshall it and handle it
		var event common.Event
		json.Unmarshal([]byte(bytes), &event)

		// if event is acqSucc and parameter correct
		if event.Name == "acqOk" && event.Username == self.laddr {
			self.acqOk<- bytes
		}
		*/
	}
}

// Lock related function
func (self *lockclient) getSid() int {
	self.sidLock.Lock()
	defer self.sidLock.Unlock()
	return self.sid
}

func (self *lockclient) setSid(sid int) {
	self.sidLock.Lock()
	defer self.sidLock.Unlock()
	self.sid = sid % len(self.saddrs)
}

// Acquire and Release
func (self *lockclient) Acquire(lu common.LUpair, reply *common.Content) error {
	fmt.Println("lockclient", "acquire")
	//set lu username
	lu.Username = self.laddr

	// find a machine that could be connected 
	sid := self.getSid()
	err := self.clts[sid].Acquire(lu, reply)
	for ; err != nil;  {
		sid = self.getSid()
		err = self.clts[sid].Acquire(lu, reply)

		if err != nil{
			self.setSid(sid+1)
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
	sid := self.getSid()
	err := self.clts[sid].Release(lu, reply)
	for ; err != nil; self.setSid(sid+1) {
		sid = self.getSid()
		err = self.clts[sid].Release(lu, reply)
		if err == nil {break}
	}

	// handle reply Head
	switch reply.Head {
	default:
	}

	return nil
}

func (self *lockclient) ListQueue(lname string, cList *common.List) error {
	sid := self.getSid()
	return self.clts[sid].ListQueue(lname, cList)
}

var _ common.LockStoreIf = new(lockclient)
