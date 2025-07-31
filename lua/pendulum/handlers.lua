local M = {}

local lsp_client = nil

local function send_to_lsp(command, args)
    if not lsp_client or lsp_client.is_stopped() then
        return
    end

    lsp_client.request("workspace/executeCommand", {
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

local function init_lsp_client(opts)
    if lsp_client then
        return lsp_client
    end

    local stat = vim.loop.fs_stat(opts.lsp_binary)

    if not stat then
        vim.notify(
            "Pendulum LSP binary not found: " .. opts.lsp_binary,
            vim.log.levels.ERROR
        )
        return nil
    end

    if not vim.fn.executable(opts.lsp_binary) then
        vim.notify(
            "Pendulum LSP binary is not executable: " .. opts.lsp_binary,
            vim.log.levels.ERROR
        )
        return nil
    end

    local client_id = vim.lsp.start({
        name = "pendulum-lsp",
        cmd = {
            opts.binary_path,
            "--csv-path",
            opts.log_file,
            "--activity-timeout",
            tostring(opts.timeout_len),
            "--check-interval",
            tostring(opts.timer_len),
        },
        root_dir = vim.loop.cwd(),
        filetypes = {},
        on_attach = function(client, bufnr)
            vim.lsp.log.debug("Pendulum LSP attached")
        end,
        on_exit = function(code, signal, _)
            lsp_client = nil
            vim.notify(
                string.format(
                    "Pendulum LSP server exited with code: %s, signal: %s",
                    tostring(code),
                    tostring(signal)
                ),
                vim.log.levels.ERROR
            )
        end,
    })

    if client_id then
        lsp_client = vim.lsp.get_client_by_id(client_id)
        vim.lsp.log.debug(
            "Pendulum LSP client started with ID: " .. client_id,
            vim.log.levels.INFO
        )
    else
        vim.notify(
            "Failed to start Pendulum LSP server: " .. opts.lsp_binary,
            vim.log.levels.ERROR
        )
    end

    return lsp_client
end

local function ping_activity()
    send_to_lsp("pendulum.activityPing")
end

local function log_full_activity()
    local ft = vim.bo.filetype ~= "" and vim.bo.filetype or "unknown_filetype"

    local data = {
        time = os.date("!%Y-%m-%d %H:%M:%S"),
        active = true,
        file = vim.fn.expand("%:p"),
        filetype = ft,
        cwd = vim.loop.cwd(),
    }

    if data.file ~= "" then
        send_to_lsp("pendulum.logActivity", data)
    end
end

function M.get_lsp_client()
    return lsp_client
end

function M.setup(opts)
    opts = opts or {}
    opts.lsp_binary = opts.lsp_binary or "pendulum-lsp"
    opts.timeout_len = opts.timeout_len or 5
    opts.timer_len = opts.timer_len or 1

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

    vim.api.nvim_create_autocmd({ "BufEnter" }, {
        group = "Pendulum",
        callback = log_full_activity,
    })

    vim.api.nvim_create_autocmd({ "VimLeave" }, {
        group = "Pendulum",
        callback = function()
            if lsp_client and not lsp_client.is_stopped() then
                log_full_activity()
                send_to_lsp("pendulum.endSession")
            end
        end,
    })

    -- Initialize LSP client immediately (deferred to next tick to avoid start-up error)
    vim.defer_fn(function()
        init_lsp_client(opts)
    end, 0)
end

return M
