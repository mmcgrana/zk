## zk

`zk` is a command line client to the [Zookeeper](http://zookeeper.apache.org/)
distributed storage service designed to be fast, easy to install,
and Unix-friendly. It's written in Go.

### Getting Started

Install with:

```console
$ go get github.com/mmcgrana/zk
```

And use with e.g.:

```console
$ echo data | zk create /path

$ zk get /path
data

$ echo new-data | zk set /path

$ zk get /path
new-data

$ zk children /
path
namespace
zookeeper
```

### Usage Details

Use `zk help` or `zk help <command>` to see full usage details:

```console
$ zk help
Usage: zk <command> [arguments] [options]

Commands:

    exists      show if node exists
    stat        show node details
    get         show node data
    create      create node with initial data
    set         write node data
    delete      delete node
    children    list node children
    help        show help

Run 'zk help <command>' for details.
```

### Watches

Commands `exists`, `get`, and `children` accept `--watch` options
to trigger the installation of corresponding watches on the
requested node. For example:

```console
$ bash -c "sleep 10; echo second-value | zk set /key"

$ zk get /key --watch   # pauses for ~10s, then returns
first-value

$ zk get /key           # returns immediately
second-value
```

### Server Configuration

By default the client targets `127.0.0.1:2181`. To configure one or
more different Zookeepers to target, export `ZOOKEPER_SERVERS` in
`host:port` format with a `,` between each server. For example:

```console
$ export ZOOKEEPER_HOSTS=23.22.49.116:2181,23.20.114.164:2181,54.197.120.188:2181
$ zk ...
```

### Other Zookeeper CLIs

You may be interested in these other Zookeper command line clients:

* [The built-in `zkCLI.sh`](http://zookeeper.apache.org/doc/trunk/zookeeperStarted.html)
* [org.linkedin.zookeeper-cli](https://github.com/pongasoft/utils-zookeeper)
* [com.loopfor.zookeeper.cli](https://github.com/davidledwards/zookeeper/tree/master/zookeeper-cli)

### hk Lineage

The `zk` project borrows much of its CLI scaffolding and therefore
CLI aesthetic from the [`hk`](https://github.com/heroku/hk) project.
Like `hk`, `zk` is designed to behave like a standard Unix tool and
be composed with other such tools.

Copyright of the borrowed portions of `zk`'s code remains with the
`hk` project.

### Contributing

Please see [CONTRIBUTING.md](contributing.md).

### License

Please see [LICENSE.md](LICENSE.md)
