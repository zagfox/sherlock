package backend

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	//"sync"
	"time"

	"sherlock/common"
	"sherlock/message"
	"sherlock/lockstore"
)

type LockServer struct {
	bc *common.BackConfig
	srvInfo   *lockstore.ServerInfo  // structure to store self: server Id, masterid, state

	srvs []common.MessageIf // method to talk to other servers
	ds   common.DataStoreIf // underlying data store with lock map and log
	ls   common.LockStoreIf // the underlying storage it should use

	chCtnt    chan common.Content  // channel for passing from msg listener to handler

/*
	Id      int //self id
	mid     int //master server id
	midLock sync.Mutex

	state     string //indicate if server state: updateview, transdata, ready
	lockState sync.Mutex
	*/
}

func NewLockServer(bc *common.BackConfig) *LockServer {
	// server info
	srvInfo := lockstore.NewServerInfo(bc.Id, 0, "ready")

	// talks to srvs
	srvs := make([]common.MessageIf, len(bc.Peers))
	for i, saddr := range bc.Peers {
		srvs[i] = message.NewMsgClient(saddr)
	}

	// data  store and lock store
	ds := lockstore.NewDataStore()
	ls := lockstore.NewLockStore(srvInfo, srvs, ds)

	// channel for passing from msg listener to handler
	chCtnt := make(chan common.Content, 1000)

	return &LockServer{
		bc: bc, srvInfo: srvInfo,
		srvs: srvs, ds: ds, ls: ls,
		chCtnt: chCtnt,
		//Id: bc.Id, mid: 0, state: "updateview",
	}
}

/*
// master mid modify interface
func (self *LockServer) setMasterId(mid int) {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	self.mid = mid
}

func (self *LockServer) getMasterId() int {
	self.midLock.Lock()
	defer self.midLock.Unlock()

	return self.mid
}

// function to set lockserver state
func (self *LockServer) getState() string {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	return self.state
}

func (self *LockServer) setState(state string) {
	self.lockState.Lock()
	defer self.lockState.Unlock()
	self.state = state
}
*/


/*
// Three interface for rpc
func (self *LockServer) Acquire(lu common.LUpair, reply *common.Content) error {
	// check if server is ready
	state := self.srvInfo.GetState()
	fmt.Println("lockserver", state)
	if state != "ready" {
		reply.Head = "NotReady"
		return nil
	}

	// check if self is master
	mid := self.srvInfo.GetMasterId()
	if self.Id != mid {
		reply.Head = "NotMaster"
		reply.Body = strconv.FormatUint(uint64(mid), 10)
		return nil
	}

	return self.ls.Acquire(lu, reply)
}

func (self *LockServer) Release(lu common.LUpair, reply *common.Content) error {
	// check if server is ready
	state := self.srvInfo.GetState()
	if state != "ready" {
		reply.Head = "NotReady"
		return nil
	}

	// check if self is master
	mid := self.srvInfo.GetMasterId()
	if self.Id != mid {
		reply.Head = "NotMaster"
		reply.Body = strconv.FormatUint(uint64(mid), 10)
		return nil
	}

	return self.ls.Release(lu, reply)
}

func (self *LockServer) ListQueue(lname string, cList *common.List) error {
	return self.ls.ListQueue(lname, cList)
}
*/

/*
 *Start several threads
 */
func (self *LockServer) Start() {
	// Start a thread to listen and handle messages
	go self.startMsgListener()
	go self.startMsgHandler()

	// Start a thread to reply log
	go self.startLogPlayer()

	// Start lock service rpc
	go self.startLockService()

	if self.bc.Ready != nil {
		self.bc.Ready <- true
	}

	select {}
}

func (self *LockServer) startLockService() {
	b := self.bc
	// listen to address
	l, e := net.Listen("tcp", b.Addr)
	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
	}

	rpcServer := rpc.NewServer()
	e = rpcServer.Register(self.ls)
	//e = rpcServer.Register(common.LockStoreIf(self))

	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
	}

	// Start service, this is blocking
	http.Serve(l, rpcServer)
}

// start msg listener
func (self *LockServer) startMsgListener() {
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
func (self *LockServer) startMsgHandler() {
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
func (self *LockServer) startLogPlayer() {
	for {
	/*	msg := common.Content{"come on", "msg from log player"}
			var reply common.Content

			srv := self.srvs[3]
		    srv.Msg(msg, &reply)
			*/

		time.Sleep(time.Second)
	}
}
