package backend

import (
	"fmt"
	"time"
	"net"
	"net/http"
	"net/rpc"

	"sherlock/common"
	"sherlock/message"
)

type lockserver struct {
	bc *common.BackConfig
}

func NewLockServer(bc *common.BackConfig) *lockserver {
	return &lockserver{bc:bc}
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
	e = rpcServer.Register(b.LockStore)

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
	msglistener := message.NewMsgListener(b.ChCtnt)

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
		ctnt := <-b.ChCtnt
		fmt.Println(b.Addr, ctnt)

		// Examine the content
		switch ctnt.Head {
		}
	}
}

// Thread that reply log
func (self *lockserver) startLogPlayer() {
	for {
		/*msg := common.Content{"come on", "msg from log player"}
		var reply common.Content

		srv := self.bc.Srvs[3]
	    srv.Msg(msg, &reply)
		*/
		time.Sleep(time.Second)
	}
}
