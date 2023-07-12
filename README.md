# worker
Deploys a set of Concourse pipelines to build releases for bosh.io. Follow usage instructions below to update release pipelines. These script will looks at the [releases repo](https://github.com/cloudfoundry/bosh-io-releases) to build pipelines by github org.

## Usage

git clone `worker` and `releases` to the same parent directory

note: `bin/test` will report `go vet` warnings because several `main` functions
are defined in the same package

```
$ fly -t bosh-io login -c https://ci.bosh-ecosystem.cf-app.com/ -n team-bosh-io
$ ./bin/sync-pipelines
```

## Notes

```
 schemaname |      relname      | n_live_tup
------------+-------------------+------------
 public     | checksums         |       5083
 public     | stemcell_notes    |         50
 public     | s3_stemcells      |          1
 public     | release_notes     |       1639
 public     | jobs              |       3779
 public     | release_versions  |       3777
 public     | release_tarballs  |       3776
```

```
$ psql db-name -t -c "select convert_from(key, 'utf-8'), convert_from(value, 'utf-8') from release_versions;"| go run import-release-versions.go ~/workspace/bosh-io/releases-index/
$ psql db-name -t -c "select convert_from(key, 'utf-8'), convert_from(value, 'utf-8') from jobs;"| go run import-release-jobs.go  ~/workspace/bosh-io/releases-index/
$ psql db-name -t -c "select convert_from(key, 'utf-8'), convert_from(value, 'utf-8') from release_tarballs;"| go run import-release-tarballs.go  ~/workspace/bosh-io/releases-index/
$ psql db-name -t -c "select convert_from(key, 'utf-8'), convert_from(value, 'utf-8') from release_notes;"| go run import-release-notes.go  ~/workspace/bosh-io/releases-index/
$ psql db-name -t -c "select convert_from(value, 'utf-8') from s3_stemcells;"| jq . > ~/workspace/bosh-io/stemcells-legacy-index/index.json
$ psql db-name -t -c "select convert_from(key, 'utf-8'), convert_from(value, 'utf-8') from stemcell_notes;"|go run import-stemcell-notes.go ~/workspace/bosh-io/stemcells-legacy-index/
```

## TODO

- docker image for release tpl pipeline
- have dedicated ci worker
- consolidate pipelines into org pipelines
