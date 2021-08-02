package main

import (
	"fmt"
	"github.com/bosh-io/worker/src/worker/releases"
	"log"
	"os"
	"strings"
)

func main() {
	index, err := releases.LoadFile(os.Args[1])
	if err != nil {
		log.Panicf("loading index: %v", err)
	}

	orgs := index.GitHubOrgs()

	fmt.Println(strings.Join(orgs, "\n"))
}
