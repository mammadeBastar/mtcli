# mtcli

A terminal-based typing test inspired by [Monkeytype](https://monkeytype.com). Practice your typing speed and accuracy right from your command line.

## Features

- **Multiple test modes**:

  - **Timer mode**: Type as many words as you can before time runs out
  - **Words mode**: Type a fixed number of words as fast as you can
  - **Quote mode**: Type famous quotes

- **Real-time feedback**: Characters change color as you type:

  - Gray: Not yet typed
  - White: Correct
  - Orange: Incorrect

- **Comprehensive metrics**:

  - WPM (Words Per Minute)
  - Raw WPM
  - Accuracy percentage
  - Speed chart over time

- **Progress tracking**: All results are saved locally in SQLite so you can track your improvement over time

## Installation

### From source

Requires Go 1.21+:

```bash
go install github.com/mmdbasi/mtcli/cmd/mtcli@latest
```

### Build locally

```bash
git clone https://github.com/mmdbasi/mtcli.git
cd mtcli
go build -o mtcli ./cmd/mtcli
```

## Usage

### Start a typing test

```bash
# Default: 25 words
mtcli test

# Timer mode - 60 seconds
mtcli test --mode timer --seconds 60

# Words mode - 50 words
mtcli test --mode words --words 50

# Quote mode - random quote
mtcli test --mode quote

# Quote mode - specific quote
mtcli test --mode quote --quote-id 5
```

### View your statistics

```bash
# Show aggregate statistics
mtcli stats

# Show test history
mtcli history

# Show history filtered by mode
mtcli history --mode timer

# Show more/fewer results
mtcli history --limit 50

# Show details of a specific test
mtcli show 42
```

### Command-line options

#### Test command

| Flag             | Description                             | Default |
| ---------------- | --------------------------------------- | ------- |
| `-m, --mode`     | Test mode: `timer`, `words`, or `quote` | `words` |
| `-s, --seconds`  | Duration in seconds (timer mode)        | `30`    |
| `-w, --words`    | Number of words (words mode)            | `25`    |
| `--quote-id`     | Specific quote ID (quote mode)          | -       |
| `--quote-random` | Use random quote (quote mode)           | `true`  |
| `--countdown`    | Countdown seconds before test starts    | `3`     |
| `--seed`         | Random seed for reproducible tests      | -       |
| `--no-color`     | Disable color output                    | `false` |
| `--wrap`         | Text wrap width (0 for auto)            | `0`     |
| `--chart`        | Show speed chart at end                 | `true`  |
| `--words-file`   | Custom words file                       | -       |
| `--quotes-file`  | Custom quotes file                      | -       |

#### History command

| Flag          | Description                | Default |
| ------------- | -------------------------- | ------- |
| `-n, --limit` | Number of sessions to show | `20`    |
| `-m, --mode`  | Filter by mode             | -       |

## Configuration

You can set default values in a config file at `~/.config/mtcli/config.toml`:

```toml
mode = "words"
seconds = 30
words = 25
countdown = 3
no_color = false
chart = true
```

Environment variables with the prefix `MTCLI_` are also supported:

```bash
export MTCLI_MODE=timer
export MTCLI_SECONDS=60
```

## Custom content

### Custom word list

Create a text file with one word per line:

```
hello
world
typing
practice
```

Use it with:

```bash
mtcli test --words-file /path/to/words.txt
```

### Custom quotes

Create a JSON file with quotes:

```json
[
  {
    "id": "1",
    "text": "Your custom quote here.",
    "source": "Author Name"
  }
]
```

Use it with:

```bash
mtcli test --mode quote --quotes-file /path/to/quotes.json
```

## Data storage

Test results are stored in a SQLite database at:

- **macOS**: `~/Library/Application Support/mtcli/mtcli.db`
- **Linux**: `~/.config/mtcli/mtcli.db`
- **Windows**: `%APPDATA%/mtcli/mtcli.db`

## Controls

During a test:

- Type characters to match the target text
- **Backspace**: Delete the last typed character
- **Ctrl+C** or **Escape**: Abort the test

## Understanding metrics

- **WPM (Words Per Minute)**: Calculated as `(correct characters / 5) / minutes`. This is your "net" speed.
- **Raw WPM**: Calculated as `(total typed characters / 5) / minutes`. Includes mistakes.
- **Accuracy**: Percentage of correctly typed characters: `correct / total * 100`

The speed chart shows both WPM (solid blocks) and Raw WPM (light blocks) over time, helping you see consistency.

## Troubleshooting

### Colors not displaying

If colors aren't showing properly:

1. Make sure your terminal supports 256 colors
2. Try using a different terminal emulator
3. Use `--no-color` flag for plain output

### Terminal not resetting after crash

If the terminal is in a weird state after the program crashes:

```bash
reset
# or
stty sane
```

### Database issues

To reset your data, delete the database file:

```bash
rm ~/Library/Application\ Support/mtcli/mtcli.db  # macOS
rm ~/.config/mtcli/mtcli.db                        # Linux
```
