# stamp

[![Actions Status](https://github.com/monochromegane/stamp/actions/workflows/test.yaml/badge.svg?branch=main)][actions]

[actions]: https://github.com/monochromegane/stamp/actions?workflow=test

A CLI tool for copying directory structures with Go template expansion.

## Features

- Copy directories with automatic template variable expansion
- **Sheets**: A sheet is a directory that contains multiple stamps (template files and regular files) applied together as a unit
- Support for `.stamp` files with Go template syntax
- Configurable stamp file extension (default: `.stamp`, customizable via `--ext`)
- Config directory with XDG Base Directory support
- Global configuration with command-line overrides
- Multiple sheet directories with layered application
- YAML config file support for variable values
- Command-line variable overrides with priority system
- Strict template variable validation
- Support for `.stamp.noop` files (copy sheets without expansion)
- Simple key-value variable format

## Installation

### Homebrew

```bash
brew tap monochromegane/tap
brew install monochromegane/tap/stamp
```

### Go Install

```bash
go install github.com/monochromegane/stamp@latest
```

## Quick Start

1. Create config directory and a sheet:
   ```bash
   mkdir -p "$(stamp config-dir)/sheets/my-app"
   echo "Hello {{.name}}!" > "$(stamp config-dir)/sheets/my-app/hello.txt.stamp"
   ```

2. Use the sheet:
   ```bash
   stamp -s my-app name=alice
   ```

3. Check the result:
   ```bash
   cat hello.txt  # Output: Hello alice!
   ```

Or collect an existing directory as a sheet:
   ```bash
   # Collect current directory as a sheet
   stamp collect -s my-project

   # Collect specific directory
   stamp collect -s my-template /path/to/directory

   # Collect as template (adds .stamp extension to files)
   stamp collect -s my-template -t /path/to/directory
   ```

## Usage

### Config Directory Setup

stamp uses a centralized config directory to store sheets and configurations.

**Default Location:**
- `$XDG_CONFIG_HOME/stamp` (if XDG_CONFIG_HOME is set)
- Platform-specific default (use `stamp config-dir` to see your path)
  - Linux: `$HOME/.config/stamp`
  - macOS: `$HOME/Library/Application Support/stamp`
  - Windows: `%AppData%\stamp`

**Checking your config directory:**
```bash
stamp config-dir
```

**Directory Structure:**

```
$(stamp config-dir)/
├── stamp.yaml                    # Global config (optional)
└── sheets/
    ├── go-cli/
    │   ├── main.go.stamp
    │   └── README.md.stamp
    └── web-app/
        └── index.html.stamp
```

Note: Run `stamp config-dir` to see your actual config directory path.

**Creating Templates:**

```bash
# Create a new sheet directory
mkdir -p "$(stamp config-dir)/sheets/my-template"

# Add stamp files
echo "{{.message}}" > "$(stamp config-dir)/sheets/my-template/output.txt.stamp"
```

**Global Config Format:**

Create a YAML file at `$(stamp config-dir)/stamp.yaml` with flat key-value pairs:

```yaml
# stamp.yaml
name: alice
org: example
repo: myproject
version: 1.0.0
```

### Basic Usage

**Note:** The `press` subcommand is now the default, so you can omit it.

```bash
# Use a sheet from config directory (destination defaults to current directory)
stamp -s my-template name=alice

# Specify destination directory
stamp -s my-template -d ./output name=alice org=acme

# Use multiple sheets (applied sequentially)
stamp -s base -s go-cli -d ./myproject name=bob

# Override config directory
stamp -s my-template -d ./output -c /custom/config/dir name=charlie
```

**Old syntax (still works):**
```bash
stamp press -s my-template -d ./output name=alice
```

**Custom sheet extension:**
```bash
# Use .stamp extension instead of .stamp (useful for chezmoi compatibility)
stamp -s my-template --ext .stamp name=alice

# Use .tpl extension
stamp -s my-template -e .tpl name=alice
```

### Stamp Files

**`.stamp` files** are processed as Go templates. The `.stamp` extension is removed from the output filename.

Example stamp file `hello.txt.stamp`:
```
Hello {{.name}} from {{.org}}!
Welcome to the {{.repo}} project.
```

After processing with `name=alice` and `org=example`:
```
Hello alice from example!
Welcome to the myproject project.
```

**`.stamp.noop` files** are copied without variable expansion, with only `.noop` removed.

Example use case - distributing stamp files:
```
Input:  config.yaml.stamp.noop   (content: "name: {{.name}}")
Output: config.yaml.stamp        (content: "name: {{.name}}" - not expanded)
```

This is useful when you want to distribute stamp files themselves rather than expanded content.

**Regular files** (without `.stamp` extension) are copied as-is without sheet processing.

### Variable Priority

Variables are merged with the following priority (highest to lowest):

1. **Command-line arguments** - Variables specified as `key=value` on the command line
2. **Global config** - Variables defined in `stamp.yaml` in the config directory

**Note:** Sheet-specific configs (`sheets/{name}/stamp.yaml`) are no longer supported. All configuration should be placed in the global `stamp.yaml` file.

**Example with global config:**
```bash
# Global config: org=global-org, name=alice
# CLI args: name=charlie

stamp -s my-template name=charlie

# Result:
# name=charlie (from CLI - highest priority)
# org=global-org (from global config)
```

**Example with config file override:**
```bash
# Global config has name=alice
# Command line specifies name=charlie
# Result: name will be "charlie" (CLI overrides config)
stamp -s base -d ./dest name=charlie
```

### Advanced Features

#### Multiple Templates

You can apply multiple sheets sequentially, with later sheets overwriting files from earlier ones:

```bash
stamp -s base -s backend -s frontend -d ./myapp name=alice
```

**How it works:**
1. All sheets are resolved and validated upfront
2. Variables are merged: CLI args > global config
3. Templates are applied in order: base → backend → frontend
4. If multiple sheets contain the same file, the last one wins

**Use cases:**
- Layering: Start with a base sheet, add specialized features
- Composition: Combine independent components (backend + frontend)
- Overrides: Use later sheets to override specific files from base sheets

#### Strict Validation

All template variables are validated before execution. Missing variables will produce a helpful error:

```
Error: missing required template variables:

  - name
    used in:
      - hello.txt.stamp
      - config.yaml.stamp
  - version
    used in:
      - package.json.stamp

Provide missing variables using:
  - Command line: stamp -s my-template name=<value> version=<value>
  - Config file: Create stamp.yaml in the config directory
```

**Note:** Variables in `.stamp.noop` files are NOT validated.

#### Custom Config Directory

Override the default config directory:

```bash
# Use custom directory
stamp -s my-template -c /path/to/configs -d ./output

# Use XDG_CONFIG_HOME environment variable
XDG_CONFIG_HOME=/custom/path stamp -s my-template
```

#### Config Directory Command

Use the `config-dir` subcommand to get the config directory path:

```bash
# Print config directory path
stamp config-dir

# Use with command substitution
mkdir -p "$(stamp config-dir)/sheets/my-app"

# Override with custom directory
stamp config-dir -c /custom/config
```

This is useful for:
- Creating sheets programmatically
- Shell scripts
- Platform-independent documentation

## License

MIT

## Author

[monochromegane](https://github.com/monochromegane)
