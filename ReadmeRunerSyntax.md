ReadMe Runner Syntax
---

A ReadMe Runner block (henceforth defined as an RR Block) is defined inline in your ReadMe file using HTML style comments.

# Basic Structure 

Every RR Block **MUST** begin with `RR` at the top of the comment. 
This identifies which comments contain executable code versus regular comments.

**Example:**
```
<!-- RR -->
```

# Naming 

While optional, naming your RR blocks is recommended. Named blocks display their name when prompting for confirmation, while unnamed blocks display the raw command instead.

**Syntax:** `RR[BlockName]`

**Example:**
```
<!-- RR[Echo] --> 
```

# Adding Commands

Commands are added on new lines within your RR block. Any syntactically correct command, tool, or script can be executed.

**Example:**
```
<!-- RR[Echo]
echo "Hello Readme Runner!"
-->
```

## Multi line commands

Multi-line commands use standard bash syntax with a trailing backslash (`\`) to continue to the next line.

**Example:**
```
<!-- RR[Multi-line]
echo "Starting process..." && \
sleep 2 && \
echo "Process complete!"
-->
```

# Variables

You can add variables to your RR command block with the following syntax.

**Syntax:** `my-var = "test"`

Variables can be used in your commands by using a `#` followed by your variable name. `#my-var`

**Example:**
```
<!-- RR[Variables]
    my-var = "test"
    echo #my-var
-->
```

**NOTE** as of right now, variables are scoped to only their respective RR Block. To share variables between blocks
you will have to take advantage of the .env file support mentioned below.

# Environment Variables

You can use environment variables from `.env` files in your RR blocks. This is useful for configuration values that you don't want to hardcode in your README file.

## Creating a .env File

Create a `.env` file in your project directory (or specify a custom path with the `--env` flag). The file uses standard `KEY=VALUE` format:

**Example `.env` file:**
```
APP_NAME=MyApplication
API_KEY=secret-key-12345
DATABASE_URL=postgresql://localhost:5432/mydb
DEBUG=true
ENVIRONMENT=development
```

**Note:** 
- Empty lines and lines starting with `#` are ignored (comments)
- Values can be quoted with single or double quotes (quotes are automatically removed)
- If no `--env` flag is provided, RR automatically looks for `.env` in the project directory

## Using Environment Variables

Environment variables can be referenced in your commands using the same `#VARIABLE_NAME` syntax as block variables.

**Example:**
```
<!-- RR[Config Test]
    echo "Application: #APP_NAME"
    echo "API Key: #API_KEY"
    echo "Database: #DATABASE_URL"
    echo "Debug mode: #DEBUG"
    echo "Environment: #ENVIRONMENT"
-->
```

## Variable Precedence

When a variable name exists in both a block variable and an environment variable, **block variables take precedence**. This means:

1. Block variables (including prompts) override environment variables
2. Environment variables are used if the variable doesn't exist in the block

**Example:**
```
<!-- RR[Variable Precedence]
    # APP_NAME exists in both .env and block
    APP_NAME = "Block Value"  # This will be used
    echo "App name: #APP_NAME"  # Outputs: "App name: Block Value"
    
    # API_KEY only exists in .env
    echo "API Key: #API_KEY"  # Uses value from .env file
-->
```

## Specifying a Custom .env File

You can specify a custom path to a `.env` file using the `--env` flag:

```bash
rr run --env /path/to/custom/.env
# or
rr run -e /path/to/custom/.env
```

# Prompting for Input

You can prompt the user for input in your RR blocks. When using the Prompt syntax, RR will prompt and wait for the user
to give inputs for each prompt before continuing on with script execution. This can be helpful to prevent committing things
such as passwords and other important secrets.

Prompts are assigned to a variable name so that RR can easily replace them in your command. Prompts require a question 
to be asked to the user as a part of the syntax

**Syntax:** `my-var = #prompt("")`

**Example:**
```
<!-- RR[Prompting]
my-name = #prompt("What is your name?")
echo "Hello #my-name"
-->
```