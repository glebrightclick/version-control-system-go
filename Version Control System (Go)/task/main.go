package main

import (
	"fmt"
	"os"
)

const (
	cHelp     = "--help"
	cConfig   = "config"
	cAdd      = "add"
	cLog      = "log"
	cCommit   = "commit"
	cCheckout = "checkout"
)

func main() {
	var command string
	args := os.Args[1:]
	if len(args) != 1 {
		command = cHelp
	} else {
		command = args[0]
	}

	switch command {
	case cHelp:
		fmt.Println(
			"These are SVCS commands:\n" +
				"config     Get and set a username.\n" +
				"add        Add a file to the index.\n" +
				"log        Show commit logs.\n" +
				"commit     Save changes.\n" +
				"checkout   Restore a file.",
		)
	case cConfig:
		fmt.Println("Get and set a username.")
	case cAdd:
		fmt.Println("Add a file to the index.")
	case cLog:
		fmt.Println("Show commit logs.")
	case cCommit:
		fmt.Println("Save changes.")
	case cCheckout:
		fmt.Println("Restore a file.")
	default:
		fmt.Printf("'%s' is not a SVCS command.\n", command)
	}
}
