package releases

import (
	"log"
	"net/url"
	"sort"
	"strings"
)

type Index []Release

func (i Index) GitHubOrgs() []string {
	orgsMap := map[string]struct{}{}

	for _, release := range i {
		orgsMap[release.GitHubOrg()] = struct{}{}
	}

	var orgs []string

	for org := range orgsMap {
		orgs = append(orgs, org)
	}

	sort.Strings(orgs)

	return orgs
}

func (i Index) FilterByOrg(org string) Index {
	var filtered Index

	for _, release := range i {
		if org != release.GitHubOrg() {
			continue
		}

		filtered = append(filtered, release)
	}

	return filtered
}

type Release struct {
	URL        ReleaseURL `yaml:"url"`
	Categories []string   `yaml:"categories"`
	MinVersion string     `yaml:"min_version"`
	Homepage   bool       `yaml:"homepage"`
}

type ReleaseURL string

func (r Release) GitHubOrg() string {
	r.URL.requireGitHubURL()

	return strings.SplitN(r.URL.Parse().Path, "/", 4)[1]
}

func (r Release) GitHubRepo() string {
	r.URL.requireGitHubURL()

	return strings.SplitN(r.URL.Parse().Path, "/", 4)[2]
}

func (r ReleaseURL) Parse() *url.URL {
	url, err := url.Parse(string(r))
	if err != nil {
		log.Panicf("failed to parse url %s: %v", r, err)
	}

	return url
}

func (r ReleaseURL) requireGitHubURL() {
	// in theory we could support other ones; being lazy now for easier downstream assumptions
	if r.Parse().Hostname() != "github.com" {
		log.Panicf("not a github.com url (and we have been making assumptions on this): %s", r)
	}
}
