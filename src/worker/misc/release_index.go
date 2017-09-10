package misc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	semver "github.com/cppforlife/go-semi-semantic/version"
	bprel "github.com/cppforlife/bosh-provisioner/release"
	bpreljob "github.com/cppforlife/bosh-provisioner/release/job"
)

type ReleaseIndex struct {
	DirPath    string
	MinVersion semver.Version
	indexPaths []string
}

func (i *ReleaseIndex) Load() error {
	var err error

	i.indexPaths, err = filepath.Glob(filepath.Join(i.DirPath, "*"))
	if err != nil {
		return fmt.Errorf("Globbing index: %s", err)
	}

	return nil
}

func (i ReleaseIndex) Missing(release Release) (bool, error) {
	ver, err := release.Version()
	if err != nil {
		return false, err
	}

	if i.MinVersion.IsGt(ver) {
		return false, nil
	}

	for _, path := range i.indexPaths {
		if filepath.Base(path) == release.NameVersion() {
			return false, nil
		}
	}

	return true, nil
}

func (i ReleaseIndex) Commit(release Release, relMeta bprel.Release, jobsMeta []bpreljob.Job, meta4Path string) error {
	releaseDir := filepath.Join(i.DirPath, release.NameVersion())

	err := os.MkdirAll(releaseDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Mkdir index directory: %s", err)
	}

	releaseBytes, err := json.MarshalIndent(relMeta, "", "  ")
	if err != nil {
		return fmt.Errorf("Marshaling release meta: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(releaseDir, "release.v1.yml"), releaseBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Writing release meta file: %s", err)
	}

	jobsBytes, err := json.MarshalIndent(jobsMeta, "", "  ")
	if err != nil {
		return fmt.Errorf("Marshaling jobs meta: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(releaseDir, "jobs.v1.yml"), jobsBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Writing jobs meta file: %s", err)
	}

	fileBytes, err := ioutil.ReadFile(meta4Path)
	if err != nil {
		return fmt.Errorf("Reading metalink file: %s", err)
	}

	err = ioutil.WriteFile(filepath.Join(releaseDir, "source.meta4"), fileBytes, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Writing metalink file: %s", err)
	}

	cmds := [][]string{
		[]string{"config", "--local", "user.email", "bosh-io-worker"},
		[]string{"config", "--local", "user.name", "bosh-io-worker"},
		[]string{"add", "-A"},
		[]string{"commit", "-m", "add " + release.NameVersion()},
	}

	for _, cmd := range cmds {
		_, err = i.execGit(cmd)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i ReleaseIndex) execGit(args []string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = i.DirPath

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("executing git: %s (stderr: %s)", err, errBuf.String())
	}

	return outBuf.Bytes(), nil
}
