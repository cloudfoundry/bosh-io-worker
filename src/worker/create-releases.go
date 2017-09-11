package main

import (
	"fmt"
	"os"
	"path/filepath"

	semver "github.com/cppforlife/go-semi-semantic/version"

	. "worker/misc"
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

	fmt.Printf("Done\n")
}

func process(releaseDirPath string, index *ReleaseIndex, meta4 Meta4) error {
	var releases []Release

	foundReleaseMFPaths, err := filepath.Glob(filepath.Join(releaseDirPath, "releases", "*", "*.yml"))
	if err != nil {
		return fmt.Errorf("Globbing release: %s", err)
	}

	releaseFactory := ReleaseFactory{}

	for _, path := range foundReleaseMFPaths {
		if filepath.Base(path) != "index.yml" {
			releases = append(releases, releaseFactory.New(releaseDirPath, path))
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

		if !missing {
			fmt.Printf("[%s] skipping\n", release.NameVersion())
			continue
		}

		fmt.Printf("[%s] importing\n", release.NameVersion())

		relMeta, jobsMeta, file, err := release.Process()
		if err != nil {
			return fmt.Errorf("Processing release: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] processed tarball\n", release.NameVersion())

		meta4Path, err := meta4.Create(file)
		if err != nil {
			return fmt.Errorf("Creating release metalink: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] created metalink\n", release.NameVersion())

		err = index.Commit(release, relMeta, jobsMeta, meta4Path)
		if err != nil {
			return fmt.Errorf("Committing to index: release=%#v %s", release, err)
		}

		fmt.Printf("[%s] imported\n", release.NameVersion())
	}

	return nil
}
