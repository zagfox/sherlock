package main

import (
	"flag"
	"fmt"
	"log"
	"sherlock/common"
	//"sherlock/message"
	//"sherlock/lockstore"
	"sherlock/backend"
	"strconv"
)

var (
	frc = flag.String("rc", common.DefaultRCPath, "config rc file path")
)

func runSrv(rc *common.RC, i int) {
	// Create backconfig

	backconfig := common.BackConfig{
		Id:        i,
		Addr:      rc.SrvPorts[i],
		Laddr:     rc.SrvMsgPorts[i],
		Peers:     rc.SrvMsgPorts,
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
