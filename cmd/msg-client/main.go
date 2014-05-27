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
	send msg: m Head Body;\n`

func runCmd(s common.MessageIf, args []string) bool {
	var reply common.Content

	if len(args)==1 && args[0]== "help" {
		fmt.Println(help)
		return true
	}

	if len(args) < 3 {
		fmt.Println("bad command, try \"help\".")
		return false
	}

	cmd := args[0]
	ctnt := common.Content{args[1], args[2]}
	//h := args[1]
	//b := args[2]
	switch cmd {
	case "m":
		logError(s.Msg(ctnt, &reply))
		fmt.Println(reply)
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
