package backend

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	//"sync"
	"log"
	"time"
	"math/rand"

	"sherlock/common"
	"sherlock/lockstore"
	"sherlock/message"
	"sherlock/paxos"
)

type LockServer struct {
	bc      *common.BackConfig
	srvs []common.MessageIf // method to talk to other servers

	srvView *paxos.ServerView // structure to store self: server Id, masterid, state

	ds   common.DataStoreIf // underlying data store with lock map and log
	ls   common.LockStoreIf // lock rpc entry
	lg	 *lockstore.LogPlayer

	clts *message.MsgClientFactory
}

func NewLockServer(bc *common.BackConfig) *LockServer {
	// srvs talk to all other servers
	srvs := make([]common.MessageIf, len(bc.Peers))
	for i, saddr := range bc.Peers {
		srvs[i] = message.NewMsgClient(saddr)
	}

	// server info
	srvView := paxos.NewServerView(bc.Id, len(bc.Peers), 0, common.SrvReady, srvs)

	clts := message.NewMsgClientFactory()
	// data  store and lock store, and log store
	ds := lockstore.NewDataStore()
	lg := lockstore.NewLogPlayer(ds, srvView, clts)
	ls := lockstore.NewLockStore(srvView, srvs, ds, lg)

	return &LockServer{
		bc: bc, srvView: srvView,
		srvs: srvs, ds: ds, ls: ls, lg: lg,
		clts: clts,
	}
}

/*
 *Start several threads
 */
func (self *LockServer) Start() {
	// Start a thread to listen and handle messages
	self.startMsgListener()

	// Start lock service rpc
	go self.startLockService()
	var ok bool
	ok = <-self.bc.Ready
	if !ok {
		log.Fatal("start lock service error")
	}

	// Start a thread to reply log
	go self.startLogPlayer()

	go self.startViewChecker()

	go self.startLeaseChecker()

	/*
		if self.bc.Ready != nil {
			self.bc.Ready <- true
		}*/

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
	go http.Serve(l, rpcServer)
	if b.Ready != nil {
		b.Ready <- true
	}
}

// start msg listener
func (self *LockServer) startMsgListener() {
	b := self.bc
	// Start msg listener, it is an rpc server
	msghandler := NewServerMsgHandler(self.srvView, self.ds, self.lg)
	msglistener := message.NewMsgListener(msghandler)
	ready := make(chan bool, 1)

	msgconfig := common.MsgConfig{
		Addr:        b.Laddr,
		MsgListener: msglistener,
		Ready:       ready,
	}
	fmt.Println("start msg listener", b.Laddr)

	// no error handling here
	go message.ServeBack(&msgconfig)

	var ok bool
	ok = <-msgconfig.Ready
	if !ok {
		log.Fatal("start msgListener error")
	}
}

// Thread that reply log
func (self *LockServer) startLogPlayer() {
	self.lg.Serve()
}

// Thread that chek if client is alive
func (self *LockServer) startLeaseChecker() {
	fmt.Println("In lease checker")

	var ctnt, reply common.Content
	mTimeout := make(map[string]bool)  // record if client is in timeout

	// check if clients are alive periodically
	for ;;time.Sleep(common.LeasePeriod) {
		if self.bc.Id != self.srvView.GetMasterId() {
			continue
		}

		// get current locks, and iterate through them
		curUser := self.ds.GetAllUser()
		for _, uname := range(curUser) {
			sender := self.clts.GetMsgClient(uname)
			err := sender.Msg(ctnt, &reply)
			if err == nil {
				continue
			}
			// has error, need to handle
			_, ok := mTimeout[uname]
			if !ok {
				// mark it as in timeout phase
				fmt.Println("lockserver", self.bc.Id, curUser, "execeed lease")
				mTimeout[uname] = true
			} else {
				// take back all its lock
				fmt.Println("lockserver", self.bc.Id, "timeout take lock back from ", curUser)
				locks := self.ds.GetUserLock(uname)
				fmt.Println(locks)
				for _, l := range(locks) {
					self.ls.Release(common.LUpair{Lockname:l, Username:uname}, &reply)
				}
				fmt.Println("lock all released")
			}
		}
	}
}

// Thread that start view checker, would issue paxos
func (self *LockServer) startViewChecker() {
	// fmt.Println("In view checker")
	var ctnt, reply common.Content
	for {
		// check if the view needs update`
		needUpdate := false
		for i, srv := range self.srvs {
			e := srv.Msg(ctnt, &reply)
			if e != nil {
				if self.srvView.NodeInView(i) {
					needUpdate = true
					//self.srvView.RequestDelNode(i)
				}
			} else {
				if !self.srvView.NodeInView(i) {
					needUpdate = true
					//self.srvView.RequestAddNode(i)
				}
			}
		}
		if needUpdate {
			// update using paxos
			//vid, view := self.srvView.GetView()
			//fmt.Println("HeartBeat", self.srvView.Id, ": request update view-> vid =", vid, " view =", view)
			self.srvView.RequestUpdateView()
			//fmt.Println("HeartBeat", self.srvView.Id, ": updateview complete")
			fmt.Println()
		}
		// sleep 1000+rand(200)ms
		rand.Seed(time.Now().UTC().UnixNano())
		time.Sleep(time.Millisecond*time.Duration(1000+(rand.Int()%500)))

	}
}


