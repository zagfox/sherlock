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
	lu := common.LUpair{Lockname: "l1", Username: "alice"}
	var succ bool

	ne(s.Acquire(lu, &succ))
	as(succ == true)

	ne(s.Acquire(lu, &succ))
	as(succ == false)

	ne(s.Release(lu, &succ))
	as(succ == true)

	ne(s.Release(lu, &succ))
	as(succ == false)
}
