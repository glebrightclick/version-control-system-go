package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

const (
	vcsDir       = "./vcs"
	configFile   = "./vcs/config.txt"
	indexFile    = "./vcs/index.txt"
	trackedFile  = "./tracked_file.txt"
	anotherFile  = "./file.txt"
	anotherFile1 = "./new_file.txt"
)

const (
	cHelp     = "--help"
	cConfig   = "config"
	cAdd      = "add"
	cLog      = "log"
	cCommit   = "commit"
	cCheckout = "checkout"
)

func handleHelp() {
	fmt.Println(
		"These are SVCS commands:\n" +
			"config     Get and set a username.\n" +
			"add        Add a file to the index.\n" +
			"log        Show commit logs.\n" +
			"commit     Save changes.\n" +
			"checkout   Restore a file.",
	)
}

type ConfigStruct struct {
	Name string `json:"name"`
}

func handleConfig(name string) {
	var config ConfigStruct
	file, _ := os.ReadFile(configFile)
	_ = json.Unmarshal(file, &config)

	if len(name) == 0 {
		// if both saved and presented names are empty, display fallback message and exit
		if len(config.Name) == 0 {
			fmt.Println("Please, tell me who you are.")
			return
		}
	} else {
		// set new name to result
		config.Name = name

		// marshal result into new config
		configStr, _ := json.Marshal(config)
		os.WriteFile(configFile, configStr, 0644)
	}

	// display current name
	fmt.Printf("The username is %s.\n", config.Name)
}

type File struct {
	Path string `json:"path"`
}

type IndexStruct struct {
	Files []File `json:"files,omitempty"`
}

func handleAdd(name string) {
	var index IndexStruct
	file, _ := os.ReadFile(indexFile)
	_ = json.Unmarshal(file, &index)

	if len(name) == 0 {
		if len(index.Files) == 0 {
			fmt.Println("Add a file to the index.")
			return
		}

		fmt.Println("Tracked files:")
		for _, file := range index.Files {
			fmt.Println(file.Path)
		}
	} else {
		// check if file exists
		_, errorToAdd := os.ReadFile(name)
		if os.IsNotExist(errorToAdd) {
			fmt.Printf("Can't find '%s'.\n", name)
			return
		}

		index.Files = append(index.Files, File{Path: name})
		// marshal result into new config
		indexStr, _ := json.Marshal(index)
		os.WriteFile(indexFile, indexStr, 0644)
		fmt.Printf("The file '%s' is tracked.\n", name)
	}
}

func init() {
	// create vcs directory to store vcs files
	err := os.MkdirAll(vcsDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// create necessary files if it isn't presented
	filePaths := []string{configFile, indexFile, trackedFile, anotherFile, anotherFile1}
	for _, path := range filePaths {
		file, err := os.OpenFile(path, os.O_CREATE, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		file.Close()
	}
}

func main() {
	var command string
	args := os.Args[1:]
	if len(args) < 1 {
		command = cHelp
	} else {
		command = args[0]
	}

	var arg1 string
	if len(args) == 2 {
		arg1 = args[1]
	}

	switch command {
	case cHelp:
		handleHelp()
	case cConfig:
		handleConfig(arg1)
	case cAdd:
		handleAdd(arg1)
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
