local M = {}

-- TODO: strip most of this code out into the LSP

local last_active_time = os.time()
local active_flag = true
local lsp_client = nil

local function update_activity()
    last_active_time = os.time()
end

local function send_to_lsp(activity_data)
    if not lsp_client or lsp_client.is_stopped() then
        return
    end

    lsp_client.request("workspace/executeCommand", {
        command = "pendulum.logActivity",
        arguments = { activity_data },
    }, function(err, _)
        if err then
            vim.notify(
                "Failed to log activity: " .. tostring(err.message or err),
                vim.log.levels.ERROR
            )
        end
    end, 0)
end

local function init_lsp_client(opts)
    if lsp_client then
        return lsp_client
    end

    -- Check if the binary exists and is executable
    local binary_path = opts.lsp_binary
    local stat = vim.loop.fs_stat(binary_path)

    if not stat then
        vim.notify(
            "Pendulum LSP binary not found: " .. binary_path,
            vim.log.levels.ERROR
        )
        return nil
    end

    if not vim.fn.executable(binary_path) then
        vim.notify(
            "Pendulum LSP binary is not executable: " .. binary_path,
            vim.log.levels.ERROR
        )
        return nil
    end

    local client_id = vim.lsp.start({
        name = "pendulum-lsp",
        cmd = { binary_path, "--csv-path", opts.log_file },
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
                vim.log.levels.WARN
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
            "Failed to start Pendulum LSP server: " .. binary_path,
            vim.log.levels.ERROR
        )
    end

    return lsp_client
end

local function log_activity(is_active, active_time)
    local time = active_time or os.time()
    local ft = vim.bo.filetype ~= "" and vim.bo.filetype or "unknown_filetype"

    local data = {
        time = os.date("!%Y-%m-%d %H:%M:%S", time),
        active = is_active,
        file = vim.fn.expand("%:p"),
        filetype = ft,
        cwd = vim.loop.cwd(),
    }

    if data.file ~= "" then
        send_to_lsp(data)
    end

    return data
end

local function check_active_status(opts)
    local is_active = os.time() - last_active_time < opts.timeout_len
    if not is_active and active_flag then
        active_flag = false
        log_activity(true, last_active_time)
    elseif is_active and not active_flag then
        active_flag = true
    end
    log_activity(is_active)
end

-- Expose the LSP client to other modules
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

    update_activity()

    vim.api.nvim_create_augroup("Pendulum", { clear = true })

    vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
        group = "Pendulum",
        callback = update_activity,
    })

    vim.api.nvim_create_autocmd({ "BufEnter" }, {
        group = "Pendulum",
        callback = function()
            log_activity(true)
        end,
    })

    vim.api.nvim_create_autocmd({ "VimLeave" }, {
        group = "Pendulum",
        callback = function()
            if lsp_client and not lsp_client.is_stopped() then
                log_activity(true)
            end
        end,
    })

    -- Initialize LSP client immediately (deferred to next tick)
    vim.defer_fn(function()
        init_lsp_client(opts)
    end, 0)

    -- Start the activity checking timer
    vim.defer_fn(function()
        vim.fn.timer_start(opts.timer_len * 1000, function()
            vim.schedule(function()
                check_active_status(opts)
            end)
        end, { ["repeat"] = -1 })
    end, 100)
end

return M
