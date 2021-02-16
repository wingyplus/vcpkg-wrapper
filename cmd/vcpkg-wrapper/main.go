package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

// Manifest is manifest file for the project.
type Manifest struct {
	// Dependencies is a list that want vcpkg to install.
	Dependencies []string
}

// TODO: install vcpkg automatically when it doesn't exists.
func main() {
	logger := log.New(os.Stderr, "[vcpkg-wrapper] ", log.LstdFlags)

	logger.Println("I! reading package.json")
	pkgf, err := os.Open("package.json")
	if err != nil {
		logger.Printf("E! cannot open package.json: %v", err)
		os.Exit(1)
	}

	var manifest Manifest
	if err := json.NewDecoder(pkgf).Decode(&manifest); err != nil {
		logger.Printf("E! cannot decoding json from package.json file: %v", err)
		os.Exit(1)
	}

	logger.Println("I! installing packages...")

	vcpkgBin := filepath.Join(".", ".vcpkg", "vcpkg", "vcpkg")
	// vcpkg binary
	_, err = os.Stat(vcpkgBin)
	if os.IsNotExist(err) {
		logger.Println("E! vcpkg binary not found, please install it before using this command")
		os.Exit(1)
	}

	for _, pkg := range manifest.Dependencies {
		cmd := exec.Command(vcpkgBin, "install", pkg)
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		if err := cmd.Start(); err != nil {
			logger.Printf("E! cannot starting execution: %v", err)
		}

		go tail(logger, stdout)
		go tail(logger, stderr)

		if err := cmd.Wait(); err != nil {
			logger.Printf("E! vcpkg install %s failed. cause: %v", pkg, err)
		}

		stdout.Close()
		stderr.Close()
	}

	logger.Println("I! DONE")
}

func tail(logger *log.Logger, rc io.ReadCloser) {
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		logger.Printf("I! %s", string(scanner.Bytes()))
	}
	if err := scanner.Err(); err != nil {
		logger.Printf("E! found an error during tailing an output: %v", err)
	}
}
