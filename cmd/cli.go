package cmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/monochromegane/stamp/internal/config"
	"github.com/monochromegane/stamp/internal/configdir"
	"github.com/monochromegane/stamp/internal/stamp"
)

const cmdName = "stamp"

type PressCmd struct {
	Template string            `required:"" help:"Template name from config directory" short:"t"`
	Dest     string            `optional:"" default:"." help:"Destination directory to copy to (default: current directory)" short:"d"`
	Config   string            `optional:"" help:"Config directory path (overrides default)" short:"c"`
	Vars     map[string]string `arg:"" optional:"" help:"Template variables in KEY=VALUE format"`
}

func (c *PressCmd) Run(ctx *kong.Context) error {
	// 1. Resolve config directory
	configDir, err := configdir.GetConfigDirWithOverride(c.Config)
	if err != nil {
		return err
	}

	// 2. Resolve template source directory
	srcDir, err := configdir.ResolveTemplateDir(configDir, c.Template)
	if err != nil {
		return err
	}

	// 3. Build merged variables with priority: CLI args > template config > global config
	mergedVars, err := c.buildVariables(configDir)
	if err != nil {
		return err
	}

	// 4. Execute stamper
	stamper := stamp.New(mergedVars)
	if err := stamper.Execute(srcDir, c.Dest); err != nil {
		return fmt.Errorf("stamp failed: %w", err)
	}

	// 5. Print success message
	fmt.Fprintf(os.Stdout, "Successfully stamped template '%s' to %s\n", c.Template, c.Dest)
	return nil
}

// buildVariables implements four-level priority:
// 1. CLI args (highest priority)
// 2. Template-specific config
// 3. Global config
// 4. Hardcoded defaults (lowest priority - handled by stamp.New)
func (c *PressCmd) buildVariables(configDir string) (map[string]string, error) {
	// Load hierarchical configs: global + template-specific
	mergedVars, err := config.LoadHierarchical(configDir, c.Template)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// Override with CLI args (highest priority)
	for k, v := range c.Vars {
		mergedVars[k] = v
	}

	return mergedVars, nil
}

type CLI struct {
	Version kong.VersionFlag `help:"Show version"`
	Press   PressCmd         `cmd:"" default:"withargs" help:"Copy directory structure with template expansion"`
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
