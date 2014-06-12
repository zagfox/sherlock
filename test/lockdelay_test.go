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

func TestLockDelay(t *testing.T) {
	//func startLocalServer
	//server := startLocalServer()
	//time.Sleep(2*time.Second)

	c := getClient(0)

	// start testing
	lu := common.LUpair{Lockname: "l1", Username: "default"}
	var reply common.Content

	start := time.Now()
	for i := 0; i < 1000; i++ {
		lu.Lockname = "l"+strconv.Itoa(i)
		c.Acquire(lu, &reply)
	}
	elapsed := time.Since(start)
	log.Println("acquire time is %s", elapsed)
	c.Release(lu, &reply)

	// close local server
	//server.Process.Kill()
}
