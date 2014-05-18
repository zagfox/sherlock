package main

import (
	"fmt"
	"flag"
	"log"
	"sherlock/message"
	"sherlock/common"
)

var (
    addr = flag.String("addr", "localhost:27000", "msg server listen address")
)

func main() {
	// Parse input addr
	flag.Parse()

	// Create backconfig
	ch := make(chan string, 1)
	l := message.NewMsgListener(ch)
	//ready := make(chan bool)
	msgconfig := common.MsgConfig{
		Addr:	        *addr,
		MsgListener:    l,
		Ready:          nil,
	}

	// start a back-end
	fmt.Println("msg listener", *addr)
	e := message.ServeBack(&msgconfig)
	if e != nil {
		log.Fatal(e)
	}
}
