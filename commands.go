package main

import (
	"github.com/samuel/go-zookeeper/zk"
	
	"bufio"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"strconv"
	"time"
)

func servers() []string {
	s := os.Getenv("ZOOKEEPER_SERVERS")
	if s == "" {
		s = "127.0.0.1:2181"
	}
	return strings.Split(s, ",")
}
func connect() *zk.Conn {
	svs := servers()
	conn, _, err := zk.Connect(svs, time.Second)
	must(err)
	return conn
}

var (
	optWatch bool
	destDir string
)

var cmdExists = &Command{
	Usage: "exists <path> [--watch]",
	Short: "show if node exists",
	Long: `
Exists checks for a node at the given path and writes "y" or "n" to
stdout according to its presence. If --watch is used, waits for a
change in the presence of the node before exiting.

Example:

    $ zk exists /path
    y`,
	Run: runWatch,
}

func runWatch(cmd *Command, args []string) {
	if !(len(args) == 1) {
		failUsage(cmd)
	}
	nodePath := args[0]
	conn := connect()
	defer conn.Close()
	var events <-chan zk.Event
	var present bool
	var err error
	if !optWatch {
		present, _, err = conn.Exists(nodePath)

	} else {
		present, _, events, err = conn.ExistsW(nodePath)
	}
	must(err)
	if present {
		outString("y\n")
	} else {
		outString("n\n")
	}
	if events != nil {
		evt := <-events
		must(evt.Err)
	}
}

var cmdStat = &Command{
	Usage: "stat <path>",
	Short: "show node details",
	Long: `
Stat writes to stdout details of the node at the given path.

Example:

    $ zk stat /path
    Czxid:          337
    Mzxid:          460
    Ctime:          2014-05-17T08:11:24-07:00
    Mtime:          2014-05-17T14:49:45-07:00
    Version:        1
    Cversion:       3
    Aversion:       0
    EphemeralOwner: 0
    DataLength:     3
    Pzxid:          413`,
	Run: runStat,
}

func formatTime(millis int64) string {
	t := time.Unix(0, millis*1000000)
	return t.Format(time.RFC3339)
}

func runStat(cmd *Command, args []string) {
	if !(len(args) == 1) {
		failUsage(cmd)
	}
	nodePath := args[0]
	conn := connect()
	defer conn.Close()
	_, stat, err := conn.Get(nodePath)
	must(err)
	outString("Czxid:          %d\n", stat.Czxid)
	outString("Mzxid:          %d\n", stat.Mzxid)
	outString("Ctime:          %s\n", formatTime(stat.Ctime))
	outString("Mtime:          %s\n", formatTime(stat.Mtime))
	outString("Version:        %d\n", stat.Version)
	outString("Cversion:       %d\n", stat.Cversion)
	outString("Aversion:       %d\n", stat.Aversion)
	outString("EphemeralOwner: %d\n", stat.EphemeralOwner)
	outString("DataLength:     %d\n", stat.DataLength)
	outString("Pzxid:          %d\n", stat.Pzxid)
}

var cmdGet = &Command{
	Usage: "get <path> [--watch]",
	Short: "show node data",
	Long: `
Get reads the node data at the given path and writes it to stdout.
If --watch is used, waits for a change to the node before exiting.

Example:

    $ zk get /path
    content`,
	Run: runGet,
}

func runGet(cmd *Command, args []string) {
	if !(len(args) == 1) {
		failUsage(cmd)
	}
	nodePath := args[0]
	conn := connect()
	defer conn.Close()
	if !optWatch {
		data, _, err := conn.Get(nodePath)
		must(err)
		outData(data)
	} else {
		data, _, events, err := conn.GetW(nodePath)
		must(err)
		outData(data)
		evt := <-events
		must(evt.Err)
	}
}

var cmdCreate = &Command{
	Usage: "create <path>",
	Short: "create node with initial data",
	Long: `
Create makes a new node at the given path with the data given by
reading stdin.

Example:

    $ echo content | zk create /path`,
	Run: runCreate,
}

func runCreate(cmd *Command, args []string) {
	if !(len(args) == 1) {
		failUsage(cmd)
	}
	nodePath := args[0]
	conn := connect()
	defer conn.Close()
	data := inData()
	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	_, err := conn.Create(nodePath, data, flags, acl)
	must(err)
}

var cmdSet = &Command{
	Usage: "set <path> [version]",
	Short: "write node data",
	Long: `
Set updates the node at the given path with the data given by
reading stdin. If a version is given, submits that version with the
write request for verification against the current version.
Otherwise set will clobber the current data regardless of its
version.

Examples:

    $ echo new-content | zk set /path

    $ zk stat /path | grep Version
    Version:        3
    $ echo new-content | zk set /path 3`,
	Run: runSet,
}

func runSet(cmd *Command, args []string) {
	if !(len(args) == 1 || len(args) == 2) {
		failUsage(cmd)
	}
	nodePath := args[0]
	clobber := len(args) == 1
	conn := connect()
	defer conn.Close()
	data := inData()
	var version int32
	if clobber {
		version = -1
	} else {
		versionParsed, err := strconv.Atoi(args[1])
		must(err)
		version = int32(versionParsed)
	}
	_, err := conn.Set(nodePath, data, version)
	must(err)
}

var cmdDelete = &Command{
	Usage: "delete <path> [version]",
	Short: "delete node",
	Long: `
Delete removes the node at the given path. If a version is given,
submits that version with the delete request for verification
against the current version. Otherwise delete will clobber the
current data regardless of its version.

Examples:

    $ zk delete /path

    $ zk stat /path | grep Version
    Version:        7
    $ zk delete /path 7`,
	Run: runDelete,
}

func runDelete(cmd *Command, args []string) {
	if !(len(args) == 1 || len(args) == 2) {
		failUsage(cmd)
	}
	nodePath := args[0]
	clobber := len(args) == 1
	conn := connect()
	defer conn.Close()
	var version int32
	if clobber {
		version = -1
	} else {
		versionParsed, err := strconv.Atoi(args[1])
		must(err)
		version = int32(versionParsed)
	}
	err := conn.Delete(nodePath, version)
	must(err)
}

var cmdChildren = &Command{
	Usage: "children <path> [--watch]",
	Short: "list node children",
	Long: `
Children lists the names of the children of the node at the given
path, one name per line. If --watch is used, it waits for a change
in the names of given node's children before returning.

Example:

    $ zk children /people
    alice
    bob
    fred`,
	Run: runChildren,
}

func runChildren(cmd *Command, args []string) {
	if !(len(args) == 1) {
		failUsage(cmd)
	}
	nodePath := args[0]
	conn := connect()
	defer conn.Close()
	if !optWatch {
		children, _, err := conn.Children(nodePath)
		must(err)
		sort.Strings(children)
		for _, child := range children {
			outString("%s\n", child)
		}
	} else {
		children, _, events, err := conn.ChildrenW(nodePath)
		must(err)
		sort.Strings(children)
		for _, child := range children {
			outString("%s\n", child)
		}
		evt := <-events
		must(evt.Err)
	}
}

var cmdMirror = &Command{
	Usage: "mirror <path> [-d <destination>]",
	Short: "recursively get node children and data, and save to disk",
	Long: `
Mirror recursively gets nodes and data starting at the node at the 
given path, saving the result to file.

Example:

    $ zk mirror /people
    `,
	Run: runMirror,
}

type NodeTree struct {
	PathPrefix string
	Name string
	Children []*NodeTree
	Data []byte
}

func NewNodeTree(name string, pathPrefix string, data []byte) *NodeTree {
	return &NodeTree{
		PathPrefix: pathPrefix,
		Name: name,
		Children: make([]*NodeTree, 0),
		Data: data,
	}
}

func (self *NodeTree) AddChild(child *NodeTree) {
	self.Children = append(self.Children, child)
}

func (self *NodeTree) Print() {
	outString("%s\n", self.Name)
	if self.Data != nil {
		outString("%s\n", string(self.Data))
	}
	for _, node := range self.Children {
		node.Print()
	}
}

func saveNodeData(nodeFsPath string, nodeTree *NodeTree) error {
	var nodeFile *os.File
	var err error
	
	if nodeFile, err = os.Create(nodeFsPath); err != nil {
		return err
	}
	defer nodeFile.Close()

	writer := bufio.NewWriter(nodeFile)
	
	writer.Write(nodeTree.Data)

	writer.Flush()

	return nil
}

func saveNodeTree(dirPath string, nodeTree *NodeTree) error {
	var err error
	
	nodeFsPath := filepath.Join(dirPath, nodeTree.Name)
	
	if len(nodeTree.Children) == 0 {
		saveNodeData(nodeFsPath, nodeTree)
	} else {
		if err = os.Mkdir(nodeFsPath, os.ModeDir | 0755); err != nil {
			if os.IsExist(err) {
				// ok
			} else {
				return err
			}
		}

		if nodeTree.Data != nil && len(nodeTree.Data) != 0 {
			nodeDataPath := filepath.Join(nodeFsPath, "_data")
			if err = saveNodeData(nodeDataPath, nodeTree); err != nil {
				return err
			}
		}

		for _, child := range nodeTree.Children {
			if err = saveNodeTree(nodeFsPath, child); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func appendPath(nodePath string, name string) string {
	return path.Join(nodePath, name)
}

func mirrorNode(conn *zk.Conn, node *NodeTree) error {
	var err error
	var children []string
	
	nodePath := appendPath(node.PathPrefix, node.Name)
	
	if children, _, err = conn.Children(nodePath); err != nil {
		return err
	}
	
	for _, child := range children {
		var data []byte
		
		childPath := appendPath(nodePath, child)
		
		if data, _, err = conn.Get(childPath); err != nil {
			return err
		}

		childNode := NewNodeTree(child, nodePath, data)
		
		node.AddChild(childNode)

		if err := mirrorNode(conn, childNode); err != nil {
			return err
		}
	}

	return nil
}

func mirrorNodePath(conn *zk.Conn, nodePath string) (*NodeTree, error) {
	root := NewNodeTree("", nodePath, nil)
	
	if err := mirrorNode(conn, root); err != nil {
		return nil, err
	}

	return root, nil
}

func runMirror(cmd *Command, args []string) {
	if !(len(args) == 1) {
		failUsage(cmd)
	}
	nodePath := args[0]
	conn := connect()
	defer conn.Close()
	nodeTree, err := mirrorNodePath(conn, nodePath)
	must(err)
	var cwd string
	cwd, err = os.Getwd()
	must(err)
	err = saveNodeTree(filepath.Join(cwd, destDir), nodeTree)
	must(err)
}

func init() {
	cmdExists.Flag.BoolVarP(&optWatch, "watch", "w", false, "watch for a change to node presence before returning")
	cmdGet.Flag.BoolVarP(&optWatch, "watch", "w", false, "watch for a change to node state before returning")
	cmdChildren.Flag.BoolVarP(&optWatch, "watch", "w", false, "watch for a change to node children names before returning")
	cmdMirror.Flag.StringVarP(&destDir, "dest", "d", "", "destination directory in which to save node tree data")
}
