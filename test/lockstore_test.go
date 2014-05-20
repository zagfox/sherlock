package test

import (
	"testing"
	"runtime/debug"

	"sherlock/lockstore"
	"sherlock/common"
)

func TestLockStore(t *testing.T) {

	ne := func(e error) {
		if e != nil {
			debug.PrintStack()
			t.Fatal(e)
		}
	}

	as := func(cond bool) {
		if !cond {
			debug.PrintStack()
			t.Fatal("assertion failed")
		}
	}

	s := lockstore.NewLockStore()
	lu1 := common.LUpair{Lockname: "l1", Username: "alice"}
	lu2 := common.LUpair{Lockname: "l2", Username: "bob"}
	var succ bool
	var cList common.List

	//basic test for one user
	ne(s.Acquire(lu1, &succ))
	as(succ == true)

	/*
	// Deadlock, don't know what to do
	ne(s.Acquire(lu1, &succ))
	as(succ == false)
	*/

	ne(s.Release(lu1, &succ))
	as(succ == true)

	ne(s.Release(lu1, &succ))
	as(succ == false)

	//test for two user
	lu1 = common.LUpair{Lockname: "l1", Username: "alice"}
	lu2 = common.LUpair{Lockname: "l1", Username: "bob"}
	ne(s.Acquire(lu1, &succ))
	ne(s.Acquire(lu2, &succ))
	as(succ == false)

	//test for queue
	ne(s.ListQueue("l1", &cList))
	as(len(cList.L) == 2)
	as(cList.L[0] == "alice")
	as(cList.L[1] == "bob")

	ne(s.ListQueue("l2", &cList))
	as(len(cList.L) == 0)
}
