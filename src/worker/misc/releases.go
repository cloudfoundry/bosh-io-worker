package misc

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

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
