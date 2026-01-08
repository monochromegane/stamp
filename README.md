# stamp

[![Actions Status](https://github.com/monochromegane/stamp/actions/workflows/test.yaml/badge.svg?branch=main)][actions]

[actions]: https://github.com/monochromegane/stamp/actions?workflow=test

A CLI tool for copying directory structures with Go template expansion.

## Features

- Copy directories with automatic template variable expansion
- Support for `.tmpl` files with Go template syntax
- Config directory with XDG Base Directory support
- Hierarchical configuration (global + template-specific)
- Multiple template directories with layered application
- YAML config file support for variable values
- Command-line variable overrides with priority system
- Strict template variable validation
- Support for `.tmpl.noop` files (copy templates without expansion)
- Simple key-value variable format

## Installation

## Quick Start

1. Create config directory and a template:
   ```bash
   # Linux
   mkdir -p ~/.config/stamp/templates/my-app
   echo "Hello {{.name}}!" > ~/.config/stamp/templates/my-app/hello.txt.tmpl

   # macOS
   mkdir -p ~/Library/Application\ Support/stamp/templates/my-app
   echo "Hello {{.name}}!" > ~/Library/Application\ Support/stamp/templates/my-app/hello.txt.tmpl
   ```

2. Use the template:
   ```bash
   stamp -t my-app name=alice
   ```

3. Check the result:
   ```bash
   cat hello.txt  # Output: Hello alice!
   ```

## Usage

### Config Directory Setup

stamp uses a centralized config directory to store templates and configurations.

**Default Location:**
- `$XDG_CONFIG_HOME/stamp` (if XDG_CONFIG_HOME is set)
- `$HOME/.config/stamp` (on Linux)
- `$HOME/Library/Application Support/stamp` (on macOS)
- `%AppData%\stamp` (on Windows)

**Directory Structure:**

Example (Linux):
```
~/.config/stamp/
├── stamp.yaml                    # Global config (optional)
└── templates/
    ├── go-cli/
    │   ├── main.go.tmpl
    │   ├── README.md.tmpl
    │   └── stamp.yaml            # Template-specific config (optional)
    └── web-app/
        ├── index.html.tmpl
        └── stamp.yaml
```

Example (macOS):
```
~/Library/Application Support/stamp/
├── stamp.yaml                    # Global config (optional)
└── templates/
    ├── go-cli/
    │   ├── main.go.tmpl
    │   ├── README.md.tmpl
    │   └── stamp.yaml            # Template-specific config (optional)
    └── web-app/
        ├── index.html.tmpl
        └── stamp.yaml
```

**Creating Templates:**

Linux:
```bash
# Create a new template directory
mkdir -p ~/.config/stamp/templates/my-template

# Add template files
echo "{{.message}}" > ~/.config/stamp/templates/my-template/output.txt.tmpl

# (Optional) Add template-specific config
cat > ~/.config/stamp/templates/my-template/stamp.yaml << EOF
message: "Default message"
version: "1.0.0"
EOF
```

macOS:
```bash
# Create a new template directory
mkdir -p ~/Library/Application\ Support/stamp/templates/my-template

# Add template files
echo "{{.message}}" > ~/Library/Application\ Support/stamp/templates/my-template/output.txt.tmpl

# (Optional) Add template-specific config
cat > ~/Library/Application\ Support/stamp/templates/my-template/stamp.yaml << EOF
message: "Default message"
version: "1.0.0"
EOF
```

**Global Config Format:**

Create a YAML file with flat key-value pairs:
- Linux: `~/.config/stamp/stamp.yaml`
- macOS: `~/Library/Application Support/stamp/stamp.yaml`

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
# Use a template from config directory (destination defaults to current directory)
stamp -t my-template name=alice

# Specify destination directory
stamp -t my-template -d ./output name=alice org=acme

# Use multiple templates (applied sequentially)
stamp -t base -t go-cli -d ./myproject name=bob

# Override config directory
stamp -t my-template -d ./output -c /custom/config/dir name=charlie
```

**Old syntax (still works):**
```bash
stamp press -t my-template -d ./output name=alice
```

### Template Files

**`.tmpl` files** are processed as Go templates. The `.tmpl` extension is removed from the output filename.

Example template file `hello.txt.tmpl`:
```
Hello {{.name}} from {{.org}}!
Welcome to the {{.repo}} project.
```

After processing with `name=alice` and `org=example`:
```
Hello alice from example!
Welcome to the myproject project.
```

**`.tmpl.noop` files** are copied without variable expansion, with only `.noop` removed.

Example use case - distributing template files:
```
Input:  config.yaml.tmpl.noop   (content: "name: {{.name}}")
Output: config.yaml.tmpl        (content: "name: {{.name}}" - not expanded)
```

This is useful when you want to distribute template files themselves rather than expanded content.

**Regular files** (without `.tmpl` extension) are copied as-is without template processing.

### Variable Priority

Variables are merged with the following priority (highest to lowest):

1. **Command-line arguments** - Variables specified as `key=value` on the command line
2. **Template-specific configs** - Variables defined in `templates/{name}/stamp.yaml` (when using multiple templates, later templates override earlier ones)
3. **Global config** - Variables defined in `stamp.yaml` in the config directory

**Example with multiple templates:**
```bash
# Global config: org=global-org
# templates/base/stamp.yaml: name=alice, version=1.0
# templates/go-cli/stamp.yaml: name=bob
# CLI args: name=charlie

stamp -t base -t go-cli name=charlie

# Result:
# name=charlie (from CLI - highest priority)
# version=1.0 (from base template)
# org=global-org (from global config - lowest priority)
```

**Example with config file override:**
```bash
# config.yaml in base template has name=alice
# Command line specifies name=charlie
# Result: name will be "charlie" (CLI overrides config)
stamp -t base -d ./dest name=charlie
```

### Advanced Features

#### Multiple Templates

You can apply multiple templates sequentially, with later templates overwriting files from earlier ones:

```bash
stamp -t base -t backend -t frontend -d ./myapp name=alice
```

**How it works:**
1. All templates are resolved and validated upfront
2. Variables are merged: CLI args > frontend config > backend config > base config > global config
3. Templates are applied in order: base → backend → frontend
4. If multiple templates contain the same file, the last one wins

**Use cases:**
- Layering: Start with a base template, add specialized features
- Composition: Combine independent components (backend + frontend)
- Overrides: Use later templates to override specific files from base templates

#### Strict Validation

All template variables are validated before execution. Missing variables will produce a helpful error:

```
Error: missing required template variables:

  - name
    used in:
      - hello.txt.tmpl
      - config.yaml.tmpl
  - version
    used in:
      - package.json.tmpl

Provide missing variables using:
  - Command line: stamp -t my-template name=<value> version=<value>
  - Config file: Create stamp.yaml in template or config directory
```

**Note:** Variables in `.tmpl.noop` files are NOT validated.

#### Custom Config Directory

Override the default config directory:

```bash
# Use custom directory
stamp -t my-template -c /path/to/configs -d ./output

# Use XDG_CONFIG_HOME environment variable
XDG_CONFIG_HOME=/custom/path stamp -t my-template
```

## License

MIT

## Author

[monochromegane](https://github.com/monochromegane)
