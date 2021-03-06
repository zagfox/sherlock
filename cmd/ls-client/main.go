package main

import (
	"fmt"
	"log"
	"bufio"
	"flag"
	"os"
	"strings"
	"strconv"
	"errors"

	"sherlock/common"
	"sherlock/frontend"
)

func logError(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		log.Fatal(e)
	}
}

func checkargs(args []string, n int) error {
	if len(args) < n {
		return errors.New("args error")
	} else {
		return nil
	}
}

const flaghelp = `Usage:
	client config id`
const help = `Usage:\n
	acquire lock: a lockname;\n
	release lock: r lockname;\n
	show queue:   q lockname;\n`

var (
	frc = flag.String("rc", common.DefaultRCPath, "config file")
)

func runCmd(s common.LockStoreIf, args []string) bool {
	var reply common.Content
	var cList common.List

	if len(args) < 1 {
		fmt.Println("bad command, try \"help\".")
		return false
	}

	cmd := args[0]
	switch cmd {
	case "a":
		logError(checkargs(args, 2))
		lu := common.LUpair{Lockname: args[1], Username: "default"}
		logError(s.Acquire(lu, &reply))
		fmt.Println(reply)
	case "r":
		logError(checkargs(args, 2))
		lu := common.LUpair{Lockname: args[1], Username: "default"}
		logError(s.Release(lu, &reply))
		fmt.Println(reply)
	case "q":
		logError(checkargs(args, 2))
		logError(s.ListQueue(args[1], &cList))
		fmt.Println(cList)
	case "l":
		logError(s.ListLock("default", &cList))
		fmt.Println(cList)
	default:
		logError(fmt.Errorf("bad command, try \"help\"."))
	}
	return false
}

func runPrompt(s common.LockStoreIf) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")

	for scanner.Scan() {
		line := scanner.Text()
		args := strings.Fields(line)
		if len(args) > 0 {
			if runCmd(s, args) {
				break
			}
		}
		fmt.Print("> ")
	}

	e := scanner.Err()
	if e != nil {
		panic(e)
	}
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, flaghelp)
		os.Exit(1)
	}
	cid, e := strconv.Atoi(args[0])
	if e != nil {
		log.Fatal(e)
	}
	// Load rc
	var rc *common.RC
	rc, e = common.LoadRC(*frc)
	if e != nil {
		log.Fatal(e)
	}

	saddrs := rc.SrvPorts
	laddr := rc.CltMsgPorts[cid]
	c := frontend.NewLockClient(saddrs, laddr)

	runPrompt(c)
}
