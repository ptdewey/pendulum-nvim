# Pendulum-nvim

Pendulum is a Neovim plugin designed for tracking time spent on projects within Neovim. It logs various events like entering and leaving buffers and idle times into a CSV file, making it easy to analyze your coding activity over time.

## Motivation

Pendulum was created to offer a privacy-focused alternative to cloud-based time tracking tools, addressing concerns about data security and ownership. This "local-first" tool ensures all data stays on the user's machine, providing full control and customization without requiring internet access. It's designed for developers who prioritize privacy and autonomy but still want to monitor their coding activities.

## Features

- **Automatic Time Tracking**: Logs time spent in each file along with the project name and git branch, if available.
- **Activity Detection**: Detects user activity based on cursor movements, buffer switches, and edits.
- **Customizable Timeout**: Configurable timeout to define user inactivity.
- **Event Logging**: Tracks buffer events and idle periods, writing these to a CSV log for later analysis.

## Installation

Install Pendulum using your favorite package manager:

Lazy:
```lua
{
    "ptdewey/pendulum-nvim",
    config = function()
        require("pendulum").setup({
            log_file = vim.fn.expand("$HOME/Documents/pendulum-log.csv"),
            timeout_len = 180,  -- in seconds
            timer_len = 120,    -- in seconds
        })
    end,
}
```

Packer:
```lua
use {
    "ptdewey/pendulum-nvim",
    config = function()
        require("pendulum").setup({
            log_file = vim.fn.expand("$HOME/Documents/pendulum-log.csv"),
            timeout_len = 180,
            timer_len = 120,
        })
    end
}
```

## Configuration

Pendulum can be customized with several options. Here is a table with configurable options:

| Option      | Description                                       | Default                             |
|-------------|---------------------------------------------------|-------------------------------------|
| `log_file`  | Path to the CSV file where logs should be written | `$HOME/Documents/pendulum-log.csv` |
| `timeout_len` | Length of time in seconds to determine inactivity | `180`                               |
| `timer_len` | Interval in seconds at which to check activity    | `120`                               |

Example configuration with custom options:

```lua
require('pendulum').setup({
    log_file = vim.fn.expand("$HOME/Documents/my_custom_log.csv"),
    timeout_len = 300,  -- 5 minutes
    timer_len = 60,     -- 1 minute
})
```

## Usage

Once configured, Pendulum runs automatically in the background. It logs each specified event into the CSV file, which includes timestamps, file names, project names (from Git), and activity states.

The CSV log file will have the columns: `time`, `active`, `file`, `filetype`, `cwd`, `project`, and `branch`.
- time: Log timestamp
- active: If the user is currently active
- file: Current filename
- filetype: Current file filetype
- cwd: current working directory
- project: current git project
- branch: current git branch

