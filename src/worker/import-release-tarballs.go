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

	err := ReleaseTarballs{ReleasesIndexDir: os.Args[1]}.Import(os.Stdin)
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

type ReleaseTarballs struct {
	ReleasesIndexDir string
}

type key struct {
	Source     string
	VersionRaw string
}

type tarballVal struct {
	SHA1   string
	BlobID string
}

// eg   {"Source":"github.com/cloudfoundry-incubator/diego-release","VersionRaw":"0.549"} |
// {"BlobID":"4cbebec9-36e1-4214-6cd5-d8118a85826b","SHA1":"18ef9e77924728f752bce6e4adeb9754c165291d","DownloadURL":"https://s3.amazonaws.com/bosh-hub-releases/4cbebec9-36e1-4214-6cd5-d8118a85826b"}
func (r ReleaseTarballs) Import(data io.Reader) error {
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

		var val tarballVal

		err = json.Unmarshal([]byte(pieces[1]), &val)
		if err != nil {
			return fmt.Errorf("Unmarshaling val: %s: %s", pieces[1], err)
		}

		if len(val.SHA1) != 40 {
			return fmt.Errorf("Invalid SHA1 '%s'", val.SHA1)
		}

		if len(val.BlobID) == 0 {
			return fmt.Errorf("Invalid Blob ID '%#v'", val)
		}

		paths, err := filepath.Glob(filepath.Join(r.ReleasesIndexDir, k.Source, "*-"+k.VersionRaw))
		if err != nil {
			return fmt.Errorf("Globbing: %s", err)
		}

		if len(paths) != 1 {
			return fmt.Errorf("Expected paths != 1")
		}

		sourcePath := filepath.Join(paths[0], "source.meta4")

		contents := fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<metalink xmlns="urn:ietf:params:xml:ns:metalink">
  <file name="%s.tgz">
    <hash type="sha-1">%s</hash>
    <url>https://s3.amazonaws.com/bosh-hub-release-tarballs/%s</url>
    <version>%s</version>
  </file>
</metalink>
`, filepath.Base(paths[0]), val.SHA1, val.BlobID, k.VersionRaw)

		err = ioutil.WriteFile(sourcePath, []byte(contents), 0644)
		if err != nil {
			return fmt.Errorf("Writing file: path=%s %s", sourcePath, err)
		}
	}

	return nil
}
