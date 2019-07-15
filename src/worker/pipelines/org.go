package pipelines

import (
	"fmt"
	"log"
	"strings"

	"worker/releases"

	"github.com/concourse/concourse/atc"
	yaml "gopkg.in/yaml.v1"
)

type OrgPipeline struct {
	name     string
	pipeline *atc.Config
}

func NewOrgPipeline(name string) *OrgPipeline {
	return &OrgPipeline{
		name: name,
		pipeline: &atc.Config{
			Groups: atc.GroupConfigs{
				atc.GroupConfig{
					Name: "all",
				},
			},
			Resources: atc.ResourceConfigs{
				atc.ResourceConfig{
					Name: "releases-index",
					Type: "git",
					Source: atc.Source{
						"uri":         "((releases_index_git_url))",
						"branch":      "master",
						"private_key": "((releases_index_private_key))",
					},
				},
				atc.ResourceConfig{
					Name: "worker",
					Type: "git",
					Source: atc.Source{
						"uri": "https://github.com/bosh-io/worker.git",
					},
				},
			},
		},
	}
}

func (o *OrgPipeline) Name() string {
	return o.name
}

func (o *OrgPipeline) PipelineBytes() []byte {
	pipelineBytes, err := yaml.Marshal(o.pipeline)
	if err != nil {
		log.Panicf("marshaling pipeline: %v", err)
	}

	return pipelineBytes
}

func (op *OrgPipeline) AddGroupJob(groupName, jobName string) {
	for groupIdx, group := range op.pipeline.Groups {
		if group.Name != groupName {
			continue
		}

		op.pipeline.Groups[groupIdx].Jobs = append(op.pipeline.Groups[groupIdx].Jobs, jobName)

		return
	}

	op.pipeline.Groups = append(
		op.pipeline.Groups,
		atc.GroupConfig{
			Name: groupName,
			Jobs: []string{jobName},
		},
	)
}

func (op *OrgPipeline) AddRelease(r releases.Release) {
	name := r.GitHubRepo()
	repoResourceName := fmt.Sprintf("%s-repo", name)
	minVersion := "0"

	if r.MinVersion != "" {
		minVersion = r.MinVersion
	}

	op.pipeline.Resources = append(
		op.pipeline.Resources,
		atc.ResourceConfig{
			Name:       repoResourceName,
			Type:       "git",
			CheckEvery: "5m",
			Source: atc.Source{
				"uri":             string(r.URL),
				"disable_ci_skip": true,
			},
		},
	)

	op.pipeline.Jobs = append(
		op.pipeline.Jobs,
		atc.JobConfig{
			Name:   name,
			Serial: true,
			Plan: atc.PlanSequence{
				atc.PlanConfig{
					InParallel: &atc.InParallelConfig{
						Steps: atc.PlanSequence{
							atc.PlanConfig{
								Get: "worker",
							},
							atc.PlanConfig{
								Get:      "release",
								Resource: repoResourceName,
								Trigger:  true,
								Params: atc.Params{
									"submodules": "none",
								},
							},
							atc.PlanConfig{
								Get: "releases-index",
							},
						},
					},
				},
				atc.PlanConfig{
					Task: "sync",
					Params: atc.Params{
						"AWS_ACCESS_KEY_ID":     "((s3_access_key_id))",
						"AWS_SECRET_ACCESS_KEY": "((s3_secret_access_key))",
					},
					TaskConfig: &atc.TaskConfig{
						Platform: "linux",
						ImageResource: &atc.ImageResource{
							Type: "docker-image",
							Source: atc.Source{
								"repository": "golang",
								"tag":        "1.11",
							},
						},
						Inputs: []atc.TaskInputConfig{
							atc.TaskInputConfig{
								Name: "release",
							},
							atc.TaskInputConfig{
								Name: "releases-index",
								Path: "releases-index-input",
							},
							atc.TaskInputConfig{
								Name: "worker",
							},
						},
						Outputs: []atc.TaskOutputConfig{
							atc.TaskOutputConfig{
								Name: "releases-index",
							},
						},
						Run: atc.TaskRunConfig{
							Path: "bash",
							Args: []string{
								"-c",
								fmt.Sprintf(
									`
set -eu
wget -O /usr/bin/bosh https://s3.amazonaws.com/bosh-cli-artifacts/bosh-cli-5.4.0-linux-amd64
echo "ecc1b6464adf9a0ede464b8699525a473e05e7205357e4eb198599edf1064f57  /usr/bin/bosh" | sha256sum -c -
chmod +x /usr/bin/bosh
wget -O /usr/bin/meta4 https://s3.amazonaws.com/dk-shared-assets/meta4-0.1.0-linux-amd64
echo "235bc60706793977446529830c2cb319e6aaf2da  /usr/bin/meta4" | shasum -c -
chmod +x /usr/bin/meta4
taskdir=$PWD
git clone releases-index-input releases-index
export GOPATH=$taskdir/worker
cd $GOPATH/src/worker
go run create-releases.go "$taskdir/release" "$taskdir/releases-index/%s" "%s" "((s3_endpoint))"
`,
									strings.TrimPrefix(string(r.URL), "https://"),
									minVersion,
								),
							},
						},
					},
					Ensure: &atc.PlanConfig{
						Put: "releases-index",
						Params: atc.Params{
							"repository": "releases-index",
							"rebase":     true,
						},
					},
				},
			},
		},
	)

	op.AddGroupJob("all", name)

	for _, group := range r.Categories {
		op.AddGroupJob(group, name)
	}
}
