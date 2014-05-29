package test

import (
	"runtime/debug"
	"testing"
	"log"

	"sherlock/common"
	"sherlock/lockstore"
	"sherlock/message"
)

func ne(e error) {
	if e != nil {
		debug.PrintStack()
		log.Fatal(e)
	}
}

func as(cond bool) {
	if !cond {
		debug.PrintStack()
		log.Fatal("assertion failed")
	}
}

func TestLockStore(t *testing.T) {

	rc, _ := common.LoadRC(common.DefaultRCPath)
	Id := 0
	bc := common.BackConfig{
		Id:    Id,
		Addr:  rc.SrvPorts[Id],
		Laddr: rc.SrvMsgPorts[Id],
		Peers: rc.SrvMsgPorts,
		Ready: nil,
	}
	srvView := lockstore.NewServerView(bc.Id, 0, "ready")
	srvs := make([]common.MessageIf, len(bc.Peers))
	for i, saddr := range bc.Peers {
		srvs[i] = message.NewMsgClient(saddr)
	}
	ds := lockstore.NewDataStore()
	s := lockstore.NewLockStore(srvView, srvs, ds)

	// Start testing here
	lu1 := common.LUpair{Lockname: "l1", Username: "alice"}
	lu2 := common.LUpair{Lockname: "l2", Username: "bob"}
	var reply common.Content
	var cList common.List

	//basic test for one user
	ne(s.Acquire(lu1, &reply))
	as(reply.Head == "LockAcquired")

	/*
		// Deadlock, don't know what to do
		ne(s.Acquire(lu1, &succ))
		as(succ == false)
	*/

	ne(s.Release(lu1, &reply))
	as(reply.Head == "LockReleased")

	ne(s.Release(lu1, &reply))
	as(reply.Head == "LockNotFound")

	//test for two user
	lu1 = common.LUpair{Lockname: "l1", Username: "alice"}
	lu2 = common.LUpair{Lockname: "l1", Username: "bob"}
	ne(s.Acquire(lu1, &reply))
	ne(s.Acquire(lu2, &reply))
	//as(succ == false)

	//test for queue
	ne(s.ListQueue("l1", &cList))
	as(len(cList.L) == 2)
	as(cList.L[0] == "alice")
	as(cList.L[1] == "bob")

	ne(s.ListQueue("l2", &cList))
	as(len(cList.L) == 0)
}
