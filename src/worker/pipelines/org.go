package pipelines

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bosh-io/worker/src/worker/releases"
	"github.com/concourse/concourse/atc"
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
						"uri":         "git@github.com:cloudfoundry/bosh-io-releases-index.git",
						"branch":      "main",
						"private_key": "((github_deploy_key_releases-index.private_key))",
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
	pipelineBytes, err := json.Marshal(o.pipeline)
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
	fiveMinuteDuration, _ := time.ParseDuration("5m")

	if r.MinVersion != "" {
		minVersion = r.MinVersion
	}

	op.pipeline.Resources = append(
		op.pipeline.Resources,
		atc.ResourceConfig{
			Name:       repoResourceName,
			Type:       "git",
			CheckEvery: &atc.CheckEvery{Interval: fiveMinuteDuration},
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
			PlanSequence: []atc.Step{
				{
					Config: &atc.InParallelStep{
						Config: atc.InParallelConfig{
							Steps: []atc.Step{
								{
									Config: &atc.GetStep{
										Name: "worker",
									},
								},
								{
									Config: &atc.GetStep{
										Name:     "release",
										Resource: repoResourceName,
										Trigger:  true,
										Params: atc.Params{
											"submodules": "none",
										},
									},
								},
								{
									Config: &atc.GetStep{
										Name: "releases-index",
									},
								},
							},
						},
					},
				},
				{
					Config: &atc.EnsureStep{
						Hook: atc.Step{
							Config: &atc.PutStep{
								Name: "releases-index",
								Params: atc.Params{
									"repository": "releases-index",
									"rebase":     true,
								},
							},
						},
						Step: &atc.TaskStep{
							Name: "sync",
							Params: atc.TaskEnv{
								"AWS_ACCESS_KEY_ID":     "((worker-release-tarballs-uploader_assume_aws_access_key.username))",
								"AWS_SECRET_ACCESS_KEY": "((worker-release-tarballs-uploader_assume_aws_access_key.password))",
							},
							Config: &atc.TaskConfig{
								Platform: "linux",
								ImageResource: &atc.ImageResource{
									Type: "docker-image",
									Source: atc.Source{
										"repository": "bosh/integration",
									},
								},
								Inputs: []atc.TaskInputConfig{
									atc.TaskInputConfig{
										Name: "release",
									},
									atc.TaskInputConfig{
										Name: "releases-index",
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
taskdir=$PWD
cd worker/src/worker
go run create-releases.go "$taskdir/release" "$taskdir/releases-index/%s" "%s" "s3://s3-external-1.amazonaws.com/boshio-release-tarballs"
`,
											strings.TrimPrefix(string(r.URL), "https://"),
											minVersion,
										),
									},
								},
							},
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
