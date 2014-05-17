package main

import (
	"fmt"
	"io/ioutil"
	"github.com/samuel/go-zookeeper/zk"
	"github.com/spf13/cobra"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

func connect() *zk.Conn {
	conn, _, err := zk.Connect([]string{"127.0.0.1:2181"}, time.Second)
	if err != nil {
		panic(err)
	}
	return conn
}

func input() []byte {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	return data
}

func fail(err error) {
	errString := strings.TrimPrefix(err.Error(), "zk: ")
	fmt.Fprintf(os.Stderr, "error: %s\n", errString)
	os.Exit(1)
}

func must(err error) {
	if err != nil {
		fail(err)
	}
}

func failUsage(cmd *cobra.Command) {
	cmd.Usage()
	os.Exit(2)
}

func outputBool(b bool) {
	var out string
	if b {
		out = "y"
	} else {
		out = "n"
	}
	fmt.Fprintln(os.Stdout, out)
}

func formatTime(millis int64) string {
	t := time.Unix(0, millis * 1000000)
	return t.Format(time.RFC3339)
}

func main() {
	cmdRoot := &cobra.Command{Use: "zk"}

	var cmdExistsWatch bool
	cmdExists := &cobra.Command{
		Use: "exists <path> [--watch]",
		Short: "show whether a node exists",
		Long: `
Exists checks for a node at the given path and writes "y" or "n" to
stdout according to its presence. If --watch is used, waits for a
change in the presence of the node before exiting.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1) {
				failUsage(cmd)
			}
			path := args[0]
			conn := connect()
			if !cmdExistsWatch {
				present, _, err := conn.Exists(path)
				must(err)
				outputBool(present)
			} else {
				present, _, events, err := conn.ExistsW(path)
				must(err)
				outputBool(present)
				evt := <-events
				must(evt.Err)
			}
		},
	}

	cmdStat := &cobra.Command {
		Use: "stat <path>",
		Short: "show node details",
		Long: `
Stat writes to stdout details of the node at the given path.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1) {
				failUsage(cmd)
			}
			path := args[0]
			conn := connect()
			_, stat, err := conn.Get(path)
			must(err)
			fmt.Fprintf(os.Stdout, "Czxid:          %d\n", stat.Czxid)
			fmt.Fprintf(os.Stdout, "Mzxid:          %d\n", stat.Mzxid)
			fmt.Fprintf(os.Stdout, "Ctime:          %s\n", formatTime(stat.Ctime))
			fmt.Fprintf(os.Stdout, "Mtime:          %s\n", formatTime(stat.Mtime))
			fmt.Fprintf(os.Stdout, "Version:        %d\n", stat.Version)
			fmt.Fprintf(os.Stdout, "Cversion:       %d\n", stat.Cversion)
			fmt.Fprintf(os.Stdout, "Aversion:       %d\n", stat.Aversion)
			fmt.Fprintf(os.Stdout, "EphemeralOwner: %d\n", stat.EphemeralOwner)
			fmt.Fprintf(os.Stdout, "DataLength:     %d\n", stat.DataLength)
			fmt.Fprintf(os.Stdout, "Pzxid:          %d\n", stat.Pzxid)
		},
	}
	
	var cmdGetWatch bool
	cmdGet := &cobra.Command{
		Use:   "get <path> [--watch]",
		Short: "show node data",
		Long: `
Get reads the node data at the given path and writes it to stdout. If
--watch is used, waits for a change to the node before exiting.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1) {
				failUsage(cmd)
			}
			path := args[0]
			conn := connect()
			if !cmdGetWatch {
				data, _, err := conn.Get(path)
				must(err)
				_, err = os.Stdout.Write(data)
				must(err)
			} else {
				data, _, events, err := conn.GetW(path)
				must(err)
				_, err = os.Stdout.Write(data)
				must(err)
				evt := <- events
				must(evt.Err)
			}
			
		},
	}
	cmdGet.Flags().BoolVarP(&cmdGetWatch, "watch", "w", false, "watch for a change before returning")

	cmdCreate := &cobra.Command {
		Use: "create <path>",
		Short: "create node with initial data",
		Long: `
Create makes a new node at the given path with the data given by
reading stdin.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1) {
				failUsage(cmd)
			}
			path := args[0]
			conn := connect()
			data := input() 
			flags := int32(0)
			acl := zk.WorldACL(zk.PermAll)
			_, err := conn.Create(path, data, flags, acl)
			must(err)
		},
	}

	cmdSet := &cobra.Command {
		Use: "set <path> [version]",
		Short: "set node data",
		Long: `
Set updates the node at the given path with the data given by reading
stdin. If a version is given, submits that version with the write
request for verification, otherwise reads the current version before
attempting a write.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1 || len(args) == 2) {
				failUsage(cmd)
			}
			path := args[0]
			readVersion := len(args) == 1
			conn := connect()
			data := input()
			var version int32
			if readVersion {
				_, stat, err := conn.Get(path)
				must(err)
				version = stat.Version
			} else {
				versionParsed, err := strconv.Atoi(args[1])
				must(err)
				version = int32(versionParsed)
			}
			_, err := conn.Set(path, data, version)
			must(err)
		},
	}

	cmdDelete := &cobra.Command {
		Use: "delete <path> [version]",
		Short: "delete node",
		Long: `
Delete removes the node at the given path. If a version is given,
submits that version with the write request for verification,
otherwise reads the current version before attempting a write.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1 || len(args) == 2) {
				failUsage(cmd)
			}
			path := args[0]
			readVersion := len(args) == 1
			conn := connect()
			var version int32
			if readVersion {
				_, stat, err := conn.Get(path)
				must(err)
				version = stat.Version
			} else {
				versionParsed, err := strconv.Atoi(args[1])
				must(err)
				version = int32(versionParsed)
			}
			err := conn.Delete(path, version)
			must(err)
		},
	}

	var cmdChildrenWatch bool
	cmdChildren := &cobra.Command {
		Use: "children <path> [--watch]",
		Short: "list children of a node",
		Long: `
Children lists the names of the children of the node at the given
path, one name per line. If --watch is used, waits for a change in the
names of given node's children before returning.`,
		Run: func(cmd *cobra.Command, args []string) {
			if !(len(args) == 1) {
				failUsage(cmd)
			}
			path := args[0]
			conn := connect()
			if !cmdChildrenWatch {
				children, _, err := conn.Children(path)
				must(err)
				sort.Strings(children)
				for _, child := range children {
					fmt.Fprintln(os.Stdout, child)
				}
			} else {
				children, _, events, err := conn.ChildrenW(path)
				must(err)
				sort.Strings(children)
				for _, child := range children {
					fmt.Fprintln(os.Stdout, child)
				}
				evt := <- events
				must(evt.Err)
			}
		},
	}
	cmdChildren.Flags().BoolVarP(&cmdChildrenWatch, "watch", "w", false, "watch for a change before returning")

	cmdRoot.AddCommand(cmdExists, cmdStat, cmdGet, cmdCreate, cmdSet, cmdDelete, cmdChildren)
	cmdRoot.Execute()
}
