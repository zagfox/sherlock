// Wait for event, used directly by user
package frontend

import (
	"fmt"
	"strconv"
	"sync"
	"time"
	//"encoding/json"
	"sherlock/common"
	"sherlock/message"
)

// struct that used by user
type lockclient struct {
	saddrs  []string             //addr of its server
	mid     int                  //id of master server
	midLock sync.Mutex           //lock for sid
	clts    []common.LockStoreIf //client to call lock rpc

	laddr string //addr for client listening
	lpid  int    //pid for listen thread
	mAcqChan map[common.LUpair]chan string //Chan to receive acq, release events from server

	mlocks map[string]bool    // map to identify what lock it has
}

func NewLockClient(saddrs []string, laddr string) common.LockStoreIf {
	// create clt that connects to all servers
	clts := make([]common.LockStoreIf, len(saddrs))
	for i, saddr := range saddrs {
		clts[i] = NewClient(saddr)
	}

	// acqOK channel, waiting for event of lock release
	mAcqChan := make(map[common.LUpair]chan string)

	mlocks := make(map[string]bool)

	// Create lockclient
	lc := lockclient{saddrs: saddrs, mid: 0, clts: clts, laddr: laddr, mAcqChan: mAcqChan, mlocks:mlocks}

	//Start msg listener and handler
	lc.startMsgListener()
	//go lc.startMsgHandler()

	return &lc
}

func (self *lockclient) startMsgListener() {
	// Start msg listener, it is an rpc server
	msghandler := NewClientMsgHandler(self.laddr, self.mAcqChan)
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
	//set lu username, kind of tricky here...
	lu.Username = self.laddr

	// create channel for this lock
	_, ok := self.mAcqChan[lu]
	if !ok {
		self.mAcqChan[lu] = make(chan string, common.ChSize)
	}

	// find a machine that could be connected
	var mid int
	var err error
	for {
		mid = self.getMid()
		err = self.clts[mid].Acquire(lu, reply)

		if err != nil {
			fmt.Println("mid=", mid, "network error, change mid")
			self.setMid(mid + 1)
			continue
		}

		fmt.Println("mid=", mid, "reply=", reply)
		if reply.Head == "NotReady" {
			time.Sleep(time.Second)
			continue
		}
		if reply.Head == "NotMaster" {
			mid, _ = strconv.Atoi(reply.Body)
			self.setMid(mid)
			continue
		}

		break
	}

	// handle reply Head
	switch reply.Head {
	case "LockQueuing":
		// in normal case, request goes to log, return this
		ch, _ := self.mAcqChan[lu]
		<-ch // channel must have been set up
		reply.Head = "LockAcquiredByEvent"

		//mark that lock is acquired
		self.mlocks[lu.Lockname] = true
	default:
	}
	return nil
}

func (self *lockclient) Release(lu common.LUpair, reply *common.Content) error {
	//set lu username
	lu.Username = self.laddr

	// find a machine that could be connected
	var mid int
	var err error
	for {
		mid = self.getMid()
		err = self.clts[mid].Release(lu, reply)

		if err != nil {
			self.setMid(mid + 1)
			continue
		}

		if reply.Head == "NotReady" {
			time.Sleep(time.Second)
			continue
		}
		if reply.Head == "NotMaster" {
			mid, _ = strconv.Atoi(reply.Body)
			self.setMid(mid)
			continue
		}

		break
	}

	// handle reply Head
	switch reply.Head {
	default:
		delete(self.mlocks, lu.Lockname)
	}

	return nil
}

func (self *lockclient) ListQueue(lname string, cList *common.List) error {
	// find a machine that could be connected
	var mid int
	//var err error
	for {
		mid = self.getMid()
		self.clts[mid].ListQueue(lname, cList)

		/*if err != nil {
			self.setMid(mid + 1)
			continue
		}

		if reply.Head == "NotReady" {
			time.Sleep(time.Second)
			continue
		}
		if reply.Head == "NotMaster" {
			mid, _ = strconv.Atoi(reply.Body)
			self.setMid(mid)
			continue
		}*/

		break
	}

	return nil

}

func (self *lockclient) ListLock(uname string, cList *common.List) error {
	//self.ListLocalLock()

	uname = self.laddr

	// find a machine that could be connected
	var mid int
	//var err error
	for {
		mid = self.getMid()
		self.clts[mid].ListLock(uname, cList)

		/*if err != nil {
			self.setMid(mid + 1)
			continue
		}

		if reply.Head == "NotReady" {
			time.Sleep(time.Second)
			continue
		}
		if reply.Head == "NotMaster" {
			mid, _ = strconv.Atoi(reply.Body)
			self.setMid(mid)
			continue
		}*/

		break
	}

	return nil
}
//not an rpc interface, check lock locks
func (self *lockclient) ListLocalLock() error {
	var cList common.List
	cList.L = make([]string, 0)
	for k,_ := range(self.mlocks) {
		cList.L = append(cList.L, k)
	}
	fmt.Println("showLocalLock", cList.L)
	return nil
}

var _ common.LockStoreIf = new(lockclient)
