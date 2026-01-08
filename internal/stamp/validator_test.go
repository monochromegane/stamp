package stamp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExtractTemplateVars_SimpleVariable tests basic variable extraction
func TestExtractTemplateVars_SimpleVariable(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl", "Hello {{.name}}!")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	expected := []string{"name"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_MultipleVariables tests multiple variables
func TestExtractTemplateVars_MultipleVariables(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl",
		"{{.name}} from {{.org}} repo {{.repo}}")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	expected := []string{"name", "org", "repo"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_IfBlock tests variables in conditional blocks
func TestExtractTemplateVars_IfBlock(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl",
		"{{if .enabled}}{{.name}} is enabled{{end}}")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	expected := []string{"enabled", "name"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_RangeBlock tests variables in range blocks
func TestExtractTemplateVars_RangeBlock(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl",
		"{{range .items}}{{.}}{{end}}")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	expected := []string{"items"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_WithBlock tests variables in with blocks
func TestExtractTemplateVars_WithBlock(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl",
		"{{with .config}}{{.value}}{{end}}")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	expected := []string{"config", "value"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_ChainedFields tests chained field access
func TestExtractTemplateVars_ChainedFields(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl",
		"{{.user.name}} {{.user.email}}")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	// Should only extract top-level field "user", not "name" or "email"
	expected := []string{"user"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_NoVariables tests template without variables
func TestExtractTemplateVars_NoVariables(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl", "Static content only")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	if len(vars) != 0 {
		t.Errorf("expected no variables, got %v", vars)
	}
}

// TestExtractTemplateVars_InvalidTemplate tests handling of invalid templates
func TestExtractTemplateVars_InvalidTemplate(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl", "Invalid {{.name")

	_, err := extractTemplateVars(tmplPath)
	if err == nil {
		t.Fatal("extractTemplateVars() should return error for invalid template")
	}
}

// TestExtractTemplateVars_DuplicateVariables tests variables used multiple times
func TestExtractTemplateVars_DuplicateVariables(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl",
		"{{.name}} and {{.name}} again")

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	// Should deduplicate
	expected := []string{"name"}
	assertVarsEqual(t, vars, expected)
}

// TestExtractTemplateVars_ComplexNesting tests complex nested structures
func TestExtractTemplateVars_ComplexNesting(t *testing.T) {
	dir := t.TempDir()
	tmplPath := createTestFile(t, dir, "test.tmpl", `
{{if .enabled}}
  {{range .items}}
    {{with .config}}
      {{.value}}
    {{end}}
  {{end}}
{{else}}
  {{.fallback}}
{{end}}`)

	vars, err := extractTemplateVars(tmplPath)
	if err != nil {
		t.Fatalf("extractTemplateVars() failed: %v", err)
	}

	expected := []string{"config", "enabled", "fallback", "items", "value"}
	assertVarsEqual(t, vars, expected)
}

// TestValidateTemplateVars_AllProvided tests validation when all vars provided
func TestValidateTemplateVars_AllProvided(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "hello.tmpl", "Hello {{.name}}!")

	stamper := New(map[string]string{"name": "alice"}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	if err != nil {
		t.Errorf("validateTemplateVars() should pass, got error: %v", err)
	}
}

// TestValidateTemplateVars_MissingVariable tests validation failure
func TestValidateTemplateVars_MissingVariable(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "hello.tmpl", "Hello {{.name}}!")

	stamper := New(map[string]string{}, ".tmpl") // No variables provided
	err := stamper.validateTemplateVars(src)

	if err == nil {
		t.Fatal("validateTemplateVars() should return error for missing variable")
	}

	// Check error contains variable name
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("error should mention missing variable 'name', got: %v", err)
	}
}

// TestValidateTemplateVars_MultipleMissingVariables tests multiple missing vars
func TestValidateTemplateVars_MultipleMissingVariables(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "info.tmpl", "{{.name}} from {{.org}}/{{.repo}}")

	stamper := New(map[string]string{"name": "alice"}, ".tmpl") // Only name provided
	err := stamper.validateTemplateVars(src)

	if err == nil {
		t.Fatal("validateTemplateVars() should return error for missing variables")
	}

	// Check error contains both missing variables
	errMsg := err.Error()
	if !strings.Contains(errMsg, "org") {
		t.Errorf("error should mention missing variable 'org', got: %v", err)
	}
	if !strings.Contains(errMsg, "repo") {
		t.Errorf("error should mention missing variable 'repo', got: %v", err)
	}
}

// TestValidateTemplateVars_MultipleTemplates tests multiple template files
func TestValidateTemplateVars_MultipleTemplates(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "hello.tmpl", "Hello {{.name}}!")
	createTestFile(t, src, "info.tmpl", "From {{.org}}")

	stamper := New(map[string]string{}, ".tmpl") // No variables
	err := stamper.validateTemplateVars(src)

	if err == nil {
		t.Fatal("validateTemplateVars() should return error")
	}

	// Verify error lists both template files
	errMsg := err.Error()
	if !strings.Contains(errMsg, "hello.tmpl") {
		t.Errorf("error should mention hello.tmpl, got: %v", err)
	}
	if !strings.Contains(errMsg, "info.tmpl") {
		t.Errorf("error should mention info.tmpl, got: %v", err)
	}
}

// TestValidateTemplateVars_NestedTemplates tests templates in subdirectories
func TestValidateTemplateVars_NestedTemplates(t *testing.T) {
	src := t.TempDir()
	subdir := filepath.Join(src, "subdir")
	os.MkdirAll(subdir, 0755)

	createTestFile(t, src, "root.tmpl", "{{.name}}")
	createTestFile(t, subdir, "nested.tmpl", "{{.org}}")

	stamper := New(map[string]string{"name": "alice"}, ".tmpl") // Missing 'org'
	err := stamper.validateTemplateVars(src)

	if err == nil {
		t.Fatal("validateTemplateVars() should return error")
	}

	// Should mention the nested template
	if !strings.Contains(err.Error(), "subdir") {
		t.Errorf("error should mention nested template path, got: %v", err)
	}
}

// TestValidateTemplateVars_NoTemplates tests directory without templates
func TestValidateTemplateVars_NoTemplates(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "readme.md", "No templates here")

	stamper := New(map[string]string{}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	if err != nil {
		t.Errorf("validateTemplateVars() should pass with no templates, got: %v", err)
	}
}

// TestExecute_ValidationFailsBeforeCopy tests that validation happens first
func TestExecute_ValidationFailsBeforeCopy(t *testing.T) {
	src := t.TempDir()
	dest := t.TempDir()

	createTestFile(t, src, "hello.tmpl", "Hello {{.name}}!")
	createTestFile(t, src, "static.txt", "Static file")

	stamper := New(map[string]string{}, ".tmpl") // Missing 'name'
	err := stamper.Execute(src, dest)

	if err == nil {
		t.Fatal("Execute() should fail validation")
	}

	// Verify no files were copied (dest should only have the temp directory)
	entries, _ := os.ReadDir(dest)
	if len(entries) > 0 {
		t.Errorf("no files should be copied when validation fails, found %d entries", len(entries))
	}
}

// TestValidationError_Format tests error message formatting
func TestValidationError_Format(t *testing.T) {
	err := &ValidationError{
		MissingVars: map[string][]string{
			"name": {"hello.tmpl", "info.tmpl"},
			"org":  {"info.tmpl"},
		},
	}

	errMsg := err.Error()

	// Check format components
	if !strings.Contains(errMsg, "missing required template variables") {
		t.Error("error should contain header")
	}
	if !strings.Contains(errMsg, "name") {
		t.Error("error should mention 'name'")
	}
	if !strings.Contains(errMsg, "org") {
		t.Error("error should mention 'org'")
	}
	if !strings.Contains(errMsg, "hello.tmpl") {
		t.Error("error should mention template file")
	}
	if !strings.Contains(errMsg, "Provide missing variables") {
		t.Error("error should contain usage hint")
	}
}

// TestValidateTemplateVars_PartialProvision tests some vars provided
func TestValidateTemplateVars_PartialProvision(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "info.tmpl", "{{.name}} from {{.org}}")

	stamper := New(map[string]string{"name": "alice"}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	if err == nil {
		t.Fatal("validateTemplateVars() should fail when not all variables are provided")
	}

	errMsg := err.Error()
	// Should only complain about missing 'org', not 'name'
	if strings.Contains(errMsg, "name") {
		t.Errorf("error should not mention provided variable 'name', got: %v", err)
	}
	if !strings.Contains(errMsg, "org") {
		t.Errorf("error should mention missing variable 'org', got: %v", err)
	}
}

// TestValidateTemplateVars_InvalidTemplateIgnored tests invalid templates are skipped
func TestValidateTemplateVars_InvalidTemplateIgnored(t *testing.T) {
	src := t.TempDir()
	createTestFile(t, src, "invalid.tmpl", "Invalid {{.name")
	createTestFile(t, src, "valid.tmpl", "Valid {{.org}}")

	stamper := New(map[string]string{}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	// Should not fail due to invalid template, should check valid ones
	if err == nil {
		t.Fatal("validateTemplateVars() should fail due to missing 'org'")
	}

	errMsg := err.Error()
	// Should only complain about missing org from valid.tmpl
	if !strings.Contains(errMsg, "org") {
		t.Errorf("error should mention missing variable 'org', got: %v", err)
	}
}

// Test helpers

func assertVarsEqual(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d variables %v, want %d variables %v",
			len(got), got, len(want), want)
	}

	// Create maps for comparison (order doesn't matter for this check)
	gotMap := make(map[string]bool)
	for _, v := range got {
		gotMap[v] = true
	}

	for _, v := range want {
		if !gotMap[v] {
			t.Errorf("missing expected variable: %s", v)
		}
	}
}

// TestValidateTemplateVars_SkipsTmplNoopFiles tests .tmpl.noop files are ignored
func TestValidateTemplateVars_SkipsTmplNoopFiles(t *testing.T) {
	src := t.TempDir()

	// Create .tmpl.noop file with undefined variables
	createTestFile(t, src, "example.tmpl.noop", "{{.undefined}}")

	// Should pass validation without providing variables
	stamper := New(map[string]string{}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	if err != nil {
		t.Errorf("validateTemplateVars() should skip .tmpl.noop files, got error: %v", err)
	}
}

// TestValidateTemplateVars_TmplAndTmplNoop tests both file types
func TestValidateTemplateVars_TmplAndTmplNoop(t *testing.T) {
	src := t.TempDir()

	createTestFile(t, src, "active.tmpl", "{{.name}}")
	createTestFile(t, src, "inactive.tmpl.noop", "{{.undefined}}")

	// Should fail only for .tmpl file
	stamper := New(map[string]string{}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	if err == nil {
		t.Fatal("validateTemplateVars() should fail for .tmpl file")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "name") {
		t.Errorf("error should mention missing 'name' variable, got: %v", err)
	}
	if strings.Contains(errMsg, "undefined") {
		t.Errorf("error should NOT mention 'undefined' from .tmpl.noop file, got: %v", err)
	}
}

// TestValidateTemplateVars_OnlyTmplNoop tests directory with only .tmpl.noop
func TestValidateTemplateVars_OnlyTmplNoop(t *testing.T) {
	src := t.TempDir()

	createTestFile(t, src, "file1.tmpl.noop", "{{.var1}}")
	createTestFile(t, src, "file2.tmpl.noop", "{{.var2}}")

	// Should pass - no .tmpl files to validate
	stamper := New(map[string]string{}, ".tmpl")
	err := stamper.validateTemplateVars(src)

	if err != nil {
		t.Errorf("validateTemplateVars() should pass with only .tmpl.noop files, got: %v", err)
	}
}
