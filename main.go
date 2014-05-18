package main

import (
	"fmt"
	"github.com/bgentry/pflag"
	"io/ioutil"
	"os"
	"strings"
)

type Command struct {
	Usage string
	Flag  pflag.FlagSet
	Short string
	Long  string
	Run   func(cmd *Command, args []string)
}

func (c *Command) PrintUsage() {
	fmt.Fprintf(os.Stderr, "Usage: zk %s\n", c.Usage)
	fmt.Fprintf(os.Stderr, "Run 'zk help %s' for details.\n", c.Name())
}

func (c *Command) PrintLongUsage() {
	fmt.Printf("Usage: zk %s\n\n", c.Usage)
	fmt.Println(strings.Trim(c.Long, "\n"))
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

var commands = []*Command{
	cmdExists,
	cmdStat,
	cmdGet,
	cmdCreate,
	cmdSet,
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
			if err := cmd.Flag.Parse(args[1:]); err == pflag.ErrHelp {
				cmdHelp.Run(cmdHelp, args[:1])
				return
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
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

func inData() []byte {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}
	return data
}

func outString(p string, args ...interface{}) {
	_, err := fmt.Fprintf(os.Stdout, p, args...)
	must(err)
}

func outData(d []byte) {
	_, err := os.Stdout.Write(d)
	must(err)
}

func must(err error) {
	errString := strings.TrimPrefix(err.Error(), "zk: ")
	fmt.Fprintf(os.Stderr, "error: %s\n", errString)
	os.Exit(1)
}

func failUsage(cmd *Command) {
	cmd.PrintUsage()
	os.Exit(2)
}
