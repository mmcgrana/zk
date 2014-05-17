package main

import (
	"fmt"
	"github.com/bgentry/pflag"
	"github.com/samuel/go-zookeeper/zk"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Command struct {
	Usage string
	Flag  pflag.FlagSet
	Short string
	Long  string
	Run   func(cmd *Command, args []string)
}

func (c *Command) PrintUsage() {
	if c.Runnable() {
		fmt.Fprintf(os.Stderr, "Usage: hk %s\n", c.FullUsage())
	}
	fmt.Fprintf(os.Stderr, "Use 'hk help %s' for more information.\n", c.Name())
}

func (c *Command) PrintLongUsage() {
	if c.Runnable() {
		fmt.Printf("Usage: hk %s\n\n", c.FullUsage())
	}
	fmt.Println(strings.Trim(c.Long, "\n"))
}

func (c *Command) FullUsage() string {
	if c.NeedsApp {
		return c.Name() + " [-a <app or remote>]" + strings.TrimPrefix(c.Usage, c.Name())
	}
	return c.Usage
}

func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

const extra = " (extra)"

func (c *Command) List() bool {
	return c.Short != "" && !strings.HasSuffix(c.Short, extra)
}

func (c *Command) ListAsExtra() bool {
	return c.Short != "" && strings.HasSuffix(c.Short, extra)
}

func (c *Command) ShortExtra() string {
	return c.Short[:len(c.Short)-len(extra)]
}

// func printUsageTo(..., ...) {
// 	...
// }
//
// func printError(...) {
//
// }

var commands = []*Command{
	cmdExists,
	cmdStat,
	cmdGet,
	cmdCreate,
	mdSet,
	cmdDelete,
	cmdChildren,
}

func main() {
	args := os.Args[1:]
	if len(args) < 1 || strings.IndexRune(args[0], '-') == 0 {
		printUsageTo(os.Stderr)
		os.Exit(2)
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.Flag.SetInterspersed(true)
			cmd.Flag.Usage = cmd.PrintUsage
			if err := cmd.Flag.Parse(args[1:]); err == flag.ErrHelp {
				cmdHelp.Run(cmdHelp, args[:1])
				return
			} else if err != nil {
				printError(err.Error())
				os.Exit(2)
			}
			cmd.Run(cmd, cmd.Flag.Args())
			return
		}
	}

	fmt.Fprintf(os.Stderr, "error: unrecognized command: %s\n", args[0])
	fmt.Fprintf(os.Stderr, "Run 'zk help' for usage.\n")
	os.Exit(2)
}

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
	t := time.Unix(0, millis*1000000)
	return t.Format(time.RFC3339)
}
