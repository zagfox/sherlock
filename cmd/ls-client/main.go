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

const flaghelp = `Usage:
	server addr, listen addr`
const help = `Usage:\n
	acquire lock: a lockname;\n
	release lock: r lockname;\n
	show queue:   q lockname;\n`

func runCmd(s common.LockStoreIf, args []string) bool {
	var succ bool
	//var str string
	var cList common.List

	if len(args) < 2 {
		fmt.Println("bad command, try \"help\".")
		return false
	}

	cmd := args[0]
	lu := common.LUpair{Lockname: args[1], Username: "default"}
	switch cmd {
	case "a":
		logError(s.Acquire(lu, &succ))
		fmt.Println(succ)
	case "r":
		logError(s.Release(lu, &succ))
		fmt.Println(succ)
	/*
	case "e":
		logError(s.ListEntry(args[1], &str))
		fmt.Println(str)
	*/
	case "q":
		logError(s.ListQueue(args[1], &cList))
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
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, flaghelp)
		os.Exit(1)
	}

	saddr := args[0]
	laddr := args[1]
	c := frontend.NewLockClient(saddr, laddr)

	runPrompt(c)
}
