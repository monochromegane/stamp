package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/monochromegane/stamp/internal/stamp"
)

const cmdName = "stamp"

type PressCmd struct {
	Src  string            `required:"" help:"Source directory to copy from" short:"s"`
	Dest string            `required:"" help:"Destination directory to copy to" short:"d"`
	Vars map[string]string `arg:"" help:"Template variables in KEY=VALUE format"`
}

func (c *PressCmd) Run(ctx *kong.Context) error {
	stamper := stamp.New(c.Vars)
	if err := stamper.Execute(c.Src, c.Dest); err != nil {
		return fmt.Errorf("stamp failed: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Successfully stamped from %s to %s\n", c.Src, c.Dest)
	return nil
}

type CLI struct {
	Version kong.VersionFlag `help:"Show version"`
	Press   PressCmd         `cmd:"press" help:"Copy directory structure with template expansion"`
}

func NewCLI() *CLI {
	return &CLI{}
}

func (c *CLI) Execute(args []string) error {
	ctx := kong.Parse(c,
		kong.Name(cmdName),
		kong.Description("A CLI tool"),
		kong.UsageOnError(),
		kong.Vars{
			"version": fmt.Sprintf("%s v%s (rev:%s)", cmdName, version, revision),
		},
	)
	return ctx.Run()
}
