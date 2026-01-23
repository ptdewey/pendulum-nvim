local M = {}

local lsp_client = nil
local lsp_ready = false
local message_queue = {}
local stored_opts = nil

-- Get the plugin installation path
local function get_plugin_path()
    -- Extract path from this file's location: .../pendulum-nvim/lua/pendulum/handlers.lua
    return debug.getinfo(1).source:sub(2):match("(.*/)lua/pendulum/")
end

-- Get the default binary path based on OS
local function get_default_bin_path()
    local path = get_plugin_path()
    if not path then
        return nil
    end

    local uname = vim.loop.os_uname().sysname
    local path_separator = (uname == "Windows_NT") and "\\" or "/"
    local bin_name = (uname == "Windows_NT") and "pendulum-lsp.exe" or "pendulum-lsp"

    return path .. "bin" .. path_separator .. bin_name
end

local function flush_queue()
    if not lsp_ready or not lsp_client or lsp_client:is_stopped() then
        return
    end

    for _, msg in ipairs(message_queue) do
        lsp_client:request("workspace/executeCommand", {
            command = msg.command,
            arguments = msg.args and { msg.args } or {},
        }, function(err, _)
            if err then
                vim.notify(
                    "Failed to execute "
                        .. msg.command
                        .. ": "
                        .. tostring(err.message or err),
                    vim.log.levels.ERROR
                )
            end
        end, 0)
    end

    message_queue = {}
end

local function send_to_lsp(command, args)
    local msg = { command = command, args = args }

    if not lsp_ready or not lsp_client or lsp_client:is_stopped() then
        -- Queue the message to send when connection is ready
        table.insert(message_queue, msg)
        return
    end

    lsp_client:request("workspace/executeCommand", {
        command = command,
        arguments = args and { args } or {},
    }, function(err, _)
        if err then
            vim.notify(
                "Failed to execute "
                    .. command
                    .. ": "
                    .. tostring(err.message or err),
                vim.log.levels.ERROR
            )
        end
    end, 0)
end

-- Synchronous version of send_to_lsp for critical operations like VimLeave
local function send_to_lsp_sync(command, args, timeout_ms)
    timeout_ms = timeout_ms or 1000

    if not lsp_ready or not lsp_client or lsp_client:is_stopped() then
        return false
    end

    local done = false
    local success = false

    lsp_client:request("workspace/executeCommand", {
        command = command,
        arguments = args and { args } or {},
    }, function(err, _)
        if err then
            vim.notify(
                "Failed to execute "
                    .. command
                    .. ": "
                    .. tostring(err.message or err),
                vim.log.levels.ERROR
            )
            success = false
        else
            success = true
        end
        done = true
    end, 0)

    -- Wait for the request to complete or timeout
    local wait_result = vim.wait(timeout_ms, function()
        return done
    end, 10)

    return wait_result and success
end

local function ping_activity()
    -- Skip special buffer types (terminal, quickfix, help, etc.)
    local buftype = vim.bo.buftype
    if buftype ~= "" then
        return
    end

    send_to_lsp("pendulum.activityPing")
end

local function log_full_activity(filepath)
    -- Skip special buffer types (terminal, quickfix, help, etc.)
    local buftype = vim.bo.buftype
    if buftype ~= "" then
        return
    end

    -- Use provided filepath or get current buffer's path
    local file = filepath or vim.fn.expand("%:p")
    if file == "" then
        return
    end

    -- Skip terminal buffers and other non-file buffers
    if file:match("^term://") or file:match("^fugitive://") or file:match("^oil://") then
        return
    end

    local ft = vim.bo.filetype ~= "" and vim.bo.filetype or "unknown_filetype"

    local data = {
        time = os.date("!%Y-%m-%d %H:%M:%S"),
        active = true,
        file = file,
        filetype = ft,
        cwd = vim.loop.cwd(),
    }

    send_to_lsp("pendulum.logActivity", data)
end

local function init_lsp_client(opts)
    if lsp_client and not lsp_client:is_stopped() then
        return lsp_client
    end

    -- Reset state
    lsp_client = nil
    lsp_ready = false

    if not opts.lsp_binary then
        -- Binary path not set, remote.lua will handle building
        return nil
    end

    local stat = vim.loop.fs_stat(opts.lsp_binary)

    if not stat then
        -- Binary doesn't exist, remote.lua will handle building
        return nil
    end

    if vim.fn.executable(opts.lsp_binary) ~= 1 then
        vim.notify(
            "Pendulum LSP binary is not executable: " .. opts.lsp_binary,
            vim.log.levels.ERROR
        )
        return nil
    end

    local client_id = vim.lsp.start({
        name = "pendulum-lsp",
        cmd = {
            opts.lsp_binary,
            "--csv-path",
            opts.log_file,
            "--activity-timeout",
            tostring(opts.timeout_len),
            "--check-interval",
            tostring(opts.timer_len),
        },
        root_dir = vim.loop.cwd(),
        filetypes = {},
        on_init = function(client, initialize_result)
            -- LSP handshake complete - server is ready to receive commands
            -- Set lsp_client immediately since on_init fires before vim.lsp.start() returns
            lsp_client = client
            lsp_ready = true
            -- Flush any queued messages
            vim.schedule(function()
                flush_queue()
            end)
        end,
        on_attach = function(client, bufnr)
            vim.lsp.log.debug("Pendulum LSP attached")
        end,
        on_exit = function(code, signal, _)
            lsp_client = nil
            lsp_ready = false
            if code ~= 0 then
                vim.notify(
                    string.format(
                        "Pendulum LSP server exited with code: %s, signal: %s",
                        tostring(code),
                        tostring(signal)
                    ),
                    vim.log.levels.WARN
                )
            end
        end,
    })

    if client_id then
        lsp_client = vim.lsp.get_client_by_id(client_id)
        vim.lsp.log.debug(
            "Pendulum LSP client started with ID: " .. client_id,
            vim.log.levels.INFO
        )
    end

    return lsp_client
end

function M.get_lsp_client()
    return lsp_client
end

-- Reinitialize LSP client (called after binary is built)
function M.reinit_lsp()
    if stored_opts then
        return init_lsp_client(stored_opts)
    end
    return nil
end

function M.setup(opts)
    opts = opts or {}
    opts.timeout_len = opts.timeout_len or 5
    opts.timer_len = opts.timer_len or 1

    -- Determine binary path - use provided path or default to plugin bin directory
    if not opts.lsp_binary then
        opts.lsp_binary = get_default_bin_path()
    end

    -- Store opts for potential reinit
    stored_opts = opts

    if not opts.log_file then
        vim.notify(
            "Pendulum: log_file is required in setup options",
            vim.log.levels.ERROR
        )
        return
    end

    vim.api.nvim_create_augroup("Pendulum", { clear = true })

    vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
        group = "Pendulum",
        callback = ping_activity,
    })

    vim.api.nvim_create_autocmd(
        { "BufReadPost", "BufEnter", "BufLeave" },
        {
            group = "Pendulum",
            callback = function()
                log_full_activity()
            end,
        }
    )

    -- VimEnter needs special handling - the buffer may not be ready yet
    vim.api.nvim_create_autocmd({ "VimEnter" }, {
        group = "Pendulum",
        callback = function()
            -- Defer to ensure buffer is loaded when opening nvim with a file argument
            vim.defer_fn(function()
                local file = vim.fn.expand("%:p")
                if file ~= "" then
                    log_full_activity(file)
                end
            end, 10)
        end,
    })

    vim.api.nvim_create_autocmd({ "VimLeave" }, {
        group = "Pendulum",
        callback = function()
            if lsp_client and not lsp_client:is_stopped() then
                -- Use synchronous logging for VimLeave to ensure logs are written before exit
                -- Skip special buffer types
                local buftype = vim.bo.buftype
                local file = vim.fn.expand("%:p")
                if file ~= "" and buftype == "" and not file:match("^term://") and not file:match("^fugitive://") and not file:match("^oil://") then
                    local ft = vim.bo.filetype ~= "" and vim.bo.filetype
                        or "unknown_filetype"
                    local data = {
                        time = os.date("!%Y-%m-%d %H:%M:%S"),
                        active = true,
                        file = file,
                        filetype = ft,
                        cwd = vim.loop.cwd(),
                    }
                    send_to_lsp_sync("pendulum.logActivity", data, 1000)
                end
                send_to_lsp_sync("pendulum.endSession", nil, 500)
            end
        end,
    })

    -- Initialize LSP client (will silently fail if binary doesn't exist)
    init_lsp_client(opts)
end

return M
