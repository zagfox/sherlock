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
	ls   common.LockStoreIf // lock rpc entry

	chCtnt    chan common.Content  // channel for passing from msg listener to handler
}

func NewLockServer(bc *common.BackConfig) *LockServer {
	// server info
	srvInfo := lockstore.NewServerInfo(bc.Id, 0, common.SrvReady)

	// srvs talk to all other servers
	srvs := make([]common.MessageIf, len(bc.Peers))
	for i, saddr := range bc.Peers {
		srvs[i] = message.NewMsgClient(saddr)
	}

	// data  store and lock store
	ds := lockstore.NewDataStore()
	ls := lockstore.NewLockStore(srvInfo, srvs, ds)

	// channel for passing from msg listener to handler
	chCtnt := make(chan common.Content, common.ChSize)

	return &LockServer{
		bc: bc, srvInfo: srvInfo,
		srvs: srvs, ds: ds, ls: ls,
		chCtnt: chCtnt,
	}
}

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
