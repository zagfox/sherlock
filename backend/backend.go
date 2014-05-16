package backend

import (
	//"fmt"
	"net"
	"net/http"
	"net/rpc"

	"sherlock/common"
)

func ServeBack(b *common.BackConfig) error {
	// listen to address
	l, e := net.Listen("tcp", b.Addr)
	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
		return e
	}

	rpcServer := rpc.NewServer()
	e = rpcServer.Register(b.Store)

	if e != nil {
		if b.Ready != nil {
			b.Ready <- false
		}
		return e
	}

	if b.Ready != nil {
		b.Ready <- true
	}
	// Start service, this is blocking
	return http.Serve(l, rpcServer)
}
