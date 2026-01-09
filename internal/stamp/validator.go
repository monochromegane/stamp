package stamp

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template/parse"
)

// ValidationError represents missing template variables with detailed context
type ValidationError struct {
	MissingVars map[string][]string // map[variableName][]templateFilePaths
}

func (e *ValidationError) Error() string {
	if len(e.MissingVars) == 0 {
		return "template validation failed"
	}

	var sb strings.Builder
	sb.WriteString("missing required template variables:\n\n")

	// Sort variable names for consistent output
	varNames := make([]string, 0, len(e.MissingVars))
	for name := range e.MissingVars {
		varNames = append(varNames, name)
	}
	sort.Strings(varNames)

	// Format each missing variable with its usage locations
	for _, varName := range varNames {
		templates := e.MissingVars[varName]
		fmt.Fprintf(&sb, "  - %s\n", varName)
		sb.WriteString("    used in:\n")
		for _, tmpl := range templates {
			fmt.Fprintf(&sb, "      - %s\n", tmpl)
		}
	}

	sb.WriteString("\nProvide missing variables using:\n")
	sb.WriteString("  - Command line: stamp -s <sheet> -d <dest> ")
	for _, varName := range varNames {
		fmt.Fprintf(&sb, "%s=<value> ", varName)
	}
	sb.WriteString("\n")
	sb.WriteString("  - Config file: Create stamp.yaml in sheet or config directory\n")

	return sb.String()
}

// validateTemplateVars scans all .tmpl files and validates required variables are provided
func (s *Stamper) validateTemplateVars(srcDir string) error {
	return s.validateMultipleTemplateVars([]string{srcDir})
}

// validateMultipleTemplateVars scans all template directories and validates variables
func (s *Stamper) validateMultipleTemplateVars(srcDirs []string) error {
	// Map to track: variableName -> []templatePaths across all templates
	varUsage := make(map[string][]string)

	// Scan all template directories
	for _, srcDir := range srcDirs {
		if err := s.collectTemplateVars(srcDir, varUsage); err != nil {
			return err
		}
	}

	// Check if any required variables are missing
	missingVars := make(map[string][]string)
	for varName, templatePaths := range varUsage {
		if _, exists := s.templateVars[varName]; !exists {
			missingVars[varName] = templatePaths
		}
	}

	// Return error if any variables are missing
	if len(missingVars) > 0 {
		return &ValidationError{MissingVars: missingVars}
	}

	return nil
}

// collectTemplateVars walks a directory and collects variable usage
func (s *Stamper) collectTemplateVars(srcDir string, varUsage map[string][]string) error {
	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-template files
		if info.IsDir() || s.isTmplNoopFile(path) || !strings.HasSuffix(path, s.templateExt) {
			return nil
		}

		// Extract variables from this template
		// If template is invalid, let it fail during normal processing
		vars, err := extractTemplateVars(path)
		if err != nil {
			return nil
		}

		// Track which templates use which variables
		relPath, _ := filepath.Rel(srcDir, path)
		for _, v := range vars {
			varUsage[v] = append(varUsage[v], relPath)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to scan templates: %w", err)
	}

	return nil
}

// extractTemplateVars extracts all variables from a template file
func extractTemplateVars(templatePath string) ([]string, error) {
	// Read template content
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template: %w", err)
	}

	// Parse template to get AST
	tree, err := parse.New(filepath.Base(templatePath)).Parse(string(content), "{{", "}}", make(map[string]*parse.Tree))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Extract unique variables
	vars := make(map[string]struct{})
	if tree.Root != nil {
		walkNode(tree.Root, &vars)
	}

	// Convert to sorted slice
	result := make([]string, 0, len(vars))
	for v := range vars {
		result = append(result, v)
	}
	sort.Strings(result)
	return result, nil
}

// walkNode recursively walks the AST to find FieldNodes
func walkNode(node parse.Node, vars *map[string]struct{}) {
	if node == nil {
		return
	}

	switch n := node.(type) {
	case *parse.FieldNode:
		// Extract first field: .name or .org (ignore chained fields like .org.repo)
		if len(n.Ident) > 0 {
			(*vars)[n.Ident[0]] = struct{}{}
		}

	case *parse.ListNode:
		// Recursively process all nodes in list
		if n.Nodes != nil {
			for _, node := range n.Nodes {
				walkNode(node, vars)
			}
		}

	case *parse.ActionNode:
		// Process pipeline
		if n.Pipe != nil {
			walkNode(n.Pipe, vars)
		}

	case *parse.PipeNode:
		// Process all commands in pipeline
		if n.Cmds != nil {
			for _, cmd := range n.Cmds {
				walkNode(cmd, vars)
			}
		}

	case *parse.CommandNode:
		// Process all arguments (can contain FieldNodes)
		if n.Args != nil {
			for _, arg := range n.Args {
				walkNode(arg, vars)
			}
		}

	case *parse.IfNode:
		walkBranchNode(&n.BranchNode, vars)

	case *parse.RangeNode:
		walkBranchNode(&n.BranchNode, vars)

	case *parse.WithNode:
		walkBranchNode(&n.BranchNode, vars)

	case *parse.TemplateNode:
		// Process template invocation pipeline
		if n.Pipe != nil {
			walkNode(n.Pipe, vars)
		}

		// Other node types (TextNode, NumberNode, StringNode, etc.) don't contain variables
	}
}

// walkBranchNode walks branch nodes (if, range, with)
func walkBranchNode(branch *parse.BranchNode, vars *map[string]struct{}) {
	// Process condition
	if branch.Pipe != nil {
		walkNode(branch.Pipe, vars)
	}
	// Process if-branch
	if branch.List != nil {
		walkNode(branch.List, vars)
	}
	// Process else-branch
	if branch.ElseList != nil {
		walkNode(branch.ElseList, vars)
	}
}
