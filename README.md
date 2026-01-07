# stamp

[![Actions Status](https://github.com/monochromegane/stamp/actions/workflows/test.yaml/badge.svg?branch=main)][actions]

[actions]: https://github.com/monochromegane/stamp/actions?workflow=test

A CLI tool for copying directory structures with Go template expansion.

## Features

- Copy directories with automatic template variable expansion
- Support for `.tmpl` files with Go template syntax
- YAML config file support for default variable values
- Command-line variable overrides with priority system
- Simple key-value variable format

## Usage

### Basic Usage

```bash
# Copy directory with default variables
stamp press -s ./source -d ./destination

# Provide variables via command line
stamp press -s ./source -d ./destination name=bob org=acme

# Use config file for default variables
stamp press -s ./source -d ./destination -c config.yaml

# Override config file values with command-line arguments
stamp press -s ./source -d ./destination -c config.yaml name=charlie
```

### Config File Format

Create a YAML file with flat key-value pairs:

```yaml
# config.yaml
name: alice
org: example
repo: myproject
version: 1.0.0
```

Use the config file with the `-c` or `--config` flag:

```bash
stamp press -s ./template -d ./output -c config.yaml
```

### Variable Priority

Variables are merged with the following priority (highest to lowest):

1. **Command-line arguments** - Variables specified as `key=value` on the command line
2. **Config file** - Variables defined in the YAML config file (via `-c` flag)
3. **Hardcoded defaults** - Built-in default values (currently `name: "alice"`)

Example:
```bash
# config.yaml has name=bob
# Command line specifies name=charlie
# Result: name will be "charlie" (CLI overrides config)
stamp press -s ./src -d ./dest -c config.yaml name=charlie
```

### Template Files

Files ending with `.tmpl` are processed as Go templates. The `.tmpl` extension is removed from the output filename.

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

Regular files (without `.tmpl` extension) are copied as-is without template processing.

## Install

## License

MIT

## Author

[monochromegane](https://github.com/monochromegane)
