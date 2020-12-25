# go-whosonfirst-dist

Go package for working with Who's On First distributions.

## Important

This is work in progress and documentation is incomplete.

## Git

This package depends on there being a platform-specific `git` and `git-lfs` binaries present on the system where this is running. There is a branch of the code that uses the `go-git` package for cloning repositories but some Who's On First repos still require `lfs` support (hello `whosonfirst-data`...) It seems like that should be possible in `go-git` but if it is I haven't figured it out.

Ultimately I'd like to build everything on top of `go-git` because then we would have a proper pure-Go distribution tool which means we could build platform-native binaries with no extra depedencies. Today, everything depends on Git. 

## What is a "distribution".

Things like the [SQLite](https://dist.whosonfirst.org/sqlite/) databases or the "[bundles](https://dist.whosonfirst.org/bundles/)".

## Build process(es)

_These are works in progress. I am still trying to work it out..._

### Building SQLite databases

* Fetch remote clone or use local checkout
* Build databases, generate indices
* Compress databases, optionally preserve uncompressed
* Clean up (remote clone or use local checkout)

For example:

```
wof-dist-build --build-sqlite-common --preserve-checkout --workdir /usr/local/dist whosonfirst-data
```

```
wof-dist-build --build-sqlite-common --local-checkout --compress-all --workdir /usr/local/dist /usr/local/data/dist/whosonfirst-data
```

_`--preserve-checkout` is assumed (and assumed to be true) if `--local-checkout` is true._

### Building meta file(s)

_Please write me._

### Building bundles

* Fetch remote clone or use remote (compressed) SQLite database or use local (uncompressed) SQLite database
* Generate metafiles
* Build bundle, generate index
* Compress bundle, optionally preserve uncompressed
* Clean up (remote clone or remote (compressed) SQLite database or local (uncompressed) SQLite database)

```
$> wof-dist-build -build-sqlite-common=false -build-bundle -local-sqlite -workdir /usr/local/dist whosonfirst-data
```

_`--preserve-sqlite` is assumed (and assumed to be true) if `--local-sqlite` is true._

## Tools

```
$> make cli
go build -mod vendor -o bin/wof-dist-build cmd/wof-dist-build/main.go
go build -mod vendor -o bin/wof-dist-fetch cmd/wof-dist-fetch/main.go
```

### wof-dist-build

Build one or more distribution files for a repository. _This is code that is actively been worked on so don't rely on it yet, or approach it accordingly._

```
$> ./bin/wof-dist-build -h
Usage of ./bin/wof-dist-build:
  -build-bundle
    	Build a bundle distribution for a repo.
  -build-meta
    	Build meta files for a repo
  -build-sqlite
    	Build a (common) SQLite distribution for a repo. This flag is DEPRECATED. (default true)
  -build-sqlite-all
    	Build a SQLite distribution for a repo, with all tables defined by the other -build-sqlite flags.
  -build-sqlite-common
    	Build a SQLite distribution for a repo, with common tables. (default true)
  -build-sqlite-rtree
    	Build a SQLite distribution for a repo, with rtree-related tables.
  -build-sqlite-search
    	Build a (common) SQLite distribution for a repo, with search-tables.
  -combined
    	Create a single combined distribution from multiple repos.
  -combined-name string
    	Distribution name for a single combined distribution from multiple repos.
  -compress-all
    	... (default true)
  -compress-bundle
    	... (default true)
  -compress-max-cpus int
    	Number of concurrent processes to use when compressing distribution items. (default 2)
  -compress-meta
    	... (default true)
  -compress-sqlite
    	... (default true)
  -custom-repo
    	Allow custom repo names (default true)
  -git-clone string
    	Indicate how to clone a repo, using either a native Git binary or the go-git implementation. Currently only the native Git binary is supported. (default "native")
  -git-organization string
    	Fetch repos from the user (or organization) (default "whosonfirst-data")
  -git-protocol string
    	Fetch repos using this protocol (default "https")
  -git-source string
    	Fetch repos from this endpoint (default "github.com")
  -index-alt-files
    	Index alternate geometry files.
  -index-relations
    	Index the records related to a feature, specifically wof:belongsto, wof:depicts and wof:involves. Alt files for relations are not indexed at this time.
  -index-relations-reader-uri string
    	A valid go-reader.Reader URI from which to read data for a relations candidate.
  -local-checkout
    	Do not fetch a repo from a remote source but instead use a local checkout on disk
  -local-sqlite
    	Do not build a new SQLite database but use a pre-existing database on disk (this expects to find the database at the same path it would be stored if the database were created from scratch)
  -preserve-all
    	...
  -preserve-bundle
    	...
  -preserve-checkout
    	Do not remove repo from disk after the build process is complete. This is automatically set to true if the -local-checkout flag is true.
  -preserve-meta
    	...
  -preserve-sqlite
    	...
  -strict
    	...
  -timings
    	Display timings during the build process
  -verbose
    	Be chatty
  -workdir string
    	Where to store temporary and final build files. If empty the code will attempt to use the current working directory.
```

For example:

```
$> mkdir tmp

$> ./bin/wof-dist-build -timings -verbose -workdir ./tmp -build-sqlite-common -build-meta whosonfirst-data-constituency-ca

13:09:59.008358 [wof-dist-build] STATUS git lfs clone --depth 1 https://github.com/whosonfirst-data/whosonfirst-data-constituency-ca.git tmp/whosonfirst-data-constituency-ca
13:10:04.008358 [wof-dist-build] STATUS time to clone whosonfirst-data-constituency-ca 4.425250127s
13:10:04.008388 [wof-dist-build] STATUS LOCAL tmp/whosonfirst-data-constituency-ca
13:10:07.620093 [wof-dist-build] STATUS CREATED tmp/whosonfirst-data-constituency-ca-latest.db
13:10:07.620126 [wof-dist-build] STATUS BUILD METAFILE sqlite tmp/whosonfirst-data-constituency-ca-latest.db
2018/06/11 13:10:08 time to prepare tmp/whosonfirst-data-constituency-ca-latest.db 874.109451ms
2018/06/11 13:10:08 time to prepare all 55 records 874.133191ms
13:10:08.613818 [wof-dist-build] STATUS time to build all 9.030797929s

$> ls -al ./tmp
-rw-r--r--  1 asc  staff  33390592 Jun 11 13:10 whosonfirst-data-constituency-ca-latest.db
-rw-r--r--  1 asc  staff     17895 Jun 11 13:10 wof-constituency-ca-latest.csv
```

#### "Combined" distributions

It is also possible to create a single combined distribution from two or more repos, passing the `-combined` and `-combined-name` flag.

Here's an example that in addition to creating a combined distributions, also assumes local and non-standard repositories, builds a "bundle" distribution and indexes alternate geometry files.

_Note that as of this writing alternate geometry files are _not_ supported for either bundles or (CSV) meta files. They will be but today that are only indexed in SQLite databases._

```
$> ./bin/wof-dist-build \
	-build-bundle -custom-repo -preserve-checkout -local-checkout -index-alt-files \
	-timings -verbose \
	-workdir /usr/local/data \
	-combined -combined-name sfomuseum-data-flights \
	sfomuseum-data-flights-2019-04 sfomuseum-data-flights-2019-05

go build -o bin/wof-dist-build cmd/wof-dist-build/main.go
go build -o bin/wof-dist-fetch cmd/wof-dist-fetch/main.go
15:24:20.003232 [wof-dist-build] STATUS local_checkouts are [/usr/local/data/sfomuseum-data-flights-2019-04 /usr/local/data/sfomuseum-data-flights-2019-05]
15:24:20.105579 [wof-dist-build] STATUS commit hashes are map[sfomuseum-data-flights-2019-04:bd913977adef7a56a5d236046ff878b261d7f289 sfomuseum-data-flights-2019-05:2a12dc4085353ea65423f09ec1369e4c6d6d6426] ([/usr/local/data/sfomuseum-data-flights-2019-04 /usr/local/data/sfomuseum-data-flights-2019-05])
15:25:20.122712 [wof-dist-build] STATUS time to index ancestors (61608) : 23.027587622s
15:25:20.122774 [wof-dist-build] STATUS time to index concordances (61608) : 3.051047626s
15:25:20.123193 [wof-dist-build] STATUS time to index geojson (61608) : 8.833956578s
15:25:20.123215 [wof-dist-build] STATUS time to index spr (61608) : 17.115882932s
15:25:20.123221 [wof-dist-build] STATUS time to index names (61608) : 6.171740731s
15:25:20.123226 [wof-dist-build] STATUS time to index all (61608) : 1m0.004661436s
15:26:20.123918 [wof-dist-build] STATUS time to index names (110124) : 11.724367851s
15:26:20.123986 [wof-dist-build] STATUS time to index ancestors (110124) : 41.615312492s
15:26:20.124001 [wof-dist-build] STATUS time to index concordances (110124) : 6.179600527s
15:26:20.124007 [wof-dist-build] STATUS time to index geojson (110124) : 17.307942063s
15:26:20.124012 [wof-dist-build] STATUS time to index spr (110124) : 39.327491403s
15:26:20.124018 [wof-dist-build] STATUS time to index all (110124) : 2m0.005381573s
15:27:20.127816 [wof-dist-build] STATUS time to index geojson (155162) : 25.833516964s
15:27:20.127834 [wof-dist-build] STATUS time to index spr (155162) : 1m1.9488226s
15:27:20.127854 [wof-dist-build] STATUS time to index names (155162) : 17.371926717s
15:27:20.127860 [wof-dist-build] STATUS time to index ancestors (155162) : 59.038325554s
15:27:20.127864 [wof-dist-build] STATUS time to index concordances (155162) : 9.573363568s
15:27:20.127868 [wof-dist-build] STATUS time to index all (155162) : 3m0.008763864s
15:27:48.281906 [wof-dist-build] STATUS Built  without any reported errors
15:27:48.281945 [wof-dist-build] STATUS local sqlite is /usr/local/data/sfomuseum-data-flights-latest.db
15:27:48.281973 [wof-dist-build] STATUS build metafile from sqlite ([/usr/local/data/sfomuseum-data-flights-latest.db])
2019/06/03 15:28:34 time to prepare /usr/local/data/sfomuseum-data-flights-latest.db 45.957440286s
2019/06/03 15:28:34 time to prepare all 154858 records 45.957729135s
15:28:35.683017 [wof-dist-build] STATUS time to build metafiles (/usr/local/data/sfomuseum-data-flights.csv) 47.400595263s
15:32:04.141721 [wof-dist-build] STATUS time to build bundles () 3m28.456791321s
15:32:04.141747 [wof-dist-build] STATUS time to build UNCOMPRESSED distributions for sfomuseum-data-flights 7m44.134569605s
15:32:04.144406 [wof-dist-build] STATUS register function to compress /usr/local/data/sfomuseum-data-flights-latest.db
15:32:04.144602 [wof-dist-build] STATUS time to wait to start compressing /usr/local/data/sfomuseum-data-flights-latest.db 581ns
15:32:04.144530 [wof-dist-build] STATUS register function to compress /usr/local/data/sfomuseum-data-flights.csv
15:32:04.145585 [wof-dist-build] STATUS time to wait to start compressing /usr/local/data/sfomuseum-data-flights.csv 437ns
15:32:04.144563 [wof-dist-build] STATUS register function to compress /usr/local/data/sfomuseum-data-flights-latest
15:32:20.892664 [wof-dist-build] STATUS All done compressing /usr/local/data/sfomuseum-data-flights.csv (throttle)
15:32:20.892781 [wof-dist-build] STATUS time to compress /usr/local/data/sfomuseum-data-flights.csv 16.747045767s
15:32:20.892860 [wof-dist-build] STATUS All done compressing /usr/local/data/sfomuseum-data-flights.csv
15:32:20.892820 [wof-dist-build] STATUS time to wait to start compressing /usr/local/data/sfomuseum-data-flights-latest 16.746488168s
15:34:21.574356 [wof-dist-build] STATUS All done compressing /usr/local/data/sfomuseum-data-flights-latest.db (throttle)
15:34:21.574376 [wof-dist-build] STATUS time to compress /usr/local/data/sfomuseum-data-flights-latest.db 2m17.428545702s
15:34:21.574380 [wof-dist-build] STATUS All done compressing /usr/local/data/sfomuseum-data-flights-latest.db
15:35:08.325796 [wof-dist-build] STATUS All done compressing /usr/local/data/sfomuseum-data-flights-latest (throttle)
15:35:08.325817 [wof-dist-build] STATUS time to compress /usr/local/data/sfomuseum-data-flights-latest 3m4.1779794s
15:35:08.325822 [wof-dist-build] STATUS All done compressing /usr/local/data/sfomuseum-data-flights-latest
15:35:08.325854 [wof-dist-build] STATUS remove uncompressed file /usr/local/data/sfomuseum-data-flights-latest.db
15:35:08.325873 [wof-dist-build] STATUS remove uncompressed file /usr/local/data/sfomuseum-data-flights.csv
15:35:08.325861 [wof-dist-build] STATUS remove uncompressed file /usr/local/data/sfomuseum-data-flights-latest
15:35:51.580311 [wof-dist-build] STATUS time to remove uncompressed files for sfomuseum-data-flights 43.254038216s
15:35:51.580393 [wof-dist-build] STATUS time to build COMPRESSED distributions for sfomuseum-data-flights 11m31.571211707s
15:35:51.580542 [wof-dist-build] STATUS time to build distributions for 2 repos 11m31.571501299s
15:35:51.581774 [wof-dist-build] STATUS Wrote inventory /usr/local/data/sfomuseum-data-flights-inventory.json

$> cat /usr/local/data/sfomuseum-data-flights-inventory.json 
[
  {
    "name": "sfomuseum-data-flights.csv",
    "type": "x-urn:whosonfirst:csv:meta#event",
    "name_compressed": "sfomuseum-data-flights.csv.bz2",
    "count": 154860,
    "size": 45900340,
    "size_compressed": 7550967,
    "sha256_compressed": "cf2023e9f895f5f9671ebbb280983149b4aa09dfecfc71c967576da5750b4de6",
    "last_updated": "2019-05-17T11:19:14-07:00",
    "last_modified": "2019-06-03T15:28:34-07:00",
    "repo": "sfomuseum-data-flights-2019-04:sfomuseum-data-flights-2019-05",
    "commit": "bd913977adef7a56a5d236046ff878b261d7f289:2a12dc4085353ea65423f09ec1369e4c6d6d6426"
  }, 
  {
    "name": "sfomuseum-data-flights-latest.db",
    "type": "x-urn:whosonfirst:database:sqlite#common",
    "name_compressed": "sfomuseum-data-flights-latest.db.bz2",
    "count": 166469,
    "size": 943042560,
    "size_compressed": 105789001,
    "sha256_compressed": "1c1afc5f337cea024da5e4cd198e67ccb3dbe9c45ed31dfdafb505f2a0a1bc4d",
    "last_updated": "2019-05-17T11:19:14-07:00",
    "last_modified": "2019-06-03T15:27:47-07:00",
    "repo": "sfomuseum-data-flights-2019-04:sfomuseum-data-flights-2019-05",
    "commit": "bd913977adef7a56a5d236046ff878b261d7f289:2a12dc4085353ea65423f09ec1369e4c6d6d6426"
  }, 
  {
    "name": "sfomuseum-data-flights-latest",
    "type": "x-urn:whosonfirst:fs:bundle#sfomuseum-data-flights-latest",
    "name_compressed": "sfomuseum-data-flights-latest.tar.bz2",
    "count": 154860,
    "size": 383617201,
    "size_compressed": 27602413,
    "sha256_compressed": "d6b74885c70107a35cc4c4a8b8707036b55aae74e875cf263d720a4ea67926f4",
    "last_updated": "2019-05-17T11:19:14-07:00",
    "last_modified": "2019-06-03T15:31:00-07:00",
    "repo": "sfomuseum-data-flights-2019-04:sfomuseum-data-flights-2019-05",
    "commit": "bd913977adef7a56a5d236046ff878b261d7f289:2a12dc4085353ea65423f09ec1369e4c6d6d6426"
  }
]
```

## See also

* https://dist.whosonfirst.org
* https://github.com/whosonfirst/go-whosonfirst-dist-publish
* https://git-scm.com/
* https://git-lfs.github.com/
* https://github.com/src-d/go-git
