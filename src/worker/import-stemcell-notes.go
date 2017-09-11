package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 2 {
		panic(fmt.Sprintf("Wrong args: bosh-io-stemcells-index-dir(eg stemcells-index)"))
	}

	err := StemcellNotes{StemcellsIndexDir: os.Args[1]}.Import(os.Stdin)
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

type StemcellNotes struct {
	StemcellsIndexDir string
}

type key struct {
	VersionRaw string
}

type noteVal struct {
	Content string
}

// eg   {"VersionRaw":"3363.24"} | {"Content":"..."}
func (r StemcellNotes) Import(data io.Reader) error {
	rd := bufio.NewReader(data)
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("Reading data: %s", err)
		}

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		pieces := strings.SplitN(line, "|", 2)
		if len(pieces) != 2 {
			return fmt.Errorf("Parsing line: '%s'", line)
		}

		pieces[0] = strings.TrimSpace(pieces[0])
		pieces[1] = strings.TrimSpace(pieces[1])

		var k key

		err = json.Unmarshal([]byte(pieces[0]), &k)
		if err != nil {
			return fmt.Errorf("Unmarshaling key: %s", pieces[0])
		}

		if len(k.VersionRaw) == 0 {
			return fmt.Errorf("Invalid key from '%s'", pieces[0])
		}

		fmt.Printf("[%#v] processing\n", k)

		var val noteVal

		err = json.Unmarshal([]byte(pieces[1]), &val)
		if err != nil {
			return fmt.Errorf("Unmarshaling val: %s: %s", pieces[1], err)
		}

		path := filepath.Join(r.StemcellsIndexDir, k.VersionRaw)

		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("Creating dir: %s", err)
		}

		sourcePath := filepath.Join(path, "notes.v1.yml")

		content, err := json.MarshalIndent(val, "", "  ")
		if err != nil {
			return fmt.Errorf("Marshaling val: %s", err)
		}

		err = ioutil.WriteFile(sourcePath, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("Writing file: path=%s %s", sourcePath, err)
		}
	}

	return nil
}
