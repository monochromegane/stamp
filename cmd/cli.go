package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/monochromegane/stamp/internal/config"
	"github.com/monochromegane/stamp/internal/configdir"
	"github.com/monochromegane/stamp/internal/stamp"
)

const cmdName = "stamp"

type PressCmd struct {
	Sheet  []string          `required:"" help:"Sheet name(s) from config directory (can specify multiple)" short:"s"`
	Dest   string            `optional:"" default:"." help:"Destination directory to copy to (default: current directory)" short:"d"`
	Config string            `optional:"" help:"Config directory path (overrides default)" short:"c"`
	Ext    string            `optional:"" default:".stamp" help:"Stamp file extension (default: .stamp)" short:"e"`
	Vars   map[string]string `arg:"" optional:"" help:"Template variables in KEY=VALUE format"`
}

func (c *PressCmd) Run(ctx *kong.Context) error {
	// 1. Resolve config directory
	configDir, err := configdir.GetConfigDirWithOverride(c.Config)
	if err != nil {
		return err
	}

	// 2. Resolve ALL sheet directories upfront
	srcDirs, err := configdir.ResolveTemplateDirs(configDir, c.Sheet)
	if err != nil {
		return err
	}

	// 3. Build merged variables with priority: CLI args > last sheet > ... > first sheet > global
	mergedVars, err := c.buildVariablesForMultipleTemplates(configDir)
	if err != nil {
		return err
	}

	// 4. Execute stamper with multiple sheets
	stamper := stamp.New(mergedVars, c.Ext)
	if err := stamper.ExecuteMultiple(srcDirs, c.Dest); err != nil {
		return fmt.Errorf("stamp failed: %w", err)
	}

	// 5. Print success message
	if len(c.Sheet) == 1 {
		fmt.Fprintf(os.Stdout, "Successfully stamped sheet '%s' to %s\n", c.Sheet[0], c.Dest)
	} else {
		fmt.Fprintf(os.Stdout, "Successfully stamped sheets %v to %s\n", c.Sheet, c.Dest)
	}
	return nil
}

// buildVariables implements four-level priority:
// 1. CLI args (highest priority)
// 2. Sheet-specific config
// 3. Global config
// 4. Hardcoded defaults (lowest priority - handled by stamp.New)
// Deprecated: Use buildVariablesForMultipleTemplates for new code
func (c *PressCmd) buildVariables(configDir string, templateName string) (map[string]string, error) {
	// Load hierarchical configs: global + sheet-specific
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
// 2. Last sheet's config
// 3. Middle sheets' configs
// 4. First sheet's config
// 5. Global config (lowest priority)
func (c *PressCmd) buildVariablesForMultipleTemplates(configDir string) (map[string]string, error) {
	// Load hierarchical configs: global + all sheets (in order)
	mergedVars, err := config.LoadHierarchicalMultiple(configDir, c.Sheet)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	// Override with CLI args (highest priority)
	for k, v := range c.Vars {
		mergedVars[k] = v
	}

	return mergedVars, nil
}

type CollectCmd struct {
	Sheet     string `required:"" help:"Sheet name to create" short:"s"`
	Source    string `arg:"" optional:"" default:"." help:"Source file or directory to collect (default: current directory)"`
	Config    string `optional:"" help:"Config directory path (overrides default)" short:"c"`
	Template  bool   `optional:"" help:"Treat collected files as templates (add .stamp extension)" short:"t"`
	Ext       string `optional:"" default:".stamp" help:"Template extension to add when --template is set (default: .stamp)" short:"e"`
	Recursive bool   `optional:"" default:"true" negatable:"" help:"Recursively copy directories (default: true, use --no-recursive to disable)" short:"r"`
}

func (c *CollectCmd) Run(ctx *kong.Context) error {
	// 1. Resolve config directory
	configDir, err := configdir.GetConfigDirWithOverride(c.Config)
	if err != nil {
		return err
	}

	// 2. Validate source path exists
	srcInfo, err := os.Stat(c.Source)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source path not found: %s", c.Source)
		}
		return fmt.Errorf("failed to stat source: %w", err)
	}

	// 3. Build destination: {configDir}/sheets/{Sheet}/
	destDir := filepath.Join(configDir, "sheets", c.Sheet)

	// 4. Check if sheet already exists
	if _, err := os.Stat(destDir); !os.IsNotExist(err) {
		return fmt.Errorf("sheet '%s' already exists at %s", c.Sheet, destDir)
	}

	// 5. Create destination directory
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create sheet directory: %w", err)
	}

	// 6. Copy files
	if srcInfo.IsDir() {
		if err := c.copyDirWithSkip(c.Source, destDir); err != nil {
			return err
		}
	} else {
		destPath := filepath.Join(destDir, filepath.Base(c.Source))
		if err := c.copyFileWithTemplate(c.Source, destPath); err != nil {
			return err
		}
	}

	// 7. Print success message
	fmt.Fprintf(os.Stdout, "Successfully collected to sheet '%s' at %s\n", c.Sheet, destDir)
	return nil
}

func (c *CollectCmd) copyDirWithSkip(src, dest string) error {
	// Non-recursive mode: only copy files directly in src directory
	if !c.Recursive {
		entries, err := os.ReadDir(src)
		if err != nil {
			return fmt.Errorf("failed to read directory: %w", err)
		}

		for _, entry := range entries {
			// Skip .git
			if entry.Name() == ".git" {
				continue
			}

			// Skip directories in non-recursive mode
			if entry.IsDir() {
				continue
			}

			srcPath := filepath.Join(src, entry.Name())
			destPath := filepath.Join(dest, entry.Name())
			if err := c.copyFileWithTemplate(srcPath, destPath); err != nil {
				return err
			}
		}
		return nil
	}

	// Recursive mode: use filepath.Walk
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		// Skip .git (both directory and file for git worktree support)
		if info.Name() == ".git" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil // Skip file
		}

		destPath := filepath.Join(dest, relPath)

		if info.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		return c.copyFileWithTemplate(path, destPath)
	})
}

func (c *CollectCmd) copyFileWithTemplate(src, dest string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", src, err)
	}

	// Add extension if template flag is set
	if c.Template {
		dest = dest + c.Ext
	}

	if err := os.WriteFile(dest, content, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", dest, err)
	}

	return nil
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
	Collect   CollectCmd       `cmd:"" help:"Collect directory or files as a new sheet"`
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
