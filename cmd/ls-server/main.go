package main

import (
	"fmt"
	"flag"
	"log"
	"sherlock/lockstore"
	"sherlock/common"
	"sherlock/backend"
)

var (
    addr = flag.String("addr", "localhost:23455", "server listen address")
)

func main() {
	// Parse input addr
	flag.Parse()

	// Create backconfig
	s := lockstore.NewLockStore()
	//ready := make(chan bool)
	backconfig := common.BackConfig{
		Addr:     *addr,
		LockStore:    s,
		Ready:    nil,
	}

	// start a back-end
	fmt.Println(*addr)
	e := backend.ServeBack(&backconfig)
	if e != nil {
		log.Fatal(e)
	}
}
