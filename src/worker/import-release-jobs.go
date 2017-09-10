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

	err := ReleaseJobs{ReleasesIndexDir: os.Args[1]}.Import(os.Stdin)
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

type ReleaseJobs struct {
	ReleasesIndexDir string
}

type key struct {
	Source     string
	VersionRaw string
}

type jobs []interface{}

// eg  {"Source":"github.com/cppforlife/turbulence-release","VersionRaw":"0.10.0"} | [{"Name":"turbulence_agent","Description":"","MonitTemplate":{"SrcPathEnd":"monit","DstPathEnd":"monit","Path":"/mnt/tmp/tar-CmdExtractor833769197/monit"},"Templates":[{"SrcPathEnd":"ctl.erb","DstPathEnd":"bin/ctl","Path":"/mnt/tmp/tar-CmdExtractor833769197/templates/ctl.erb"},{"SrcPathEnd":"config.json.erb","DstPathEnd":"config/config.json","Path":"/mnt/tmp/tar-CmdExtractor833769197/templates/config.json.erb"}],"Packages":[{"Name":"turbulence"},{"Name":"stress"}],"Properties":[{"Name":"debug","Description":"Show debug logs","Default":true,"Example":null,"Examples":null}]},{"Name":"turbulence_api","Description":"","MonitTemplate":{"SrcPathEnd":"monit","DstPathEnd":"monit","Path":"/mnt/tmp/tar-CmdExtractor107680167/monit"},"Templates":[{"SrcPathEnd":"ctl.erb","DstPathEnd":"bin/ctl","Path":"/mnt/tmp/tar-CmdExtractor107680167/templates/ctl.erb"},{"SrcPathEnd":"cert","DstPathEnd":"config/cert","Path":"/mnt/tmp/tar-CmdExtractor107680167/templates/cert"},{"SrcPathEnd":"private_key","DstPathEnd":"config/private_key","Path":"/mnt/tmp/tar-CmdExtractor107680167/templates/private_key"},{"SrcPathEnd":"config.json.erb","DstPathEnd":"config/config.json","Path":"/mnt/tmp/tar-CmdExtractor107680167/templates/config.json.erb"}],"Packages":[{"Name":"turbulence"}],"Properties":[{"Name":"listen_address","Description":"Address to listen on","Default":"0.0.0.0","Example":null,"Examples":null},{"Name":"listen_port","Description":"Operator API listen port","Default":8080,"Example":null,"Examples":null},{"Name":"director.host","Description":"Director host","Default":null,"Example":"192.168.50.4","Examples":null},{"Name":"username","Description":"Username for the API authentication","Default":"turbulence","Example":null,"Examples":null},{"Name":"password","Description":"Password for the API authentication","Default":null,"Example":null,"Examples":null},{"Name":"director.client","Description":"Director client (username in case of basic auth)","Default":null,"Example":null,"Examples":null},{"Name":"datadog.app_key","Description":"Datadog application key used for incident reporting","Default":"","Example":null,"Examples":null},{"Name":"advertised_host","Description":"Advertised hostname of the API server","Default":"","Example":null,"Examples":null},{"Name":"cert","Description":"API server certificate","Default":null,"Example":null,"Examples":null},{"Name":"director.cert.ca","Description":"CA certificate to verify Director certificate (uses system CA certificates by default)","Default":"","Example":null,"Examples":null},{"Name":"datadog.api_key","Description":"Datadog API key","Default":"","Example":null,"Examples":null},{"Name":"debug","Description":"Show debug logs","Default":true,"Example":null,"Examples":null},{"Name":"agent_listen_port","Description":"Agent API listen port","Default":8081,"Example":null,"Examples":null},{"Name":"director.port","Description":"Director port","Default":25555,"Example":null,"Examples":null},{"Name":"director.client_secret","Description":"Director client secret (password in case of basic auth)","Default":null,"Example":null,"Examples":null}]}]
func (r ReleaseJobs) Import(data io.Reader) error {
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

		var js jobs

		err = json.Unmarshal([]byte(pieces[1]), &js)
		if err != nil {
			return fmt.Errorf("Unmarshaling value: %s", pieces[1])
		}

		jsBytes, err := json.MarshalIndent(js, "", "  ")
		if err != nil {
			return fmt.Errorf("Marshaling value: %s", err)
		}

		paths, err := filepath.Glob(filepath.Join(r.ReleasesIndexDir, k.Source, "*-"+k.VersionRaw))
		if err != nil {
			return fmt.Errorf("Globbing: %s", err)
		}

		if len(paths) != 1 {
			return fmt.Errorf("Expected paths != 1")
		}

		jobsPath := filepath.Join(paths[0], "jobs.v1.yml")

		err = ioutil.WriteFile(jobsPath, jsBytes, 0644)
		if err != nil {
			return fmt.Errorf("Writing file: path=%s %s", jobsPath, err)
		}
	}

	return nil
}
