package misc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	semver "github.com/cppforlife/go-semi-semantic/version"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	boshsys "github.com/cloudfoundry/bosh-utils/system"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	bpdload "github.com/cppforlife/bosh-provisioner/downloader"
	bptar "github.com/cppforlife/bosh-provisioner/tar"
	bprel "github.com/cppforlife/bosh-provisioner/release"
	bpreljob "github.com/cppforlife/bosh-provisioner/release/job"
	gouuid "github.com/nu7hatch/gouuid"
	"gopkg.in/yaml.v2"
)

type ReleaseFactory struct {}

func (f ReleaseFactory) New(releaseDirPath, mfPath string) Release {
	logger := boshlog.NewWriterLogger(boshlog.LevelDebug, os.Stderr)
	fs := boshsys.NewOsFileSystem(logger)
	runner := boshsys.NewExecCmdRunner(logger)

	extractor := bptar.NewCmdExtractor(runner, fs, logger)
	downloader := bpdload.NewDefaultMuxDownloader(fs, runner, nil, logger)
	releaseReaderFactory := bprel.NewReaderFactory(downloader, extractor, fs, logger)
	jobReaderFactory := bpreljob.NewReaderFactory(downloader, extractor, fs, logger)

	return Release{
		DirPath: releaseDirPath,
		MFPath: mfPath,

		releaseReaderFactory: releaseReaderFactory,
		jobReaderFactory: jobReaderFactory,
	}
}

type Release struct {
	DirPath string
	MFPath  string

	releaseReaderFactory bprel.ReaderFactory
	jobReaderFactory     bpreljob.ReaderFactory
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

func (r Release) Process() (bprel.Release, []bpreljob.Job, File, error) {
	var relMeta bprel.Release
	var jobsMeta []bpreljob.Job

	fileUUID, err := gouuid.NewV4()
	if err != nil {
		return relMeta, jobsMeta, File{}, fmt.Errorf("Generating tarball uuid: %s", err)
	}

	tarballPath := "/tmp/release-"+fileUUID.String()

	out, err := r.execute("bosh", []string{"create-release", r.MFPath, "--tarball", tarballPath, "--json"}, r.DirPath)
	if err != nil {
		return relMeta, jobsMeta, File{}, fmt.Errorf("Building tarball: %s", err)
	}

	var result CreateReleaseResultJSON

	err = json.Unmarshal(out, &result)
	if err != nil {
		return relMeta, jobsMeta, File{}, fmt.Errorf("Unmarshaling create release result: %s", err)
	}

	row := result.Tables[0].Rows[0]
	file := File{Path: tarballPath, Name: row.Name + "-" + row.Version + ".tgz", Version: row.Version}

	relMeta, jobsMeta, err = r.extractReleaseAndJobs(tarballPath)
	if err != nil {
		return relMeta, jobsMeta, File{}, fmt.Errorf("Extracting release meta: %s", err)
	}

	return relMeta, jobsMeta, file, nil
}

func (r Release) extractReleaseAndJobs(tgzPath string) (bprel.Release, []bpreljob.Job, error) {
	var rel bprel.Release

	relReader := r.releaseReaderFactory.NewTarReader("file://" + tgzPath)

	rel, err := relReader.Read()
	if err != nil {
		return rel, nil, bosherr.WrapError(err, "Reading release")
	}

	defer relReader.Close()

	var relJobs []bpreljob.Job

	for _, j := range rel.Jobs {
		relJobReader := r.jobReaderFactory.NewTarReader("file://" + j.TarPath)

		relJob, err := relJobReader.Read()
		if err != nil {
			return rel, nil, bosherr.WrapErrorf(err, "Reading release job '%s'", j.Name)
		}

		defer relJobReader.Close()

		relJobs = append(relJobs, relJob)
	}

	return rel, relJobs, nil
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
