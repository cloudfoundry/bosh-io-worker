package releases

import (
	"io/ioutil"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v1"
)

func LoadFile(path string) (Index, error) {
	indexBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading file")
	}

	var index Index

	err = yaml.Unmarshal(indexBytes, &index)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling")
	}

	return index, nil
}
