local M = {}

local options = {}
local plugin_path = nil
local bin_path = nil

-- Get the plugin installation path
local function get_plugin_path()
    if plugin_path then
        return plugin_path
    end
    -- Extract path from this file's location: .../pendulum-nvim/lua/pendulum/remote.lua
    plugin_path = debug.getinfo(1).source:sub(2):match("(.*/)lua/pendulum/")
    return plugin_path
end

-- Get the binary path based on OS
local function get_bin_path()
    if bin_path then
        return bin_path
    end

    local path = get_plugin_path()
    if not path then
        return nil
    end

    local uname = vim.loop.os_uname().sysname
    local path_separator = (uname == "Windows_NT") and "\\" or "/"
    local bin_name = (uname == "Windows_NT") and "pendulum-lsp.exe"
        or "pendulum-lsp"

    bin_path = path .. "bin" .. path_separator .. bin_name
    return bin_path
end

-- Check if the binary exists
local function binary_exists()
    local path = get_bin_path()
    if not path then
        return false
    end

    local stat = vim.loop.fs_stat(path)
    return stat ~= nil
end

-- Build the LSP binary
local function build_binary(callback)
    local path = get_plugin_path()
    if not path then
        vim.notify("Could not determine plugin path", vim.log.levels.ERROR)
        if callback then
            callback(false)
        end
        return
    end

    vim.notify("Building Pendulum LSP binary with Go...", vim.log.levels.INFO)

    local target_bin = get_bin_path()
    local build_cmd = string.format(
        "cd %s && go build -o %s .",
        vim.fn.shellescape(path),
        vim.fn.shellescape(target_bin)
    )

    vim.fn.jobstart(build_cmd, {
        on_exit = function(_, code, _)
            if code == 0 then
                vim.schedule(function()
                    vim.notify(
                        "Pendulum LSP binary compiled successfully.",
                        vim.log.levels.INFO
                    )
                    if callback then
                        callback(true)
                    end
                end)
            else
                vim.schedule(function()
                    vim.notify(
                        "Failed to compile Pendulum LSP binary. Make sure Go is installed.",
                        vim.log.levels.ERROR
                    )
                    if callback then
                        callback(false)
                    end
                end)
            end
        end,
        on_stderr = function(_, data, _)
            for _, line in ipairs(data) do
                if line ~= "" then
                    vim.schedule(function()
                        vim.notify("Build: " .. line, vim.log.levels.WARN)
                    end)
                end
            end
        end,
    })
end

-- Function to create a buffer with the content received from the LSP
local function create_buffer(content, filetype)
    -- Create a new scratch buffer
    local buf = vim.api.nvim_create_buf(false, true)

    -- Set buffer options
    vim.api.nvim_set_option_value(
        "filetype",
        filetype or "markdown",
        { buf = buf }
    )

    -- Process content - LSP always returns a string
    local lines = {}

    if type(content) == "string" then
        -- Split string on newlines and add each line
        for line in content:gmatch("[^\r\n]+") do
            table.insert(lines, line)
        end
    end

    -- Set buffer content line by line
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)

    -- Set buffer keymap for close
    vim.api.nvim_buf_set_keymap(
        buf,
        "n",
        "q",
        "<cmd>close!<CR>",
        { silent = true }
    )

    return buf
end

-- Function to create a popup window with a buffer
local function create_popup_window(buf)
    -- Get screen dimensions
    local screen_width = vim.api.nvim_get_option_value("columns", {})
    local screen_height = vim.api.nvim_get_option_value("lines", {})

    -- Calculate popup dimensions
    local popup_width = math.floor(screen_width * 0.85)
    local popup_height = math.floor(screen_height * 0.85)

    -- Create window config
    local win_config = {
        relative = "editor",
        row = math.floor((screen_height - popup_height) / 2) - 1,
        col = math.floor((screen_width - popup_width) / 2),
        width = popup_width,
        height = popup_height,
        style = "minimal",
        border = "rounded",
        zindex = 50,
    }

    -- Open window with buffer
    local win = vim.api.nvim_open_win(buf, true, win_config)

    return win
end

-- Function to handle LSP report generation responses
local function handle_lsp_response(err, result, context)
    if err then
        vim.notify(
            "Error generating report: " .. vim.inspect(err),
            vim.log.levels.ERROR
        )
        return
    end

    if not result then
        vim.notify("No data returned from LSP", vim.log.levels.WARN)
        return
    end

    -- Create buffer with the content
    local buf = create_buffer(result, "markdown")

    -- Create popup window
    create_popup_window(buf)
end

-- Setup pendulum commands for report generation
local function setup_pendulum_commands(lsp_client)
    -- Valid time range options for command completion
    local time_range_options = { "all", "day", "week", "month", "year", "hour" }

    vim.api.nvim_create_user_command("Pendulum", function(args)
        if not lsp_client or lsp_client:is_stopped() then
            vim.notify(
                "Pendulum LSP client not available. Try :PendulumRebuild",
                vim.log.levels.ERROR
            )
            return
        end

        -- Parse time range from command argument (e.g., :Pendulum week)
        local time_range = "all"
        if args.args and args.args ~= "" then
            time_range = args.args
        end
        options.time_range = time_range
        options.view = "metrics"

        -- Send request to LSP for metrics report
        lsp_client:request("workspace/executeCommand", {
            command = "pendulum.generateMetricsReport",
            arguments = { options },
        }, handle_lsp_response)
    end, {
        nargs = "?",
        force = true,
        complete = function(arg_lead, cmd_line, cursor_pos)
            -- Filter options based on what user has typed
            local matches = {}
            for _, opt in ipairs(time_range_options) do
                if opt:find("^" .. arg_lead) then
                    table.insert(matches, opt)
                end
            end
            return matches
        end,
    })

    vim.api.nvim_create_user_command("PendulumHours", function()
        if not lsp_client or lsp_client:is_stopped() then
            vim.notify(
                "Pendulum LSP client not available. Try :PendulumRebuild",
                vim.log.levels.ERROR
            )
            return
        end

        options.view = "hours"

        -- Send request to LSP for hours report
        lsp_client:request("workspace/executeCommand", {
            command = "pendulum.generateHourlyReport",
            arguments = { options },
        }, handle_lsp_response)
    end, { nargs = 0, force = true })
end

-- Setup the PendulumRebuild command (always available)
local function setup_rebuild_command()
    vim.api.nvim_create_user_command("PendulumRebuild", function()
        -- Stop existing LSP client if running
        local handlers = require("pendulum.handlers")
        local client = handlers.get_lsp_client()
        if client and not client:is_stopped() then
            vim.notify(
                "Stopping Pendulum LSP before rebuild...",
                vim.log.levels.INFO
            )
            client:stop()
        end

        build_binary(function(success)
            if success then
                -- Reinitialize the LSP client
                vim.defer_fn(function()
                    local new_client = handlers.reinit_lsp()
                    if new_client then
                        setup_pendulum_commands(new_client)
                        vim.notify(
                            "Pendulum LSP reinitialized.",
                            vim.log.levels.INFO
                        )
                    else
                        vim.notify(
                            "Pendulum LSP binary built. Restart Neovim to use it.",
                            vim.log.levels.INFO
                        )
                    end
                end, 500)
            end
        end)
    end, { nargs = 0, force = true })
end

-- Report generation setup using LSP
function M.setup(opts)
    options = opts

    -- Determine binary path - use provided path or default to plugin bin directory
    if not opts.lsp_binary then
        opts.lsp_binary = get_bin_path()
    end

    -- Always set up the PendulumRebuild command
    setup_rebuild_command()

    -- Check if binary exists, build if missing
    if not binary_exists() then
        vim.notify(
            "Pendulum LSP binary not found, attempting to build...",
            vim.log.levels.INFO
        )
        build_binary(function(success)
            if success then
                -- Reinitialize the LSP client after build
                vim.defer_fn(function()
                    local handlers = require("pendulum.handlers")
                    local new_client = handlers.reinit_lsp()
                    if new_client then
                        setup_pendulum_commands(new_client)
                    end
                end, 500)
            end
        end)
        return
    end

    -- Binary exists, initialize commands
    M.initialize_lsp_commands()
end

-- Initialize LSP commands (called after binary is available)
function M.initialize_lsp_commands()
    -- Get LSP client from handlers module
    local handlers = require("pendulum.handlers")
    local lsp_client = handlers.get_lsp_client()

    -- Set up commands if client is available or wait for it
    if lsp_client then
        setup_pendulum_commands(lsp_client)
    else
        -- If client not available yet, wait for it to be initialized
        vim.defer_fn(function()
            local client = handlers.get_lsp_client()
            if client then
                setup_pendulum_commands(client)
            else
                vim.notify(
                    "Pendulum LSP client not available. Try :PendulumRebuild",
                    vim.log.levels.WARN
                )
            end
        end, 1000) -- Wait 1 second for LSP to initialize
    end
end

return M
