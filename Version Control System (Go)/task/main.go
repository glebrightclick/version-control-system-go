package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
)

const (
	vcsDir     = "./vcs"
	commitsDir = "./vcs/commits"

	configFile = "./vcs/config.txt"
	indexFile  = "./vcs/index.txt"
	logFile    = "./vcs/log.txt"
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

func getConfig() ConfigStruct {
	var config ConfigStruct
	file, _ := os.ReadFile(configFile)
	_ = json.Unmarshal(file, &config)
	return config
}

func handleConfig(name string) {
	config := getConfig()

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
		save(configFile, config)
	}

	// display current name
	fmt.Printf("The username is %s.\n", config.Name)
}

type File struct {
	Path  string `json:"path"`
	IsDir bool   `json:"-"`
}

func (f *File) prettyPath() string {
	return regexp.MustCompile("^./(.*)$").ReplaceAllString(f.Path, `$1`)
}

type IndexStruct struct {
	Files []File `json:"files,omitempty"`
}

func getIndex() IndexStruct {
	var index IndexStruct
	file, _ := os.ReadFile(indexFile)
	_ = json.Unmarshal(file, &index)
	return index
}

func getIndexWithDirectories() []File {
	index := getIndex()
	var result []File

	for _, file := range index.Files {
		path := file.prettyPath()
		directories := strings.Split(path, "/")
		parentPath := "./"
		for i := 0; i < len(directories)-1; i++ {
			result = append(result, File{Path: parentPath + directories[i], IsDir: true})
			parentPath += directories[i] + `/`
		}

		result = append(result, file)
	}

	return result
}

func handleAdd(name string) {
	index := getIndex()

	if len(name) == 0 {
		if len(index.Files) == 0 {
			fmt.Println("Add a file to the index.")
			return
		}

		fmt.Println("Tracked files:")
		for _, file := range index.Files {
			fmt.Println(file.prettyPath())
		}
	} else {
		// check if file exists
		_, errorToAdd := os.ReadFile(name)
		if os.IsNotExist(errorToAdd) {
			fmt.Printf("Can't find '%s'.\n", name)
			return
		}

		pathName := fmt.Sprintf("%s/%s", ".", name)
		index.Files = append(index.Files, File{Path: pathName, IsDir: false})
		// marshal result into new config
		indexStr, _ := json.Marshal(index)
		os.WriteFile(indexFile, indexStr, 0644)
		fmt.Printf("The file '%s' is tracked.\n", name)
	}
}

type Commit struct {
	Hash    string `json:"hash"`
	Author  string `json:"author"`
	Message string `json:"message"`
}

type LogStruct struct {
	Commits []Commit `json:"commits"`
}

func getLog() LogStruct {
	var log LogStruct
	file, _ := os.ReadFile(logFile)
	_ = json.Unmarshal(file, &log)

	return log
}

func save(fileName string, content interface{}) {
	str, _ := json.Marshal(content)
	os.WriteFile(fileName, str, 0644)
}

func handleLog() {
	logData := getLog()

	if len(logData.Commits) == 0 {
		fmt.Println("No commits yet.")
	} else {
		// display all commits one after another
		for j := len(logData.Commits) - 1; j >= 0; j-- {
			commit := logData.Commits[j]
			fmt.Printf(
				"commit %s\nAuthor: %s\n%s\n\n",
				commit.Hash,
				commit.Author,
				commit.Message,
			)
		}
	}
}

func calculateHash() string {
	hash := md5.New()
	for _, file := range getIndex().Files {
		file, _ := os.ReadFile(file.Path)
		hash.Write(file)
	}

	// go through current dir - if file is a dir - get full name
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func handleCommit(message string) {
	if len(message) == 0 {
		fmt.Println("Message was not passed.")
		return
	}

	// 1. take last commit hash from log
	lastCommitHash, logData := "", getLog()
	if len(logData.Commits) > 0 {
		lastCommit := logData.Commits[len(logData.Commits)-1]
		lastCommitHash = lastCommit.Hash
	}

	// 2. calculate current state of project hash
	currentHash := calculateHash()

	// 3. compare these two, if equals - display "Nothing to commit."
	if lastCommitHash == currentHash {
		fmt.Println("Nothing to commit.")
	} else {
		// perform commit:
		// 1. create directory with hash name in ./vcs/commits directory
		currentDir := fmt.Sprintf("%s/%s", commitsDir, currentHash)
		err := os.MkdirAll(currentDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		// go through the entire project and copy
		trackedFiles := getIndexWithDirectories()
		for _, file := range trackedFiles {
			path := file.Path
			fullPath := fmt.Sprintf(
				"%s/%s",
				currentDir,
				file.prettyPath(),
			)

			_, err := os.ReadDir(path)
			if err == nil {
				os.MkdirAll(fullPath, os.ModePerm)
				continue
			}

			_, err = copy(path, fullPath)
			if err != nil {
				log.Fatal(err)
			}
		}

		// add to log
		config := getConfig()
		logData.Commits = append(
			logData.Commits,
			Commit{Hash: currentHash, Author: config.Name, Message: message},
		)
		save(logFile, logData)
		fmt.Println("Changes are committed.")
	}
}

func handleCheckout(commit string) {
	if len(commit) == 0 {
		fmt.Println("Commit id was not passed.")
		return
	}

	dir := fmt.Sprintf("%s/%s", commitsDir, commit)
	_, errorToAdd := os.ReadDir(dir)
	if os.IsNotExist(errorToAdd) {
		fmt.Println("Commit does not exist.")
		return
	}

	// delete all files from index.txt
	for _, file := range getIndex().Files {
		path := file.Path
		errorToRemove := os.Remove(path)
		if errorToRemove != nil {
			log.Fatal(errorToRemove)
		}

		// copy from commit path
		_, errorToCopy := copy(dir+`/`+file.prettyPath(), file.Path)
		if errorToCopy != nil {
			log.Fatal(errorToCopy)
		}
	}

	fmt.Printf("Switched to commit %s.\n", commit)
}

func init() {
	// create necessary directories
	dirPaths := []string{vcsDir, commitsDir}
	for _, dirPath := range dirPaths {
		err := os.MkdirAll(dirPath, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
	}

	// create necessary files
	filePaths := []string{configFile, indexFile, logFile}
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
		handleLog()
	case cCommit:
		handleCommit(arg1)
	case cCheckout:
		handleCheckout(arg1)
	default:
		fmt.Printf("'%s' is not a SVCS command.\n", command)
	}
}
