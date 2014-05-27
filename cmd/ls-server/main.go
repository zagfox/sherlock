package main

import (
	"flag"
	"fmt"
	"log"
	"sherlock/backend"
	"sherlock/common"
	"sherlock/message"
	"sherlock/lockstore"
	"strconv"
)

var (
	frc = flag.String("rc", common.DefaultRCPath, "config rc file path")
	//addr = flag.String("addr", "localhost:23455", "server listen address")
)

func runSrv(rc *common.RC, i int) {
	// Create backconfig

	// talks to srvs
	srvs := make([]common.MessageIf, len(rc.SrvMsgPorts))
	for i, saddr := range rc.SrvMsgPorts {
		srvs[i] = message.NewMsgClient(saddr)
	}

	// data  store and lock store
	ds := lockstore.NewDataStore()
	ls := lockstore.NewLockStore(i, ds, srvs)

	ChCtnt := make(chan common.Content, 1000)
	//ready := make(chan bool)
	backconfig := common.BackConfig{
		Addr:      rc.SrvPorts[i],
		Laddr:     rc.SrvMsgPorts[i],
		Peers:     rc.SrvMsgPorts,
		DataStore: ds,
		LockStore: ls,
		Srvs:      srvs,
		ChCtnt:    ChCtnt,
		Ready:     nil,
	}

	// start a back-end
	fmt.Println(rc.SrvPorts[i])
	/*e := backend.ServeBack(&backconfig)
	if e != nil {
		log.Fatal(e)
	}*/
	server := backend.NewLockServer(&backconfig)
	server.Start()
}

func main() {
	// Parse input addr
	rc, _ := common.LoadRC(*frc)

	flag.Parse()
	args := flag.Args()

	fmt.Println(len(args))
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
