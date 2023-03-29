package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var Stdout = make(chan string)
var StdoutJson = make(chan any)
var Stderr = make(chan string)
var StderrJson = make(chan any)

func writePipe(w *bufio.Writer, pipe chan string) {
	defer w.Flush()
	for res := range pipe {
		fmt.Fprintln(w, res)
	}
}

func writePipeJson(w *bufio.Writer, pipe chan any) {
	defer w.Flush()
	enc := json.NewEncoder(w)
	for res := range pipe {
		err := enc.Encode(res)
		if err != nil {
			log.Println(err)
		}
	}
}

func CloseWriter() {
	close(Stderr)
	close(StderrJson)
	close(Stdout)
	close(StdoutJson)
}

func Writer() {
	stdout := bufio.NewWriter(os.Stdout)
	stderr := bufio.NewWriter(os.Stderr)
	var wg sync.WaitGroup
	wg.Add(4)
	go func() {
		defer wg.Done()
		writePipe(stdout, Stdout)
	}()
	go func() {
		defer wg.Done()
		writePipe(stderr, Stderr)
	}()
	go func() {
		defer wg.Done()
		writePipeJson(stdout, StdoutJson)
	}()
	go func() {
		defer wg.Done()
		writePipeJson(stderr, StderrJson)
	}()
	wg.Wait()
}
