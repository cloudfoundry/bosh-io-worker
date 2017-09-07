package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	semver "github.com/cppforlife/go-semi-semantic/version"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) != 5 {
		panic(fmt.Sprintf("Wrong args: release-dir(eg bosh/) release-index-repo-dir(eg blah/github.com/cloudfoundry/bosh/) min-version(eg 0.0) tarball-dst(eg s3://...)"))
	}

	index := &ReleaseIndex{
		DirPath:    os.Args[2],
		MinVersion: semver.MustNewVersionFromString(os.Args[3]),
	}

	err := process(os.Args[1], index, Meta4{Dst: os.Args[4]})
	if err != nil {
		panic(fmt.Sprintf("Failed: %s", err))
	}
}

func process(releaseDirPath string, index *ReleaseIndex, meta4 Meta4) error {
	var releases []Release

	foundReleaseMFPaths, err := filepath.Glob(filepath.Join(releaseDirPath, "releases", "*", "*.yml"))
	if err != nil {
		return fmt.Errorf("Globbing release: %s", err)
	}

	for _, path := range foundReleaseMFPaths {
		if filepath.Base(path) != "index.yml" {
			releases = append(releases, Release{DirPath: releaseDirPath, MFPath: path})
		}
	}

	err = index.Load()
	if err != nil {
		return fmt.Errorf("Loading release index: %s", err)
	}

	for _, release := range releases {
		missing, err := index.Missing(release)
		if err != nil {
			return fmt.Errorf("Checking if release is missing: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] skipping\n", release.NameVersion())

		if !missing {
			continue
		}

		fmt.Printf("[%s] importing\n", release.NameVersion())

		file, err := release.Process()
		if err != nil {
			return fmt.Errorf("Processing release: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] processing tarball\n", release.NameVersion())

		meta4Path, err := meta4.Create(file)
		if err != nil {
			return fmt.Errorf("Creating release metalink: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] created metalink\n", release.NameVersion())

		err = index.Commit(release, meta4Path)
		if err != nil {
			return fmt.Errorf("Committing to index: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] imported\n", release.NameVersion())
	}

	return nil
}

type Release struct {
	DirPath string
	MFPath  string
}

type CreateReleaseResultJSON struct {
	Tables []CLITableJSON
}

type CLITableJSON struct {
	Rows []CreateReleaseResultRowJSON
}

type CreateReleaseResultRowJSON struct {
	Name    string
	Version string
}

type ReleaseManifest struct {
	Version string
}

func (r Release) Version() (semver.Version, error) {
	manifestBytes, err := ioutil.ReadFile(r.MFPath)
	if err != nil {
		return semver.Version{}, fmt.Errorf("Reading release manifest: %s", err)
	}

	var manifest ReleaseManifest

	err = yaml.Unmarshal(manifestBytes, &manifest)
	if err != nil {
		return semver.Version{}, fmt.Errorf("Deserializing manifest: %s", err)
	}

	ver, err := semver.NewVersionFromString(manifest.Version)
	if err != nil {
		return semver.Version{}, fmt.Errorf("Parsing release version: %s", err)
	}

	return ver, nil
}

func (r Release) NameVersion() string {
	return strings.TrimSuffix(filepath.Base(r.MFPath), ".yml")
}

func (r Release) Process() (File, error) {
	tarballPath := "/tmp/release"

	out, err := r.execute("bosh", []string{"create-release", r.MFPath, "--tarball", tarballPath, "--json"}, r.DirPath)
	if err != nil {
		return File{}, fmt.Errorf("Building tarball: %s", err)
	}

	var result CreateReleaseResultJSON

	err = json.Unmarshal(out, &result)
	if err != nil {
		return File{}, fmt.Errorf("Unmarshaling create release result: %s", err)
	}

	row := result.Tables[0].Rows[0]

	return File{Path: tarballPath, Name: row.Name + "-" + row.Version + ".tgz", Version: row.Version}, nil
}

func (Release) execute(path string, args []string, dir string) ([]byte, error) {
	cmd := exec.Command(path, args...)
	cmd.Dir = dir

	cmd.Env = append(
		os.Environ(),
		"BOSH_NON_INTERACTIVE=true",
		"HOME=/tmp",
	)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("executing %s: %s (stdout: %s stderr: %s)", path, err, outBuf.Bytes(), errBuf.String())
	}

	return outBuf.Bytes(), nil
}

type File struct {
	Path    string
	Name    string
	Version string
}

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

func (i ReleaseIndex) Commit(release Release, meta4Path string) error {
	releaseDir := filepath.Join(i.DirPath, release.NameVersion())

	err := os.Mkdir(releaseDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("Mkdir index directory: %s", err)
	}

	err = os.Rename(meta4Path, filepath.Join(releaseDir, "source.meta4"))
	if err != nil {
		return fmt.Errorf("Mkdir index directory: %s", err)
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

type Meta4 struct {
	Dst string
}

func (m Meta4) Create(file File) (string, error) {
	meta4Path := "/tmp/metalink"

	_, err := m.execute([]string{"create", "--metalink", meta4Path})
	if err != nil {
		return "", err
	}

	_, err = m.execute([]string{
		"import-file",
		fmt.Sprintf("file://%s", file.Path),
		"--version", file.Version,
		"--file", file.Name,
		"--metalink", meta4Path,
	})
	if err != nil {
		return "", err
	}

	uuid, err := exec.Command("uuidgen").Output()
	if err != nil {
		return "", fmt.Errorf("Generating file upload guid: %s", err)
	}

	uuidStr := strings.TrimSpace(strings.ToLower(string(uuid)))

	_, err = m.execute([]string{
		"file-upload",
		fmt.Sprintf("file://%s", file.Path),
		fmt.Sprintf("%s/%s", m.Dst, uuidStr),
		"--file", file.Name,
		"--metalink", meta4Path,
	})
	if err != nil {
		return "", err
	}

	return meta4Path, nil
}

func (Meta4) execute(args []string) ([]byte, error) {
	cmd := exec.Command("meta4", args...)

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("executing meta4: %s %v (stderr: %s)", err, args, errBuf.String())
	}

	return outBuf.Bytes(), nil
}
