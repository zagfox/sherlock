package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"sherlock/common"
	"sherlock/frontend"
)

func logError(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

const help = "Usage:"

func runCmd(uname *string, s common.LockStoreIf, args []string) bool {
	var succ bool

	if len(args) < 2 {
		fmt.Println("bad command, try \"help\".")
		return false
	}

	cmd := args[0]
	lu := common.LUpair{Lockname: args[1], Username: *uname}
	switch cmd {
	case "a":
		logError(s.Acquire(lu, &succ))
		fmt.Println(succ)
	case "r":
		logError(s.Release(lu, &succ))
		fmt.Println(succ)
	default:
		logError(fmt.Errorf("bad command, try \"help\"."))
	}
	return false
}

func runPrompt(s common.LockStoreIf) {
	var uname string
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
