# ReadMe Runner Test File

This README contains test blocks for ReadMe Runner syntax.

## Commands Outside Blocks (Should Be Ignored)

The following commands are NOT inside HTML comment blocks and should be completely ignored:

```bash
echo "This command should be ignored - it's outside any RR block"
ls -la
pwd
```

## Basic RR Block

<!-- RR 
echo "This is a basic unnamed RR block!!!!"
-->

## Named RR Block

<!-- RR[Echo Test] 
echo "This is a named RR block called 'Echo Test'"
-->

## Variables

<!-- RR[Variables Test]
    my-var = "Hello from variable"
    echo #my-var
    another-var = "World"
    echo "Message: #my-var #another-var"
-->

## Prompts

<!-- RR[Prompt Test]
    my-name = #prompt("What is your name?")
    echo "Hello #my-name!"
-->

## Multi-line Commands

<!-- RR[Multi-line Test]
    echo "Starting process..." && \
    sleep 1 && \
    echo "Process complete!"
-->

## Combined Features

<!-- RR[Combined Test]
    greeting = "Welcome"
    user-name = #prompt("Enter your name:")
    echo "#greeting, #user-name!" && \
    echo "This is a multi-line command with variables and prompts"
-->

## Multiple Commands

<!-- RR[Multiple Commands]
    test-var = "Test Value"
    echo "First command with #test-var"
    echo "Second command"
    echo "Third command with #test-var again"
-->

## More Commands Outside Blocks (Should Be Ignored)

These commands are also outside HTML comment blocks and should be ignored:

```bash
echo "Another ignored command"
rm -rf /  # This dangerous command should be ignored since it's not in an RR block
```

## Environment Variables Test

<!-- RR[Env Variables Test]
    echo "Testing .env file variables:"
    echo "APP_NAME: #APP_NAME"
    echo "APP_VERSION: #APP_VERSION"
    echo "ENVIRONMENT: #ENVIRONMENT"
    echo "DATABASE_URL: #DATABASE_URL"
    echo "API_KEY: #API_KEY"
    echo "DEBUG: #DEBUG"
-->

## External tools
<!-- RR[Docker Hello]
docker run docker/whalesay cowsay Hello RR
-->

