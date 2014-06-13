package test

import (
	"log"
	"os/exec"
	"testing"
	"strconv"
	"flag"
	"time"

	"sherlock/common"
	"sherlock/frontend"
)

var (
    frc = flag.String("rc", common.DefaultRCPath, "config file")
)

func startLocalServer() *exec.Cmd {
	//func startLocalServer
	bin_path := "/classes/cse223b/sp14/cse223y4duan/gopath/bin/"
	server := exec.Command(bin_path + "ls-server")
	err := server.Start()
	if err != nil {
		log.Fatal(err)
	}

	return server
}

func getClient(cid int) common.LockStoreIf {
	var rc *common.RC
	var e error

	// Load rc
	rc, e = common.LoadRC(*frc)
	if e != nil {
		log.Fatal(e)
	}

	// create client
	saddrs := rc.SrvPorts
	laddr := rc.CltMsgPorts[cid]
	c := frontend.NewLockClient(saddrs, laddr)

	return c
}

func TestThroughputMultiple(t *testing.T) {
	//func startLocalServer
	//server := startLocalServer()
	//time.Sleep(2*time.Second)

	log.Println("started")
	cli := make([]common.LockStoreIf, 100)
	for i, _ := range cli{
		cli[i] = getClient(i)
	}

	ch := make(chan bool, 1000)
	// start testing

	start := time.Now()
	for j, cl := range cli {
		go func(j int, cl common.LockStoreIf){
			for i := 0; i < 100; i++{
				lu := common.LUpair{}
				var reply common.Content
				lu.Lockname = "l"+strconv.Itoa(j)+"-"+strconv.Itoa(i)
				cl.Acquire(lu, &reply)
			}
			ch <- true
		}(j, cl)
	}
	for i := 0; i < len(cli); i++{
		<-ch
	}
	elapsed := time.Since(start)
	log.Println("acquire time is %s", elapsed)

	for j, cl := range cli {
		go func(j int, cl common.LockStoreIf){
			for i := 0; i < 100; i++{
				lu := common.LUpair{}
				var reply common.Content
				lu.Lockname = "l"+strconv.Itoa(j)+"-"+strconv.Itoa(i)
				cl.Release(lu, &reply)
			}
			ch <- true
		}(j, cl)
	}
	for i := 0; i < len(cli); i++{
		<-ch
	}

	// close local server
	//server.Process.Kill()
}
