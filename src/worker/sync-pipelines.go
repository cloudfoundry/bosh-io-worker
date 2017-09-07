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
		panic(fmt.Sprintf("Wrong args: bosh-io-releases-index-path(eg releases/index.yml) fly-release-tpl-path(eg pipelines/release-tpl.yml) fly-vars"))
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

type Concourse struct {
	ReleaseTplPath string
	VarsPath       string
}

func (c Concourse) SyncPipelines(defs []ReleaseDef) error {
	currPipelines, err := c.pipelines()
	if err != nil {
		return err
	}

	err = c.checkBoshIOTeam(currPipelines)
	if err != nil {
		return err
	}

	releaseTplPipelines := map[string]struct{}{}

	for _, name := range currPipelines {
		if strings.HasPrefix(name, "release:") {
			releaseTplPipelines[name] = struct{}{}
		}
	}

	for _, def := range defs {
		name := "release:" + def.PipelineSlug()
		fmt.Printf("[%s] updating\n", name)

		err = c.updatePipeline(name, def)
		if err != nil {
			return err
		}

		delete(releaseTplPipelines, name)
	}

	for name, _ := range releaseTplPipelines {
		fmt.Printf("[%s] deleting\n", name)

		err = c.deletePipeline(name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Concourse) checkBoshIOTeam(pipelines []string) error {
	expectedName := "bosh-io-team-check"

	for _, name := range pipelines {
		if name == expectedName {
			return nil
		}
	}

	return fmt.Errorf("Expected to be in bosh-io team (missing '%s' pipeline)", expectedName)
}

func (c Concourse) updatePipeline(name string, def ReleaseDef) error {
	_, err := c.execute([]string{
		"set-pipeline",
		"-p", name,
		"-c", c.ReleaseTplPath,
		"-l", c.VarsPath,
		"-v", "release_git_url=" + def.URL,
		"-v", "release_min_version=" + def.MinVersion,
		"-v", "release_repo=" + def.IndexDirectory(),
	})
	if err != nil {
		return fmt.Errorf("Updating pipeline: %s", err)
	}

	_, err = c.execute([]string{"unpause-pipeline", "-p", name})
	if err != nil {
		return fmt.Errorf("Unpausing pipeline: %s", err)
	}

	return nil
}

func (c Concourse) deletePipeline(name string) error {
	_, err := c.execute([]string{"destroy-pipeline", "-p", name})
	if err != nil {
		return fmt.Errorf("Deleting pipeline: %s", err)
	}

	return nil
}

func (c Concourse) pipelines() ([]string, error) {
	out, err := c.execute([]string{"pipelines"})
	if err != nil {
		return nil, fmt.Errorf("Fetching pipelines: %s", err)
	}

	var pipelineNames []string

	for _, line := range strings.Split(string(out), "\n") {
		if len(line) > 0 {
			pieces := strings.SplitN(line, "  ", 2)
			if len(pieces) == 2 && len(pieces[0]) > 0 {
				pipelineNames = append(pipelineNames, pieces[0])
			}
		}
	}

	return pipelineNames, nil
}

func (c Concourse) execute(args []string) ([]byte, error) {
	cmd := exec.Command("fly", append([]string{"-t", "production"}, args...)...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	cmd.Stdin = bytes.NewBufferString("y\n")

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("executing meta4: %s %v (stderr: %s)", err, args, errBuf.String())
	}

	return outBuf.Bytes(), nil
}
