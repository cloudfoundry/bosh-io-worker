# worker

Currently implemented via Concourse pipelines.

## Usage

```
$ fly -t bosh-io login -c https://main.bosh-ci.cf-app.com -n bosh-io
$ source .envrc
$ lpass login <your accoutn>
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
