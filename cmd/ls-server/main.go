package main

import (
	"flag"
	"fmt"
	"log"
	"sherlock/backend"
	"sherlock/common"
	"sherlock/lockstore"
	"strconv"
)

var (
	frc  = flag.String("rc", common.DefaultRCPath, "config rc file path")
	addr = flag.String("addr", "localhost:23455", "server listen address")
)

func runSrv(rc *common.RC, i int) {
		// Create backconfig
		s := lockstore.NewLockStore()
		//ready := make(chan bool)
		backconfig := common.BackConfig{
			Addr:      rc.SrvPorts[i],
			LockStore: s,
			Ready:     nil,
		}

		// start a back-end
		fmt.Println(rc.SrvPorts[i])
		e := backend.ServeBack(&backconfig)
		if e != nil {
			log.Fatal(e)
		}

}

func main() {
	// Parse input addr
	//flag.Parse()
	rc, _ := common.LoadRC(*frc)

	args := flag.Args()

	if len(args) == 0 {
		for i, _ := range rc.SrvPorts {
			go runSrv(rc, i)
		}
	} else {
		for _, a := range args {
			i, e := strconv.Atoi(a)
			if e != nil {
				log.Fatal(e)
			}
			go runSrv(rc, i)
		}
	}

	//wait here
	select {}
}
