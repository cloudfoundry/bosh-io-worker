package main

import (
	"fmt"
	"log"
	"os"

	"worker/pipelines"
	"worker/releases"
)

func main() {
	index, err := releases.LoadFile(os.Args[1])
	if err != nil {
		log.Panicf("loading index: %v", err)
	}

	org := pipelines.NewOrgPipeline(os.Args[2])

	for _, release := range index.FilterByOrg(org.Name()) {
		org.AddRelease(release)
	}

	fmt.Println(string(org.PipelineBytes()))
}
