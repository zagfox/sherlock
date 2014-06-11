package test

import (
	"os/exec"
	"testing"
	"strconv"
	//"strings"
	//"bytes"
	"log"
	"math/rand"
	"time"
)

func aliveNum(alive []int) int {
	ret := 0
	for i := 0; i < 5; i++ {
		if alive[i] == 1 {
			ret++
		}
	}
	return ret
}

func TestLockServerBasic(t *testing.T) {
	var err error
	//var out bytes.Buffer

	alive := make([]int, 5)
	//Start Backs
	bin_path := "/classes/cse223b/sp14/cse223f7zhu/gopath/bin/"
	servers := make([]*exec.Cmd, 10)
	for i := 0; i < 5; i++ {
		servers[i] = exec.Command(bin_path+"ls-server", strconv.FormatInt((int64)(i), 10))
		/*if i == 4 {
			servers[i].Stdout = &out
		}*/
		err = servers[i].Start()
		if err != nil {
			log.Fatal(err)
		}
		alive[i] = 1
	}

	// begin revert state
	var id int
	rand.Seed(time.Now().UTC().UnixNano())
	for {
		id = rand.Int() % 5
		if alive[id] == 1{
			if aliveNum(alive) > 2 {
				servers[id].Process.Kill()
				alive[id] = 0
			} else {
				continue
			}
		} else {
			servers[id] = exec.Command(bin_path+"ls-server", strconv.FormatInt((int64)(id), 10))
			servers[id].Start()
			alive[id] = 1
		}
		log.Println(alive)
		time.Sleep(5*time.Second)

	}

	//time.Sleep(time.Second)
	//log.Println(servers[0].Process)
	for i := 0; i < 5; i++ {
		servers[i].Process.Kill()
	}
}
