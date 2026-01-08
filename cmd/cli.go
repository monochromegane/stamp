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
	Template []string          `required:"" help:"Template name(s) from config directory (can specify multiple)" short:"t"`
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

	// 2. Resolve ALL template directories upfront
	srcDirs, err := configdir.ResolveTemplateDirs(configDir, c.Template)
	if err != nil {
		return err
	}

	// 3. Build merged variables with priority: CLI args > last template > ... > first template > global
	mergedVars, err := c.buildVariablesForMultipleTemplates(configDir)
	if err != nil {
		return err
	}

	// 4. Execute stamper with multiple templates
	stamper := stamp.New(mergedVars)
	if err := stamper.ExecuteMultiple(srcDirs, c.Dest); err != nil {
		return fmt.Errorf("stamp failed: %w", err)
	}

	// 5. Print success message
	if len(c.Template) == 1 {
		fmt.Fprintf(os.Stdout, "Successfully stamped template '%s' to %s\n", c.Template[0], c.Dest)
	} else {
		fmt.Fprintf(os.Stdout, "Successfully stamped templates %v to %s\n", c.Template, c.Dest)
	}
	return nil
}

// buildVariables implements four-level priority:
// 1. CLI args (highest priority)
// 2. Template-specific config
// 3. Global config
// 4. Hardcoded defaults (lowest priority - handled by stamp.New)
// Deprecated: Use buildVariablesForMultipleTemplates for new code
func (c *PressCmd) buildVariables(configDir string, templateName string) (map[string]string, error) {
	// Load hierarchical configs: global + template-specific
	mergedVars, err := config.LoadHierarchical(configDir, templateName)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// Override with CLI args (highest priority)
	for k, v := range c.Vars {
		mergedVars[k] = v
	}

	return mergedVars, nil
}

// buildVariablesForMultipleTemplates implements hierarchical priority:
// 1. CLI args (highest priority)
// 2. Last template's config
// 3. Middle templates' configs
// 4. First template's config
// 5. Global config (lowest priority)
func (c *PressCmd) buildVariablesForMultipleTemplates(configDir string) (map[string]string, error) {
	// Load hierarchical configs: global + all templates (in order)
	mergedVars, err := config.LoadHierarchicalMultiple(configDir, c.Template)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// Override with CLI args (highest priority)
	for k, v := range c.Vars {
		mergedVars[k] = v
	}

	return mergedVars, nil
}

type ConfigDirCmd struct {
	Config string `optional:"" help:"Config directory path (overrides default)" short:"c"`
}

func (c *ConfigDirCmd) Run(ctx *kong.Context) error {
	configDir, err := configdir.GetConfigDirWithOverride(c.Config)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "%s\n", configDir)
	return nil
}

type CLI struct {
	Version   kong.VersionFlag `help:"Show version"`
	Press     PressCmd         `cmd:"" default:"withargs" help:"Copy directory structure with template expansion"`
	ConfigDir ConfigDirCmd     `cmd:"" help:"Print config directory path"`
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
