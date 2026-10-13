package releases

import (
	"fmt"
	"io/ioutil"

	yaml "gopkg.in/yaml.v1"
)

func LoadFile(path string) (Index, error) {
	indexBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file: %w", err)
	}

	var index Index

	err = yaml.Unmarshal(indexBytes, &index)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling: %w", err)
	}

	return index, nil
}
