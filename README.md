# go-whosonfirst-dist

Go package for working with Who's On First distributions.

## Install

You will need to have both `Go` (specifically a version of Go more recent than 1.6 so let's just assume you need [Go 1.8](https://golang.org/dl/) or higher) and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Important

This is work in progress. It's probably not worth trying to use yet unless you're me.

## What is a "distribution".

Things like the [SQLite](https://dist.whosonfirst.org/sqlite/) databases or the "[bundles](https://dist.whosonfirst.org/bundles/)".

## Tools

### wof-dist-fetch

Fetch all (or some) of the files listed in a distribution inventory file. `wof-dist-fetch` will uncompress files as they are written to disk (it's possible this should or needs to be an optional flag...)

```
./bin/wof-dist-fetch -h
Usage of ./bin/wof-dist-fetch:
  -dest string
    	Where distribution files should be written (default "/tmp")
  -exclude value
    	A valid regular expression for comparing against an item's filename, for exclusion
  -include value
    	A valid regular expression for comparing against an item's filename, for inclusion
  -inventory string
    	The URL of a valid distribution inventory file (default "https://dist.whosonfirst.org/sqlite/inventory.json")
```

For example:

```
./bin/wof-dist-fetch -include '.*constituency-.*-latest.*'
2018/02/14 15:46:28 WROTE /tmp/whosonfirst-data-constituency-ca-latest.db (33398784 bytes)
2018/02/14 15:47:29 WROTE /tmp/whosonfirst-data-constituency-us-latest.db (1147850752 bytes)
```

## See also

* https://whosonfirst.mapzen.com/sqlite
* https://sqlite.org/
