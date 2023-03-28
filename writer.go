package main

import (
	"fmt"
	"os"
	"log"
	"bufio"
	"encoding/json"
)

var Output = make(chan string)
var JsonOutput = make(chan any)

func writer() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for res := range Output {
		fmt.Fprintln(w, res)
	}
}

func jsonWriter() {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	for res := range JsonOutput {
		err := json.NewEncoder(w).Encode(res)
		if err != nil {
			log.Println(err)
		}
	}
}
