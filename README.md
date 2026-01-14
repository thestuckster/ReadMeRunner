# ReadMe Runner (RR)

**ReadMe Runner** is a command-line tool that automatically executes instructions embedded in your README files. Define executable code blocks directly in your documentation using simple HTML comments, and let RR run them safely with built-in confirmation prompts and approval tracking.

## Features

- 🚀 **Execute commands from README files** - Embed executable code blocks directly in your documentation
- 🔒 **Safety first** - Confirmation prompts for each block before execution
- ✅ **Approval tracking** - Automatically remembers approved blocks using hash-based tracking
- 🔑 **Variable support** - Use variables and prompts for dynamic command execution
- 🌍 **Environment variables** - Load variables from `.env` files for configuration
- 📝 **Multi-line commands** - Support for complex multi-line bash commands
- 🎯 **Selective execution** - Only executes code within designated RR blocks, ignores everything else

## Installation

### From Source

```bash
git clone <repository-url>
cd readmerunner
go build -o rr
```

Move the `rr` binary to a location in your PATH, or use it directly from the project directory.

## Quick Start

1. **Create an RR block in your README file:**

```markdown
<!-- RR[Hello World]
echo "Hello from ReadMe Runner!"
-->
```

2. **Run the commands:**

```bash
rr run
```

3. **Confirm execution when prompted:**

```
--- Block 1 of 1 ---
Block Name: Hello World
Commands to execute:
  1. echo "Hello from ReadMe Runner!"

Execute this block? (y/n): y
```

## Usage

### Basic Command

```bash
rr run
```

Executes all RR blocks found in `README.md` or `readme.md` in the current directory.

### Command-Line Options

#### `--path` / `-p`

Specify a custom project directory:

```bash
rr run --path /path/to/project
# or
rr run -p /path/to/project
```

#### `--trust` / `-t`

Auto-execute all blocks without prompts (skips hash checking):

```bash
rr run --trust
# or
rr run -t
```

**Note:** When using `--trust`, blocks are executed immediately without confirmation prompts or hash tracking.

#### `--env` / `-e`

Specify a custom path to a `.env` file:

```bash
rr run --env /path/to/.env
# or
rr run -e /path/to/.env
```

**Note:** If `--env` is not provided, RR will automatically look for a `.env` file in the project directory (specified by `--path` or current directory).

## How It Works

### RR Blocks

RR blocks are defined using HTML-style comments in your README file. Only content within `<!-- RR ... -->` blocks is executed. Everything else is completely ignored.

### Environment Variables

ReadMe Runner supports loading environment variables from `.env` files:

- **Automatic discovery**: If no `--env` flag is provided, RR looks for `.env` in the project directory
- **Custom path**: Use `--env` to specify a custom `.env` file location
- **Standard format**: Supports standard `.env` file format (`KEY=VALUE`)
- **Variable precedence**: Block variables override environment variables if names conflict
- **Quoted values**: Supports both single and double-quoted values in `.env` files

Environment variables are loaded before block execution and can be used in commands using the `#VARIABLE_NAME` syntax.

### Block Approval Tracking

ReadMe Runner uses a `.rr` file in your project directory to track approved blocks:

- **First run**: You'll be prompted to approve each block
- **Approved blocks**: Once approved, blocks are hashed and saved to `.rr`
- **Subsequent runs**: Previously approved blocks (with matching content) execute automatically
- **Modified blocks**: If a block's content changes, its hash changes and you'll be prompted again

This ensures you're always aware of what's being executed while avoiding repetitive confirmations for trusted blocks.

## Examples

### Basic Block

```markdown
<!-- RR
echo "This is a basic unnamed RR block"
-->
```

### Named Block

```markdown
<!-- RR[Setup Environment]
export NODE_ENV=production
npm install
-->
```

### Using Variables

```markdown
<!-- RR[Deploy]
    environment = "staging"
    echo "Deploying to #environment environment"
    deploy.sh --env #environment
-->
```

### User Prompts

```markdown
<!-- RR[Login]
    username = #prompt("Enter your username:")
    password = #prompt("Enter your password:")
    login.sh -u #username -p #password
-->
```

### Multi-line Commands

```markdown
<!-- RR[Build and Deploy]
    echo "Building application..." && \
    npm run build && \
    echo "Deploying..." && \
    deploy.sh
-->
```

### Combined Features

```markdown
<!-- RR[Full Setup]
    project-name = #prompt("What is your project name?")
    env = "development"
    echo "Setting up #project-name in #env mode" && \
    npm install && \
    npm run setup
-->
```

### Using Environment Variables

Environment variables from `.env` files can be used in your commands:

**`.env` file:**
```
APP_NAME=MyApp
API_KEY=secret-key-123
DEBUG=true
```

**RR Block:**
```markdown
<!-- RR[Config Test]
    echo "App: #APP_NAME"
    echo "API Key: #API_KEY"
    echo "Debug mode: #DEBUG"
-->
```

Variables from `.env` files are automatically loaded and can be referenced using `#VARIABLE_NAME` syntax. Block variables take precedence over environment variables if there's a name conflict.

## Syntax Reference

For complete syntax documentation, including all available features and detailed examples, see [ReadmeRunerSyntax.md](./ReadmeRunerSyntax.md).

## Safety Features

### Command Isolation

- **Only RR blocks execute**: Commands outside HTML comment blocks are completely ignored
- **No accidental execution**: Regular markdown code blocks and inline commands are safe

### Confirmation System

- **Interactive prompts**: Each block requires explicit approval before execution
- **Block information**: See exactly what commands will run before confirming
- **Skip option**: Choose to skip any block you're unsure about

### Hash-Based Tracking

- **Content verification**: Blocks are hashed based on their content (name, commands, variables)
- **Automatic approval**: Previously approved blocks run without prompts
- **Change detection**: Modified blocks require re-approval

## File Structure

```
your-project/
├── README.md          # Your README with RR blocks
├── .env               # Environment variables (optional)
├── .rr                # Approval tracking file (auto-generated)
└── ...
```

- **`.env`**: Optional file containing environment variables in `KEY=VALUE` format. Automatically loaded if present in the project directory.
- **`.rr`**: Automatically created in your project directory when you first approve a block. It contains SHA256 hashes of approved blocks.

## Best Practices

1. **Name your blocks**: Named blocks are easier to identify in confirmation prompts
2. **Use variables for secrets**: Use `#prompt()` for sensitive information like passwords, or store them in `.env` files (and add `.env` to `.gitignore`)
3. **Keep blocks focused**: Each block should have a single, clear purpose
4. **Review before approving**: Always review the commands before confirming execution
5. **Version control `.rr`**: Consider adding `.rr` to your `.gitignore` if you want per-developer approval tracking
6. **Environment variables**: Use `.env` files for configuration that varies between environments (development, staging, production)

## Troubleshooting

### No blocks found

If you see "No RR blocks found in readme file", ensure:
- Your README file is named `README.md` or `readme.md`
- RR blocks use the correct syntax: `<!-- RR ... -->`
- Blocks are properly closed with `-->`

### Blocks not executing

- Check that you've confirmed the block with `y` when prompted
- Verify the block syntax is correct
- Ensure commands are inside the HTML comment block

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](./LICENSE) file for details.
