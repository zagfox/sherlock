package test

import (
	"log"
	"runtime/debug"
	"testing"

	"sherlock/common"
	"sherlock/lockstore"
	"sherlock/message"
	"sherlock/paxos"
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

func startLockStore(Id int) common.LockStoreIf {
	rc, _ := common.LoadRC(common.DefaultRCPath)
	bc := common.BackConfig{
		Id:    Id,
		Addr:  rc.SrvPorts[Id],
		Laddr: rc.SrvMsgPorts[Id],
		Peers: rc.SrvMsgPorts,
		Ready: make(chan bool, 10),
	}

	srvs := make([]common.MessageIf, len(bc.Peers))
	for i, saddr := range bc.Peers {
		srvs[i] = message.NewMsgClient(saddr)
	}

	msg := message.NewMsgClientFactory()

	// srvView, num:1, mid:0
	srvView := paxos.NewServerView(bc.Id, 1, 0, "ready", srvs)
	ds := lockstore.NewDataStore()
	lg := lockstore.NewLogPlayer(ds, srvView, msg)
	ls := lockstore.NewLockStore(srvView, srvs, ds, lg)
	return ls
}

func basicTestLockStore(s common.LockStoreIf) {
	// Start testing here
	lu := common.LUpair{Lockname: "l1", Username: "alice"}
	var reply common.Content

	//basic test for one user
	ne(s.Acquire(lu1, &reply))
	as(reply.Head == "LockQueuing")

	/*	// Assume NO Deadlock
		ne(s.Acquire(lu1, &succ))
		as(succ == false)
	*/

	/*ne(s.Release(lu1, &reply))
	as(reply.Head == "LockReleased")

	ne(s.Release(lu1, &reply))
	as(reply.Head == "LockNotFound")

	//test for two user acquire one lock
	lu1 = common.LUpair{Lockname: "l1", Username: "alice"}
	lu2 = common.LUpair{Lockname: "l1", Username: "bob"}
	ne(s.Acquire(lu1, &reply))
	ne(s.Acquire(lu2, &reply))
	as(reply.Head == "LockQueuing")

	//check queue
	ne(s.ListQueue("l1", &cList))
	as(len(cList.L) == 2)
	as(cList.L[0] == "alice")
	as(cList.L[1] == "bob")

	// one of them release it
	ne(s.Release(lu1, &reply))
	ne(s.ListQueue("l1", &cList))
	as(len(cList.L) == 1)
	as(cList.L[0] == "bob")

	ne(s.ListQueue("l2", &cList))
	as(len(cList.L) == 0)
	*/
}

func TestLockStore(t *testing.T) {
	ls := startLockStore(0)
	basicTestLockStore(ls)
}
