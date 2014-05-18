package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"sherlock/common"
	"sherlock/message"
)

func logError(e error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

const help = `Usage:\n
	send msg: m lockname;\n`

func runCmd(s common.MessageIf, args []string) bool {
	var succ bool

	if len(args)==1 && args[0]== "help" {
		fmt.Println(help)
		return true
	}

	if len(args) < 2 {
		fmt.Println("bad command, try \"help\".")
		return false
	}

	cmd := args[0]
	msg := args[1]
	switch cmd {
	case "m":
		logError(s.Msg(msg, &succ))
		fmt.Println(succ)
	default:
		logError(fmt.Errorf("bad command, try \"help\"."))
	}
	return false
}

func runPrompt(s common.MessageIf) {
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
		fmt.Fprintln(os.Stderr, help)
		os.Exit(1)
	}

	addr := args[0]
	c := message.NewMsgClient(addr)

	runPrompt(c)
}
