package lockstore

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"sync"
	"time"
	"strconv"

	"sherlock/common"
	"sherlock/message"
)

var _ common.LockStoreIf = new(lockserver)

type lockserver struct {
	bc *common.BackConfig

	srvs []common.MessageIf // method to talk to other servers
	ds   common.DataStoreIf // underlying data store with lock map and log
	ls   common.LockStoreIf // the underlying storage it should use

	chCtnt    chan common.Content  // channel for passing from msg listener to handler

	Id      int //self id
	mid     int //master server id
	midLock sync.Mutex

	state     string //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex
}

func NewLockServer(bc *common.BackConfig) *lockserver {
	// talks to srvs
	srvs := make([]common.MessageIf, len(bc.Peers))
	for i, saddr := range bc.Peers {
		srvs[i] = message.NewMsgClient(saddr)
	}

	// data  store and lock store
	ds := NewDataStore()
	ls := NewLockStore(bc.Id, ds, srvs)

	// channel for passing from msg listener to handler
	chCtnt := make(chan common.Content, 1000)

	return &lockserver{
		bc: bc, srvs: srvs, ds: ds, ls: ls,
		chCtnt: chCtnt,
		Id: bc.Id, mid: 0, state: "ready",
	}
}

// master mid modify interface
func (self *lockserver) setMasterId(mid int) {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	self.mid = mid
}

func (self *lockserver) getMasterId() int {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	return self.mid
}

// function to set lockserver state
func (self *lockserver) getState() string {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	return self.state
}

func (self *lockserver) setState(state string) {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	self.state = state
}

// Three interface for rpc
func (self *lockserver) Acquire(lu common.LUpair, reply *common.Content) error {
	//first check if self is master
	mid := self.getMasterId()
	if self.Id != mid {
		reply.Head = "NotMaster"
		reply.Body = strconv.FormatUint(uint64(mid), 10)
		return nil
	}
	return self.ls.Acquire(lu, reply)
}

func (self *lockserver) Release(lu common.LUpair, reply *common.Content) error {
	//first check if self is master
	mid := self.getMasterId()
	if self.Id != mid {
		reply.Head = "NotMaster"
		reply.Body = strconv.FormatUint(uint64(mid), 10)
		return nil
	}
	return self.ls.Release(lu, reply)
}

func (self *lockserver) ListQueue(lname string, cList *common.List) error {
	return self.ls.ListQueue(lname, cList)
}

func (self *lockserver) Start() {
	// Start lock service rpc
	go self.startLockService()

	// Start a thread to listen and handle messages
	go self.startMsgListener()
	go self.startMsgHandler()

	// Start a thread to reply log
	go self.startLogPlayer()

	if self.bc.Ready != nil {
		self.bc.Ready <- true
	}

	select {}
}

/*func ServeBack(b *common.BackConfig) error {
	// Start lock service rpc
	go startLockService(b)

	// Start a thread to listen and handle messages
	go startMsgListener(b)
	go startMsgHandler(b)

	// Start a thread to reply log
	go startLogPlayer(b)

	if b.Ready != nil {
		b.Ready <- true
	}

	select {}
}*/

func (self *lockserver) startLockService() {
	b := self.bc
	// listen to address
	l, e := net.Listen("tcp", b.Addr)
	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
	}

	rpcServer := rpc.NewServer()
	//e = rpcServer.Register(b.LockStore)
	e = rpcServer.Register(self.ls)

	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
	}

	// Start service, this is blocking
	http.Serve(l, rpcServer)
}

// start msg listener
func (self *lockserver) startMsgListener() {
	b := self.bc
	// Start msg listener, it is an rpc server
	msglistener := message.NewMsgListener(self.chCtnt)

	msgconfig := common.MsgConfig{
		Addr:        b.Laddr,
		MsgListener: msglistener,
		Ready:       nil,
	}
	fmt.Println("start msg listener", b.Laddr)

	//no error handling here
	message.ServeBack(&msgconfig)
}

// Start msg handler, it reads message from channel
func (self *lockserver) startMsgHandler() {
	b := self.bc
	for {
		// Read event string from channel
		ctnt := <-self.chCtnt
		fmt.Println(b.Addr, ctnt)

		// Examine the content
		switch ctnt.Head {
		}
	}
}

// Thread that reply log
func (self *lockserver) startLogPlayer() {
	for {
	/*	msg := common.Content{"come on", "msg from log player"}
			var reply common.Content

			srv := self.srvs[3]
		    srv.Msg(msg, &reply)
			*/

		time.Sleep(time.Second)
	}
}
