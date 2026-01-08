package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseRRBlocks_BasicBlock(t *testing.T) {
	content := `<!-- RR
echo "Hello World"
-->`

	blocks := parseRRBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	block := blocks[0]
	if block.Name != "" {
		t.Errorf("Expected empty name, got %s", block.Name)
	}
	if len(block.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(block.Commands))
	}
	if block.Commands[0] != `echo "Hello World"` {
		t.Errorf("Expected command 'echo \"Hello World\"', got '%s'", block.Commands[0])
	}
}

func TestParseRRBlocks_NamedBlock(t *testing.T) {
	content := `<!-- RR[Test Block]
echo "Test"
-->`

	blocks := parseRRBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	block := blocks[0]
	if block.Name != "Test Block" {
		t.Errorf("Expected name 'Test Block', got '%s'", block.Name)
	}
}

func TestParseRRBlocks_MultipleBlocks(t *testing.T) {
	content := `<!-- RR[First]
echo "First"
-->
Some text in between
<!-- RR[Second]
echo "Second"
-->`

	blocks := parseRRBlocks(content)
	if len(blocks) != 2 {
		t.Fatalf("Expected 2 blocks, got %d", len(blocks))
	}

	if blocks[0].Name != "First" {
		t.Errorf("Expected first block name 'First', got '%s'", blocks[0].Name)
	}
	if blocks[1].Name != "Second" {
		t.Errorf("Expected second block name 'Second', got '%s'", blocks[1].Name)
	}
}

func TestParseRRBlocks_IgnoresOutsideBlocks(t *testing.T) {
	content := `echo "This should be ignored"
<!-- RR[Test]
echo "This should be parsed"
-->
echo "This should also be ignored"`

	blocks := parseRRBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if len(blocks[0].Commands) != 1 {
		t.Fatalf("Expected 1 command in block, got %d", len(blocks[0].Commands))
	}
	if !strings.Contains(blocks[0].Commands[0], "This should be parsed") {
		t.Errorf("Block should contain 'This should be parsed', got '%s'", blocks[0].Commands[0])
	}
}

func TestProcessBlockContent_Variables(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`my-var = "test value"`,
		`echo #my-var`,
	}

	processBlockContent(block, lines)

	if block.Variables["my-var"] != "test value" {
		t.Errorf("Expected variable 'my-var' to be 'test value', got '%s'", block.Variables["my-var"])
	}
	if len(block.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(block.Commands))
	}
	if block.Commands[0] != "echo #my-var" {
		t.Errorf("Expected command 'echo #my-var', got '%s'", block.Commands[0])
	}
}

func TestProcessBlockContent_MultipleVariables(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`var1 = "value1"`,
		`var2 = "value2"`,
		`echo #var1 #var2`,
	}

	processBlockContent(block, lines)

	if block.Variables["var1"] != "value1" {
		t.Errorf("Expected var1 to be 'value1', got '%s'", block.Variables["var1"])
	}
	if block.Variables["var2"] != "value2" {
		t.Errorf("Expected var2 to be 'value2', got '%s'", block.Variables["var2"])
	}
}

func TestProcessBlockContent_Prompts(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`my-name = #prompt("What is your name?")`,
		`echo "Hello #my-name"`,
	}

	processBlockContent(block, lines)

	expectedPrompt := "#PROMPT:What is your name?"
	if block.Variables["my-name"] != expectedPrompt {
		t.Errorf("Expected prompt variable to be '%s', got '%s'", expectedPrompt, block.Variables["my-name"])
	}
}

func TestProcessBlockContent_MultiLineCommands(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`echo "First" && \`,
		`echo "Second" && \`,
		`echo "Third"`,
	}

	processBlockContent(block, lines)

	if len(block.Commands) != 1 {
		t.Fatalf("Expected 1 command, got %d", len(block.Commands))
	}

	expected := `echo "First" &&  echo "Second" &&  echo "Third"`
	if block.Commands[0] != expected {
		t.Errorf("Expected multi-line command '%s', got '%s'", expected, block.Commands[0])
	}
}

func TestProcessBlockContent_MultipleCommands(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`echo "First command"`,
		`echo "Second command"`,
		`echo "Third command"`,
	}

	processBlockContent(block, lines)

	if len(block.Commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(block.Commands))
	}
}

func TestProcessBlockContent_VariablesAndCommands(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`my-var = "test"`,
		`echo #my-var`,
		`another-var = "value"`,
		`echo #another-var`,
	}

	processBlockContent(block, lines)

	if block.Variables["my-var"] != "test" {
		t.Errorf("Expected my-var to be 'test'")
	}
	if block.Variables["another-var"] != "value" {
		t.Errorf("Expected another-var to be 'value'")
	}
	if len(block.Commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(block.Commands))
	}
}

func TestHashBlock_Consistency(t *testing.T) {
	block1 := RRBlock{
		Name: "Test",
		Variables: map[string]string{
			"var1": "value1",
			"var2": "value2",
		},
		Commands: []string{"echo test"},
	}

	block2 := RRBlock{
		Name: "Test",
		Variables: map[string]string{
			"var1": "value1",
			"var2": "value2",
		},
		Commands: []string{"echo test"},
	}

	hash1 := hashBlock(block1)
	hash2 := hashBlock(block2)

	if hash1 != hash2 {
		t.Errorf("Expected same hash for identical blocks, got %s and %s", hash1, hash2)
	}
}

func TestHashBlock_DifferentContent(t *testing.T) {
	block1 := RRBlock{
		Name:     "Test",
		Variables: map[string]string{"var": "value1"},
		Commands: []string{"echo test"},
	}

	block2 := RRBlock{
		Name:     "Test",
		Variables: map[string]string{"var": "value2"},
		Commands: []string{"echo test"},
	}

	hash1 := hashBlock(block1)
	hash2 := hashBlock(block2)

	if hash1 == hash2 {
		t.Error("Expected different hashes for different content")
	}
}

func TestHashBlock_IncludesName(t *testing.T) {
	block1 := RRBlock{
		Name:     "Block1",
		Variables: map[string]string{},
		Commands: []string{"echo test"},
	}

	block2 := RRBlock{
		Name:     "Block2",
		Variables: map[string]string{},
		Commands: []string{"echo test"},
	}

	hash1 := hashBlock(block1)
	hash2 := hashBlock(block2)

	if hash1 == hash2 {
		t.Error("Expected different hashes for blocks with different names")
	}
}

func TestSubstituteVariables(t *testing.T) {
	cmd := "echo #my-var and #another-var"
	variables := map[string]string{
		"my-var":     "value1",
		"another-var": "value2",
	}

	result := substituteVariables(cmd, variables)
	expected := "echo value1 and value2"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestSubstituteVariables_UnknownVariable(t *testing.T) {
	cmd := "echo #unknown-var"
	variables := map[string]string{}

	result := substituteVariables(cmd, variables)
	expected := "echo #unknown-var"

	if result != expected {
		t.Errorf("Expected unknown variable to remain unchanged, got '%s'", result)
	}
}

func TestSubstituteVariables_MultipleOccurrences(t *testing.T) {
	cmd := "echo #var and #var again"
	variables := map[string]string{
		"var": "test",
	}

	result := substituteVariables(cmd, variables)
	expected := "echo test and test again"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestLoadApprovedHashes_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()
	hashes := loadApprovedHashes(tempDir)

	if len(hashes) != 0 {
		t.Errorf("Expected empty map for non-existent file, got %d entries", len(hashes))
	}
}

func TestLoadApprovedHashes_ExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	rrFile := filepath.Join(tempDir, ".rr")
	
	content := "hash1\nhash2\nhash3\n"
	err := os.WriteFile(rrFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create .rr file: %v", err)
	}

	hashes := loadApprovedHashes(tempDir)

	if len(hashes) != 3 {
		t.Fatalf("Expected 3 hashes, got %d", len(hashes))
	}

	if !hashes["hash1"] {
		t.Error("Expected hash1 to be in approved hashes")
	}
	if !hashes["hash2"] {
		t.Error("Expected hash2 to be in approved hashes")
	}
	if !hashes["hash3"] {
		t.Error("Expected hash3 to be in approved hashes")
	}
}

func TestLoadApprovedHashes_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	rrFile := filepath.Join(tempDir, ".rr")
	
	err := os.WriteFile(rrFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create .rr file: %v", err)
	}

	hashes := loadApprovedHashes(tempDir)

	if len(hashes) != 0 {
		t.Errorf("Expected empty map for empty file, got %d entries", len(hashes))
	}
}

func TestSaveBlockHash_NewHash(t *testing.T) {
	tempDir := t.TempDir()
	hash := "test-hash-123"

	saveBlockHash(tempDir, hash)

	// Verify hash was saved
	hashes := loadApprovedHashes(tempDir)
	if !hashes[hash] {
		t.Error("Expected hash to be saved")
	}
}

func TestSaveBlockHash_DuplicateHash(t *testing.T) {
	tempDir := t.TempDir()
	hash := "test-hash-123"

	// Save hash twice
	saveBlockHash(tempDir, hash)
	saveBlockHash(tempDir, hash)

	// Read file and count occurrences
	rrFile := filepath.Join(tempDir, ".rr")
	content, err := os.ReadFile(rrFile)
	if err != nil {
		t.Fatalf("Failed to read .rr file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Errorf("Expected hash to appear once, found %d times", len(lines))
	}
}

func TestSaveBlockHash_MultipleHashes(t *testing.T) {
	tempDir := t.TempDir()
	hash1 := "hash1"
	hash2 := "hash2"
	hash3 := "hash3"

	saveBlockHash(tempDir, hash1)
	saveBlockHash(tempDir, hash2)
	saveBlockHash(tempDir, hash3)

	hashes := loadApprovedHashes(tempDir)
	if len(hashes) != 3 {
		t.Fatalf("Expected 3 hashes, got %d", len(hashes))
	}

	if !hashes[hash1] || !hashes[hash2] || !hashes[hash3] {
		t.Error("Expected all hashes to be saved")
	}
}

func TestFindReadme_READMEExists(t *testing.T) {
	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "README.md")
	
	err := os.WriteFile(readmePath, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create README: %v", err)
	}

	path, exists := findReadme(tempDir)
	if !exists {
		t.Fatal("Expected README to be found")
	}
	// Note: findReadme checks simple-example.md first, then readme.md, then README.md
	// So we just verify it exists, not the exact path
	if !strings.HasSuffix(path, ".md") {
		t.Errorf("Expected path to end with .md, got '%s'", path)
	}
}

func TestFindReadme_ReadmeExists(t *testing.T) {
	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "readme.md")
	
	err := os.WriteFile(readmePath, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create readme: %v", err)
	}

	path, exists := findReadme(tempDir)
	if !exists {
		t.Fatal("Expected readme to be found")
	}
	if path != readmePath {
		t.Errorf("Expected path '%s', got '%s'", readmePath, path)
	}
}

func TestFindReadme_SimpleExampleExists(t *testing.T) {
	tempDir := t.TempDir()
	readmePath := filepath.Join(tempDir, "simple-example.md")
	
	err := os.WriteFile(readmePath, []byte("# Test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create simple-example: %v", err)
	}

	path, exists := findReadme(tempDir)
	if !exists {
		t.Fatal("Expected simple-example to be found")
	}
	if path != readmePath {
		t.Errorf("Expected path '%s', got '%s'", readmePath, path)
	}
}

func TestFindReadme_NotFound(t *testing.T) {
	tempDir := t.TempDir()

	_, exists := findReadme(tempDir)
	if exists {
		t.Error("Expected readme not to be found")
	}
}

func TestParseRRBlocks_ComplexExample(t *testing.T) {
	content := `# Project README

Some documentation here.

<!-- RR[Setup]
    env = "development"
    echo "Setting up #env environment"
-->

More documentation.

<!-- RR[Deploy]
    project-name = #prompt("Project name?")
    echo "Deploying #project-name"
    deploy.sh #project-name
-->

<!-- RR[Multi-line]
    echo "Starting..." && \
    sleep 1 && \
    echo "Done"
-->
`

	blocks := parseRRBlocks(content)
	if len(blocks) != 3 {
		t.Fatalf("Expected 3 blocks, got %d", len(blocks))
	}

	// Check first block
	if blocks[0].Name != "Setup" {
		t.Errorf("Expected first block name 'Setup', got '%s'", blocks[0].Name)
	}
	if blocks[0].Variables["env"] != "development" {
		t.Errorf("Expected env variable to be 'development'")
	}

	// Check second block
	if blocks[1].Name != "Deploy" {
		t.Errorf("Expected second block name 'Deploy', got '%s'", blocks[1].Name)
	}
	if !strings.HasPrefix(blocks[1].Variables["project-name"], "#PROMPT:") {
		t.Error("Expected project-name to be a prompt variable")
	}

	// Check third block
	if blocks[2].Name != "Multi-line" {
		t.Errorf("Expected third block name 'Multi-line', got '%s'", blocks[2].Name)
	}
	if len(blocks[2].Commands) != 1 {
		t.Fatalf("Expected 1 multi-line command, got %d", len(blocks[2].Commands))
	}
}

func TestParseRRBlocks_UnclosedBlock(t *testing.T) {
	content := `<!-- RR[Test]
echo "Test"
`

	blocks := parseRRBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block even if unclosed, got %d", len(blocks))
	}
}

func TestProcessBlockContent_EmptyLines(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		"",
		`echo "test"`,
		"   ",
		`echo "another"`,
		"",
	}

	processBlockContent(block, lines)

	if len(block.Commands) != 2 {
		t.Fatalf("Expected 2 commands (empty lines should be ignored), got %d", len(block.Commands))
	}
}

func TestHashBlock_EmptyBlock(t *testing.T) {
	block := RRBlock{
		Name:      "",
		Variables: map[string]string{},
		Commands:  []string{},
	}

	hash := hashBlock(block)
	if hash == "" {
		t.Error("Expected non-empty hash even for empty block")
	}
}

func TestHashBlock_OrderIndependent(t *testing.T) {
	// Test that variable order doesn't affect hash (due to sorting)
	block1 := RRBlock{
		Name: "Test",
		Variables: map[string]string{
			"a": "1",
			"b": "2",
		},
		Commands: []string{"echo test"},
	}

	block2 := RRBlock{
		Name: "Test",
		Variables: map[string]string{
			"b": "2",
			"a": "1",
		},
		Commands: []string{"echo test"},
	}

	hash1 := hashBlock(block1)
	hash2 := hashBlock(block2)

	if hash1 != hash2 {
		t.Error("Expected same hash for blocks with variables in different order")
	}
}

// Integration tests for CLI flags and full execution flow

func TestExecute_WithPathFlag(t *testing.T) {
	// Test that the path flag exists and can be accessed
	// We test the flag definition rather than redefining it
	pathFlag := runCmd.Flags().Lookup("path")
	if pathFlag == nil {
		t.Fatal("Expected 'path' flag to be defined")
	}
	if pathFlag.Name != "path" {
		t.Errorf("Expected flag name 'path', got '%s'", pathFlag.Name)
	}
}

func TestExecute_WithTrustFlag(t *testing.T) {
	// Test that the trust flag exists and can be accessed
	trustFlag := runCmd.Flags().Lookup("trust")
	if trustFlag == nil {
		t.Fatal("Expected 'trust' flag to be defined")
	}
	if trustFlag.Name != "trust" {
		t.Errorf("Expected flag name 'trust', got '%s'", trustFlag.Name)
	}
}

func TestExecute_FlagsDefined(t *testing.T) {
	// Verify both flags are properly defined
	pathFlag := runCmd.Flags().Lookup("path")
	trustFlag := runCmd.Flags().Lookup("trust")

	if pathFlag == nil {
		t.Error("Expected 'path' flag to be defined")
	}
	if trustFlag == nil {
		t.Error("Expected 'trust' flag to be defined")
	}
}

func TestParseRRBlocks_RealWorldExample(t *testing.T) {
	content := `# My Project

This is a project README with embedded commands.

<!-- RR[Install Dependencies]
npm install
-->

<!-- RR[Build]
    env = "production"
    npm run build --env #env
-->

<!-- RR[Deploy]
    project = #prompt("Project name?")
    deploy.sh --project #project
-->

Regular markdown content here.

` + "```" + `bash
echo "This code block should be ignored"
` + "```" + `

<!-- RR[Cleanup]
rm -rf node_modules
-->
`

	blocks := parseRRBlocks(content)
	if len(blocks) != 4 {
		t.Fatalf("Expected 4 blocks, got %d", len(blocks))
	}

	// Verify blocks are parsed correctly
	if blocks[0].Name != "Install Dependencies" {
		t.Errorf("Expected first block name 'Install Dependencies'")
	}

	if blocks[1].Name != "Build" {
		t.Errorf("Expected second block name 'Build'")
	}
	if blocks[1].Variables["env"] != "production" {
		t.Error("Expected env variable in Build block")
	}

	if blocks[2].Name != "Deploy" {
		t.Errorf("Expected third block name 'Deploy'")
	}
	if !strings.HasPrefix(blocks[2].Variables["project"], "#PROMPT:") {
		t.Error("Expected project to be a prompt variable")
	}

	if blocks[3].Name != "Cleanup" {
		t.Errorf("Expected fourth block name 'Cleanup'")
	}
}

func TestProcessBlockContent_ComplexScenario(t *testing.T) {
	block := &RRBlock{
		Name:      "Complex",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`var1 = "value1"`,
		`var2 = #prompt("Enter value:")`,
		`echo "First: #var1"`,
		`echo "Second: #var2" && \`,
		`echo "Third line"`,
		`var3 = "value3"`,
		`echo "Fourth: #var3"`,
	}

	processBlockContent(block, lines)

	// Check variables
	if block.Variables["var1"] != "value1" {
		t.Error("Expected var1 to be set")
	}
	if !strings.HasPrefix(block.Variables["var2"], "#PROMPT:") {
		t.Error("Expected var2 to be a prompt")
	}
	if block.Variables["var3"] != "value3" {
		t.Error("Expected var3 to be set")
	}

	// Check commands
	if len(block.Commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(block.Commands))
	}

	// First command
	if !strings.Contains(block.Commands[0], "First: #var1") {
		t.Errorf("Expected first command to contain 'First: #var1', got '%s'", block.Commands[0])
	}

	// Second command (multi-line)
	if !strings.Contains(block.Commands[1], "Second: #var2") {
		t.Errorf("Expected second command to contain 'Second: #var2', got '%s'", block.Commands[1])
	}
	if !strings.Contains(block.Commands[1], "Third line") {
		t.Errorf("Expected second command to contain 'Third line', got '%s'", block.Commands[1])
	}

	// Third command
	if !strings.Contains(block.Commands[2], "Fourth: #var3") {
		t.Errorf("Expected third command to contain 'Fourth: #var3', got '%s'", block.Commands[2])
	}
}

func TestSubstituteVariables_WithPrompts(t *testing.T) {
	cmd := "echo Hello #name, your age is #age"
	variables := map[string]string{
		"name": "John",
		"age":  "30",
	}

	result := substituteVariables(cmd, variables)
	expected := "echo Hello John, your age is 30"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestSubstituteVariables_NoVariables(t *testing.T) {
	cmd := "echo Hello World"
	variables := map[string]string{}

	result := substituteVariables(cmd, variables)

	if result != cmd {
		t.Errorf("Expected command to remain unchanged, got '%s'", result)
	}
}

func TestHashBlock_WithPrompts(t *testing.T) {
	block1 := RRBlock{
		Name: "Test",
		Variables: map[string]string{
			"var": "#PROMPT:Enter value:",
		},
		Commands: []string{"echo test"},
	}

	block2 := RRBlock{
		Name: "Test",
		Variables: map[string]string{
			"var": "#PROMPT:Enter value:",
		},
		Commands: []string{"echo test"},
	}

	hash1 := hashBlock(block1)
	hash2 := hashBlock(block2)

	if hash1 != hash2 {
		t.Error("Expected same hash for blocks with same prompt variables")
	}
}

func TestLoadApprovedHashes_WithWhitespace(t *testing.T) {
	tempDir := t.TempDir()
	rrFile := filepath.Join(tempDir, ".rr")
	
	content := "hash1\n  hash2  \n\nhash3\n"
	err := os.WriteFile(rrFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create .rr file: %v", err)
	}

	hashes := loadApprovedHashes(tempDir)

	// Should have 3 hashes, whitespace should be trimmed
	if len(hashes) != 3 {
		t.Fatalf("Expected 3 hashes, got %d", len(hashes))
	}

	if !hashes["hash1"] || !hashes["hash2"] || !hashes["hash3"] {
		t.Error("Expected all hashes to be loaded (whitespace trimmed)")
	}
}

func TestSaveAndLoadHashes_Integration(t *testing.T) {
	tempDir := t.TempDir()
	hash1 := "abc123"
	hash2 := "def456"
	hash3 := "ghi789"

	// Save hashes
	saveBlockHash(tempDir, hash1)
	saveBlockHash(tempDir, hash2)
	saveBlockHash(tempDir, hash3)

	// Load and verify
	hashes := loadApprovedHashes(tempDir)
	if len(hashes) != 3 {
		t.Fatalf("Expected 3 hashes after save, got %d", len(hashes))
	}

	// Verify each hash
	if !hashes[hash1] {
		t.Error("hash1 not found after save/load")
	}
	if !hashes[hash2] {
		t.Error("hash2 not found after save/load")
	}
	if !hashes[hash3] {
		t.Error("hash3 not found after save/load")
	}
}

func TestParseRRBlocks_IgnoresCodeBlocks(t *testing.T) {
	content := `# README

` + "```" + `bash
echo "This should be ignored"
` + "```" + `

<!-- RR[Test]
echo "This should be parsed"
-->

` + "```" + `bash
rm -rf /  # This dangerous command should be ignored
` + "```" + `
`

	blocks := parseRRBlocks(content)
	if len(blocks) != 1 {
		t.Fatalf("Expected 1 block, got %d", len(blocks))
	}

	if blocks[0].Name != "Test" {
		t.Errorf("Expected block name 'Test', got '%s'", blocks[0].Name)
	}
}

func TestProcessBlockContent_CommandWithVariables(t *testing.T) {
	block := &RRBlock{
		Name:      "Test",
		Variables:  make(map[string]string),
		Commands:  []string{},
	}

	lines := []string{
		`var = "test"`,
		`echo "Value: #var"`,
		`echo "Another: #var"`,
	}

	processBlockContent(block, lines)

	if block.Variables["var"] != "test" {
		t.Error("Expected var to be set")
	}
	if len(block.Commands) != 2 {
		t.Fatalf("Expected 2 commands, got %d", len(block.Commands))
	}
}

func TestHashBlock_CommandOrderMatters(t *testing.T) {
	block1 := RRBlock{
		Name:     "Test",
		Variables: map[string]string{},
		Commands: []string{"echo first", "echo second"},
	}

	block2 := RRBlock{
		Name:     "Test",
		Variables: map[string]string{},
		Commands: []string{"echo second", "echo first"},
	}

	hash1 := hashBlock(block1)
	hash2 := hashBlock(block2)

	if hash1 == hash2 {
		t.Error("Expected different hashes for blocks with different command order")
	}
}

