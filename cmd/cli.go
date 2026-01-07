package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/monochromegane/stamp/internal/config"
	"github.com/monochromegane/stamp/internal/stamp"
)

const cmdName = "stamp"

type PressCmd struct {
	Src    string            `required:"" help:"Source directory to copy from" short:"s"`
	Dest   string            `required:"" help:"Destination directory to copy to" short:"d"`
	Config string            `optional:"" help:"Path to YAML config file" short:"c"`
	Vars   map[string]string `arg:"" optional:"" help:"Template variables in KEY=VALUE format"`
}

func (c *PressCmd) Run(ctx *kong.Context) error {
	// Build merged variables with priority: CLI args > config file > defaults
	mergedVars, err := c.buildVariables()
	if err != nil {
		return err
	}

	stamper := stamp.New(mergedVars)
	if err := stamper.Execute(c.Src, c.Dest); err != nil {
		return fmt.Errorf("stamp failed: %w", err)
	}
	fmt.Fprintf(os.Stdout, "Successfully stamped from %s to %s\n", c.Src, c.Dest)
	return nil
}

// buildVariables implements three-level priority:
// 1. CLI args (highest priority)
// 2. Config file values
// 3. Hardcoded defaults (lowest priority - handled by stamp.New)
func (c *PressCmd) buildVariables() (map[string]string, error) {
	mergedVars := make(map[string]string)

	// Layer 1: Load config file if specified
	if c.Config != "" {
		configVars, err := config.Load(c.Config)
		if err != nil {
			return nil, fmt.Errorf("config file error: %w", err)
		}
		for k, v := range configVars {
			mergedVars[k] = v
		}
	}

	// Layer 2: Override with CLI args (highest priority)
	for k, v := range c.Vars {
		mergedVars[k] = v
	}

	return mergedVars, nil
}

type CLI struct {
	Version kong.VersionFlag `help:"Show version"`
	Press   PressCmd         `cmd:"press" help:"Copy directory structure with template expansion"`
}

func NewCLI() *CLI {
	return &CLI{}
}

func (c *CLI) Execute(args []string) error {
	parser := kong.Must(c,
		kong.Name(cmdName),
		kong.Description("A CLI tool"),
		kong.UsageOnError(),
		kong.Vars{
			"version": fmt.Sprintf("%s v%s (rev:%s)", cmdName, version, revision),
		},
	)
	ctx, err := parser.Parse(args)
	if err != nil {
		return err
	}
	return ctx.Run()
}
