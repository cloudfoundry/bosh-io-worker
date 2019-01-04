package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"worker/releases"
)

func main() {
	index, err := releases.LoadFile(os.Args[1])
	if err != nil {
		log.Panicf("loading index: %v", err)
	}

	orgs := index.GitHubOrgs()

	fmt.Println(strings.Join(orgs, "\n"))
}
