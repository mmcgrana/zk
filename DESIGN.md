```
$ zk
$ zk help
$ zk -h
$ zk --help
Usage: zk <command> [arguments] [options]                            | out
                                                                     | out
Commands:                                                            | out
                                                                     | out
    exists      show if node exists                                  | out
    stat        show node details                                    | out
    get         show node data                                       | out
    create      create node with initial data                        | out
    set         write node data                                      | out
    delete      delete node                                          | out
    children    list node children                                   | out
    help        show help                                            | out
                                                                     | out
Run 'zk help <command>' for details.                                 | out 2, (zk help = 0)
```

```
$ zk help get
Usage: zk get <path> [--watch]                                       | out
                                                                     | out
Get reads the node data at the given path and writes it to stdout.   | out
If --watch is used, waits for a change to the node before exiting.   | out
                                                                     | out
Example:                                                             | out
                                                                     | out
    $ zk get /path                                                   | out
    bar                                                              | out 0
```

```
$ zk get /path
content                                                              | out 0
```

```
$ zk get /path --watch
content                                                              | out 0
```

```
$ zk get                                                             | err
Usage: zk get <path> [--watch]                                       | err                                                                       | err
Run 'zk help get' for details.                                       | err 2
```

```
$ echo "try" | zk create /path
error: node already exists                                           | err 1
```

```
$ zk wat
error: unrecognized command: wat                                     | err
Run 'zk help' for usage.                                             | err 2
