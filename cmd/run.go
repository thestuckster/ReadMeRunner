package cmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs commands",
	Long:  `Runs the commands defined in your readme file in the order they appear.`,
	Run: func(cmd *cobra.Command, args []string) {
		execute(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("path", "p", "", "Full path to the project directory containing the README file")
	runCmd.Flags().BoolP("trust", "t", false, "Auto-trust all blocks and skip confirmation prompts")
	runCmd.Flags().StringP("env", "e", "", "Path to .env file (if not provided, looks for .env in project directory)")
}

// RRBlock represents a parsed ReadMe Runner block
type RRBlock struct {
	Name      string
	Variables map[string]string
	Commands  []string
}

func execute(cmd *cobra.Command, args []string) {
	var workDir string
	var err error

	//use provided path if set.
	projectPath, _ := cmd.Flags().GetString("path")
	if projectPath != "" {
		workDir, err = filepath.Abs(projectPath)
		if err != nil {
			fmt.Printf("Error resolving project path: %v\n", err)
			os.Exit(-1)
		}

		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			fmt.Printf("Project path does not exist: %s\n", workDir)
			os.Exit(-1)
		}
	} else {
		workDir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	readmePath, exists := findReadme(workDir)
	if !exists {
		fmt.Printf("No readme found in directory %s\n", workDir)
		os.Exit(-1)
	}

	content, err := os.ReadFile(readmePath)
	if err != nil {
		fmt.Printf("Error reading readme file: %v\n", err)
		os.Exit(-1)
	}

	blocks := parseRRBlocks(string(content))
	if len(blocks) == 0 {
		fmt.Println("No RR blocks found in readme file")
		return
	}

	// Load environment variables from .env file
	envVars := loadEnvVars(cmd, workDir)

	trust, _ := cmd.Flags().GetBool("trust")

	var approvedHashes map[string]bool
	if !trust {
		approvedHashes = loadApprovedHashes(workDir)
	}

	for i, block := range blocks {
		// If trust flag is set, skip all hash operations and execute directly
		if trust {
			if err := executeBlock(block, envVars); err != nil {
				fmt.Printf("Error executing block %s: %v\n", block.Name, err)
				os.Exit(-1)
			}
			continue
		}

		// Check hash and prompt if not approved
		blockHash := hashBlock(block)
		isApproved := approvedHashes[blockHash]

		if !isApproved {
			if !promptForBlock(block, i+1, len(blocks)) {
				fmt.Println("Skipping block...")
				continue
			}
			saveBlockHash(workDir, blockHash)
		}

		if err := executeBlock(block, envVars); err != nil {
			fmt.Printf("Error executing block %s: %v\n", block.Name, err)
			os.Exit(-1)
		}
	}
}

func findReadme(workDir string) (string, bool) {
	readmePaths := []string{
		filepath.Join(workDir, "readme.md"),
		filepath.Join(workDir, "README.md"),
	}

	for _, path := range readmePaths {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	return "", false
}

// findEnvFile finds the .env file in the specified directory or returns the provided path
func findEnvFile(envPath string, workDir string) (string, bool) {
	// If env path is provided, use it directly
	if envPath != "" {
		absPath, err := filepath.Abs(envPath)
		if err != nil {
			return "", false
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, true
		}
		return "", false
	}

	envFilePath := filepath.Join(workDir, ".env")
	if _, err := os.Stat(envFilePath); err == nil {
		return envFilePath, true
	}

	return "", false
}

// loadEnvVars loads environment variables from .env file
func loadEnvVars(cmd *cobra.Command, workDir string) map[string]string {
	envVars := make(map[string]string)

	envPath, _ := cmd.Flags().GetString("env")

	envFilePath, exists := findEnvFile(envPath, workDir)
	if !exists {
		return envVars // Return empty map if no .env file found
	}

	content, err := os.ReadFile(envFilePath)
	if err != nil {
		// If we can't read the env file, return an empty map. 
		return envVars
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes
			if len(value) >= 2 {
				if (strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`)) ||
					(strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`)) {
					value = value[1 : len(value)-1]
				}
			}

			envVars[key] = value
		}
	}

	return envVars
}

// hashBlock creates a SHA256 hash of the block content
// The hash includes block name, commands, and variables to uniquely identify the block
func hashBlock(block RRBlock) string {
	var content strings.Builder
	content.WriteString("name:" + block.Name + "\n")

	var varKeys []string
	for k := range block.Variables {
		varKeys = append(varKeys, k)
	}

	for i := 0; i < len(varKeys)-1; i++ {
		for j := i + 1; j < len(varKeys); j++ {
			if varKeys[i] > varKeys[j] {
				varKeys[i], varKeys[j] = varKeys[j], varKeys[i]
			}
		}
	}
	for _, k := range varKeys {
		content.WriteString("var:" + k + "=" + block.Variables[k] + "\n")
	}

	for _, cmd := range block.Commands {
		content.WriteString("cmd:" + cmd + "\n")
	}

	hash := sha256.Sum256([]byte(content.String()))
	return hex.EncodeToString(hash[:])
}

// loadApprovedHashes reads the .rr file and returns a map of approved block hashes
func loadApprovedHashes(workDir string) map[string]bool {
	rrFilePath := filepath.Join(workDir, ".rr")
	approvedHashes := make(map[string]bool)

	if _, err := os.Stat(rrFilePath); os.IsNotExist(err) {
		return approvedHashes
	}

	content, err := os.ReadFile(rrFilePath)
	if err != nil {
		//don't crash if we can't read the file.
		return approvedHashes
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			approvedHashes[line] = true
		}
	}

	return approvedHashes
}

// saveBlockHash appends a block hash to the .rr file
func saveBlockHash(workDir string, hash string) {
	rrFilePath := filepath.Join(workDir, ".rr")

	approvedHashes := loadApprovedHashes(workDir)
	if approvedHashes[hash] {
		return
	}

	file, err := os.OpenFile(rrFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		//don't crash if we can't write to the file.
		return
	}
	defer file.Close()

	if _, err := file.WriteString(hash + "\n"); err != nil {
		return
	}
}

// parseRRBlocks extracts all RR blocks from the readme content
func parseRRBlocks(content string) []RRBlock {
	var blocks []RRBlock

	// Regex to match HTML comments that start with RR
	// Matches: <!-- RR --> or <!-- RR[BlockName] -->
	rrBlockRegex := regexp.MustCompile(`<!--\s*RR(\[([^\]]+)\])?\s*`)

	lines := strings.Split(content, "\n")
	inBlock := false
	var currentBlock *RRBlock
	var blockLines []string

	for _, line := range lines {
		// Check if this line starts an RR block
		if rrBlockRegex.MatchString(line) {
			if inBlock {
				// Close previous block if we encounter a new one
				if currentBlock != nil {
					processBlockContent(currentBlock, blockLines)
					blocks = append(blocks, *currentBlock)
				}
			}

			// Extract block name
			matches := rrBlockRegex.FindStringSubmatch(line)
			blockName := ""
			if len(matches) > 2 && matches[2] != "" {
				blockName = matches[2]
			}

			currentBlock = &RRBlock{
				Name:      blockName,
				Variables: make(map[string]string),
				Commands:  []string{},
			}
			blockLines = []string{}
			inBlock = true
			continue
		}

		// Check if this line ends the block
		if inBlock && strings.Contains(line, "-->") {
			// Remove the closing --> from the last line
			lastLine := strings.TrimSuffix(strings.TrimSpace(line), "-->")
			if lastLine != "" {
				blockLines = append(blockLines, lastLine)
			}

			processBlockContent(currentBlock, blockLines)
			blocks = append(blocks, *currentBlock)
			inBlock = false
			currentBlock = nil
			blockLines = nil
			continue
		}

		if inBlock {
			blockLines = append(blockLines, line)
		}
		// If not inBlock, the line is outside any RR block and is ignored
	}

	// Handle case where block doesn't close properly
	if inBlock && currentBlock != nil {
		processBlockContent(currentBlock, blockLines)
		blocks = append(blocks, *currentBlock)
	}

	return blocks
}

// processBlockContent processes the content of an RR block to extract variables, prompts, and commands
func processBlockContent(block *RRBlock, lines []string) {
	var currentCommand strings.Builder
	var commands []string

	varAssignRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_-]+)\s*=\s*"([^"]+)"\s*$`)
	promptRegex := regexp.MustCompile(`^\s*([a-zA-Z0-9_-]+)\s*=\s*#prompt\("([^"]+)"\)\s*$`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for variable assignment
		if matches := varAssignRegex.FindStringSubmatch(line); matches != nil {
			// Save any pending command before processing variable
			if currentCommand.Len() > 0 {
				cmd := strings.TrimSpace(currentCommand.String())
				if cmd != "" {
					commands = append(commands, cmd)
				}
				currentCommand.Reset()
			}
			block.Variables[matches[1]] = matches[2]
			continue
		}

		// Check for prompt assignment
		if matches := promptRegex.FindStringSubmatch(line); matches != nil {
			// Save any pending command before processing prompt
			if currentCommand.Len() > 0 {
				cmd := strings.TrimSpace(currentCommand.String())
				if cmd != "" {
					commands = append(commands, cmd)
				}
				currentCommand.Reset()
			}
			// Prompt will be handled during execution
			block.Variables[matches[1]] = "#PROMPT:" + matches[2]
			continue
		}

		// This is a command line
		if currentCommand.Len() > 0 {
			currentCommand.WriteString(" ")
		}

		// Handle multi-line commands (backslash continuation)
		trimmedLine := strings.TrimRight(line, " \t")
		if strings.HasSuffix(trimmedLine, "\\") {
			currentCommand.WriteString(strings.TrimSuffix(trimmedLine, "\\"))
			continue //next line
		}

		// Add the line to current command
		currentCommand.WriteString(line)

		// If no backslash, this command is complete
		cmd := strings.TrimSpace(currentCommand.String())
		if cmd != "" {
			commands = append(commands, cmd)
		}
		currentCommand.Reset()
	}

	// Add any remaining command
	if currentCommand.Len() > 0 {
		cmd := strings.TrimSpace(currentCommand.String())
		if cmd != "" {
			commands = append(commands, cmd)
		}
	}

	block.Commands = commands
}

// promptForBlock prompts the user for confirmation before executing a block
// Returns true if user confirms with "y", false otherwise
func promptForBlock(block RRBlock, blockNum, totalBlocks int) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n--- Block %d of %d ---\n", blockNum, totalBlocks)
	if block.Name != "" {
		fmt.Printf("Block Name: %s\n", block.Name)
	} else {
		// Show first command as identifier if no name
		if len(block.Commands) > 0 {
			firstCmd := block.Commands[0]
			if len(firstCmd) > 50 {
				firstCmd = firstCmd[:50] + "..."
			}
			fmt.Printf("Command: %s\n", firstCmd)
		}
	}

	// Show commands that will be executed
	if len(block.Commands) > 0 {
		fmt.Println("Commands to execute:")
		for i, cmd := range block.Commands {
			fmt.Printf("  %d. %s\n", i+1, cmd)
		}
	}

	// Show variables if any
	if len(block.Variables) > 0 {
		fmt.Println("Variables:")
		for varName, varValue := range block.Variables {
			if strings.HasPrefix(varValue, "#PROMPT:") {
				fmt.Printf("  %s = #prompt(\"%s\")\n", varName, strings.TrimPrefix(varValue, "#PROMPT:"))
			} else {
				fmt.Printf("  %s = \"%s\"\n", varName, varValue)
			}
		}
	}

	// Prompt for confirmation
	fmt.Print("\nExecute this block? (y/n): ")
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("\nError reading input: %v\n", err)
		return false
	}

	fmt.Println() //ensure next prompt appears on new line

	response := strings.TrimSpace(strings.ToLower(input))
	return response == "y" || response == "yes"
}

// executeBlock executes a single RR block
func executeBlock(block RRBlock, envVars map[string]string) error {
	// First, handle prompts and populate variables
	reader := bufio.NewReader(os.Stdin)
	for varName, varValue := range block.Variables {
		if strings.HasPrefix(varValue, "#PROMPT:") {
			question := strings.TrimPrefix(varValue, "#PROMPT:")
			// Ensure prompt appears on a new line
			fmt.Printf("\n%s ", question)
			input, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("error reading input for prompt: %v", err)
			}
			block.Variables[varName] = strings.TrimSpace(input)
		}
	}

	// Execute each command
	for _, cmd := range block.Commands {
		// Replace variable references in command
		// Block variables take precedence over env variables
		mergedVars := make(map[string]string)
		// First add env variables
		for k, v := range envVars {
			mergedVars[k] = v
		}
		// Then add block variables (they override env variables)
		for k, v := range block.Variables {
			mergedVars[k] = v
		}
		// Substitute variables (block vars override env vars)
		cmd = substituteVariables(cmd, mergedVars)

		// Display block name or command for confirmation
		if block.Name != "" {
			fmt.Printf("\n[%s]\nExecuting: %s\nOutput:\n", block.Name, cmd)
		} else {
			fmt.Printf("\nExecuting: %s\nOutput:\n", cmd)
		}

		// Execute the command
		shellCmd := exec.Command("sh", "-c", cmd)
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr
		shellCmd.Stdin = os.Stdin

		if err := shellCmd.Run(); err != nil {
			return fmt.Errorf("command failed: %v", err)
		}
	}

	return nil
}

// substituteVariables replaces variable references (#var-name) with their values
func substituteVariables(cmd string, variables map[string]string) string {
	varUsageRegex := regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)

	return varUsageRegex.ReplaceAllStringFunc(cmd, func(match string) string {
		varName := strings.TrimPrefix(match, "#")
		if value, exists := variables[varName]; exists {
			return value
		}
		return match
	})
}
