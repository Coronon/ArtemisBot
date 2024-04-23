# ArtemisBot

ArtemisBot simplifies navigating the intricacies of the Artemis platform by automating the submission of programming exercises.

## Usage

```sh
artemisbot [command]
```

### Available Commands:

- `completion`: Generate the autocompletion script for the specified shell.
- `help`: Get help about any command.
- `retrigger`: Retrigger Artemis build tasks.

### Flags:

- `--config`: Specify the configuration file (default is `$HOME/.artemisbot.yaml`).
- `-h, --help`: Display help information.
- `-i, --interactive`: Enter credentials interactively.
- `-v, --verbose`: Enable verbose logging.
- `-d, --workdir`: Specify the directory to store data (default is `$TEMP_DIR`).

Use `artemisbot [command] --help` for more details about a specific command.

## `retrigger` Subcommand

Trigger the Artemis build task until the desired percentage is reached.

### Usage:

```sh
artemisbot retrigger [flags]
```

### Flags:

- `-t, --artemis-url`: URL of the Artemis task to automate (e.g. `"https://artemis.in.tum.de/courses/?/exercises/?"`).
- `-h, --help`: Display help for the `retrigger` command.
- `-p, --percentage`: Percentage of points to reach (default is `100`).

### Global Flags:

- `--config`: Specify the configuration file (default is `$HOME/.artemisbot.yaml`).
- `-i, --interactive`: Enter credentials interactively.
- `-v, --verbose`: Enable verbose logging.
- `-d, --workdir`: Specify the directory to store data (default is `$TEMP_DIR`).

All flags can be passed via command line, configuration file (JSON, YAML, or TOML), or environment variables prefixed with "ARTEMISBOT" (e.g., "ARTEMISBOT_NUMBER" for "--number").

**Note:** Artemis may experience occasional flakiness. In such cases, the tool simply restarts in the main loop to ensure seamless operation.

**Disclaimer:** The use of ArtemisBot is entirely at your own risk. The creator of ArtemisBot holds no responsibility for any damages or issues that may arise from its usage. Users are advised to use the program with caution and understand that any actions performed by ArtemisBot are irreversible. By using ArtemisBot, you agree to indemnify and hold harmless the creator from any liabilities, damages, or losses. Use it responsibly and ensure that you have appropriate permissions before automating any tasks on the Artemis platform.
