# go-whosonfirst-dist

Go package for working with Who's On First distributions.

## Install

You will need to have both `Go` (specifically [version 1.12](https://golang.org/dl/) or higher because we're using [Go modules](https://github.com/golang/go/wiki/Modules)) and the `make` programs installed on your computer. Assuming you do just type:

```
make tools
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

You will need to manually install the [Git LFS](https://git-lfs.github.com/) dependency for **wof-dist-build**.

## Important

This is work in progress. It seems to work, though. Until it doesn't. It still needs to be properly documented.

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
wof-dist-build --build-sqlite --preserve-checkout --workdir /usr/local/dist whosonfirst-data
```

```
wof-dist-build --build-sqlite --local-checkout --compress-all --workdir /usr/local/dist /usr/local/data/dist/whosonfirst-data
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
wof-dist-build --build-sqlite=false --build-bundle --local-sqlite --workdir /usr/local/dist whosonfirst-data
```

_`--preserve-sqlite` is assumed (and assumed to be true) if `--local-sqlite` is true._

## Tools

### wof-dist-build

Build one or more distribution files for a repository. _This is code that is actively been worked on so don't rely on it yet, or approach it accordingly._

```
./bin/wof-dist-build -h
Usage of ./bin/wof-dist-build:
  -build-bundle
    	Build a bundle distribution for a repo.
  -build-meta
    	Build meta files for a repo
  -build-sqlite
    	Build a (common) SQLite distribution for a repo (default true)
  -compress-all
    	... (default true)
  -compress-bundle
    	... (default true)
  -compress-max-cpus int
    	Number of concurrent processes to use when compressing distribution items. (default 4)
  -compress-meta
    	... (default true)
  -compress-sqlite
    	... (default true)
  -git-clone string
    	Indicate how to clone a repo, using either a native Git binary or the go-git implementation. Currently only the native Git binary is supported. (default "native")
  -git-organization string
    	Fetch repos from the user (or organization) (default "whosonfirst-data")
  -git-protocol string
    	Fetch repos using this protocol (default "https")
  -git-source string
    	Fetch repos from this endpoint (default "github.com")
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

$> ./bin/wof-dist-build -timings -verbose -workdir ./tmp -build-sqlite -build-meta whosonfirst-data-constituency-ca

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

### Combined

```
./bin/wof-dist-build -build-bundle -custom-repo -workdir /usr/local/data -preserve-checkout -local-checkout -index-alt-files -combined -combined-name sfomuseum-data-flights -timings -verbose sfomuseum-data-flights-2019-04 sfomuseum-data-flights-2019-05
```

## See also

* https://dist.whosonfirst.org
* https://github.com/whosonfirst/go-whosonfirst-dist-publish
* https://git-scm.com/
* https://git-lfs.github.com/
* https://github.com/src-d/go-git
