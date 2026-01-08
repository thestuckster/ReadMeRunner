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

**NOTE** as of right now, variables are scoped to only their respective RR Block. You cannot share variables between blocks.
This is something we would like to add in the future.

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