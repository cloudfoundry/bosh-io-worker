package main

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) != 4 {
		panic(fmt.Sprintf("Wrong args: bosh-io-releases-index-path(eg releases/index.yml) "))
	}

	err := process(Releases{os.Args[1]}, Concourse{ReleaseTplPath: os.Args[2], VarsPath: os.Args[3]})
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

func process(releases Releases, concourse Concourse) error {
	rels, err := releases.Releases()
	if err != nil {
		return err
	}

	periodicGithubNoteImporter := NewPeriodicGithubNoteImporter(
		options.Period,
		options.GithubPersonalAccessToken,
		make(chan struct{}),
		repos.ReleasesRepo(),
		logger,
	)

	return concourse.SyncPipelines(rels)
}

type Releases struct {
	IndexPath string
}

type ReleaseDef struct {
	URL        string `yaml:"url"`
	MinVersion string `yaml:"min_version"`
}

func (d ReleaseDef) PipelineSlug() string {
	return strings.Replace(strings.TrimPrefix(d.URL, "https://"), "/", ":", -1)
}

func (d ReleaseDef) IndexDirectory() string {
	return strings.TrimPrefix(d.URL, "https://")
}

func (r Releases) Releases() ([]ReleaseDef, error) {
	indexBytes, err := ioutil.ReadFile(r.IndexPath)
	if err != nil {
		return nil, fmt.Errorf("Reading releases index: %s", err)
	}

	var defs []ReleaseDef

	err = yaml.Unmarshal(indexBytes, &defs)
	if err != nil {
		return nil, fmt.Errorf("Unmarshaling releases index: %s", err)
	}

	return defs, nil
}
