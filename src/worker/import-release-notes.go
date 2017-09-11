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
		panic(fmt.Sprintf("Wrong args: bosh-io-releases-index-dir(eg releases-index)"))
	}

	err := ReleaseNotes{ReleasesIndexDir: os.Args[1]}.Import(os.Stdin)
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

type ReleaseNotes struct {
	ReleasesIndexDir string
}

type key struct {
	Source     string
	VersionRaw string
}

type noteVal struct {
	Content string
}

// eg   {"Source":"github.com/cloudfoundry-incubator/diego-release","VersionRaw":"0.549"} | {"Content":"..."}
func (r ReleaseNotes) Import(data io.Reader) error {
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

		if len(k.Source) == 0 || len(k.VersionRaw) == 0 {
			return fmt.Errorf("Invalid key from '%s'", pieces[0])
		}

		// deleted release -> skip
		switch {
		case k.Source == "github.com/cloudfoundry/uaa-release" && k.VersionRaw == "v15":
			continue
		case k.Source == "github.com/cloudfoundry/dotnet-core-buildpack-release" && k.VersionRaw == "v1.0.3":
			continue
		}

		fmt.Printf("[%#v] processing\n", k)

		var val noteVal

		err = json.Unmarshal([]byte(pieces[1]), &val)
		if err != nil {
			return fmt.Errorf("Unmarshaling val: %s: %s", pieces[1], err)
		}

		paths, err := filepath.Glob(filepath.Join(r.ReleasesIndexDir, k.Source, "*-"+k.VersionRaw))
		if err != nil {
			return fmt.Errorf("Globbing: %s", err)
		}

		if len(paths) != 1 {
			return fmt.Errorf("Expected paths != 1")
		}

		sourcePath := filepath.Join(paths[0], "notes.v1.yml")

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
