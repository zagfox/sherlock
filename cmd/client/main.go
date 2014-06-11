package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"errors"

	"sherlock/common"
	"sherlock/frontend"
)

func logError(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

func checkargs(args []string, n int) error {
	if len(args) < n {
		return errors.New("args error")
	} else {
		return nil
	}
}

const help = `Usage:\n
	acquire lock: a lockname;\n
	release lock: r lockname;\n
	show entry:   e lockname;\n
	show queue:   q lockname;\n`

func runCmd(uname *string, s common.LockStoreIf, args []string) bool {
	var reply common.Content
	//var str string
	var cList common.List

	if len(args) < 1 {
		fmt.Println("bad command, try \"help\".")
		return false
	}

	cmd := args[0]
	lu := common.LUpair{Lockname: args[1], Username: *uname}
	switch cmd {
	case "u":
		*uname = args[1]
		fmt.Println(*uname)
	case "a":
		logError(checkargs(args, 2))
		logError(s.Acquire(lu, &reply))
		fmt.Println(reply)
	case "r":
		logError(checkargs(args, 2))
		logError(s.Release(lu, &reply))
		fmt.Println(reply)
	/*
	case "e":
		logError(s.ListEntry(args[1], &str))
		fmt.Println(str)
	*/
	case "q":
		logError(s.ListQueue(args[1], &cList))
		fmt.Println(cList)
	case "l":
		logError(s.ListLock(args[1], &cList))
		fmt.Println(cList)
	default:
		logError(fmt.Errorf("bad command, try \"help\"."))
	}
	return false
}

func runPrompt(s common.LockStoreIf) {
	uname := "default"
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")

	for scanner.Scan() {
		line := scanner.Text()
		args := strings.Fields(line)
		if len(args) > 0 {
			if runCmd(&uname, s, args) {
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
		fmt.Fprintln(os.Stderr, help)
		os.Exit(1)
	}

	addr := args[0]
	c := frontend.NewClient(addr)

	runPrompt(c)
}
