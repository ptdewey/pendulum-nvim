local M = {}

local last_active_time = os.time()
local flag = true
local lsp_client = nil

local csv_path_set = false
local pending_queue = {}

local function update_activity()
    last_active_time = os.time()
end

local function flush_pending()
    if not lsp_client then
        return
    end

    if #pending_queue > 0 then
        for _, activity in ipairs(pending_queue) do
            lsp_client.request("workspace/executeCommand", {
                command = "pendulum.logActivity",
                arguments = { activity },
            }, function(err, _)
                if err then
                    vim.notify(
                        "Failed to log queued activity: "
                            .. tostring(err.message or err),
                        vim.log.levels.ERROR
                    )
                end
            end, 0)
        end
        pending_queue = {}
    end
end

local function send_to_lsp(activity_data)
    if not lsp_client or lsp_client.is_stopped() then
        -- vim.notify("Pendulum LSP client not initialized", vim.log.levels.WARN)
        return
    end

    if not csv_path_set then
        table.insert(pending_queue, activity_data)
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

    local client_id = vim.lsp.start({
        name = "pendulum-lsp",
        cmd = { opts.lsp_binary },
        root_dir = vim.fn.getcwd(),
        filetypes = {},
        on_attach = function(client, bufnr)
            client:request("workspace/executeCommand", {
                command = "pendulum.setCsvPath",
                arguments = { opts.log_file },
            }, function(err, _)
                if err then
                    vim.notify(
                        "Failed to set CSV path: "
                            .. tostring(err.message or err),
                        vim.log.levels.ERROR
                    )
                else
                    csv_path_set = true
                    flush_pending()
                end
            end, bufnr)
        end,
        on_init = function(_)
            flush_pending()
        end,
        on_exit = function(code, _, _)
            lsp_client = nil
            vim.notify(
                "Pendulum LSP server exited with code: " .. tostring(code),
                vim.log.levels.WARN
            )
        end,
    })

    if client_id then
        lsp_client = vim.lsp.get_client_by_id(client_id)
    else
        vim.notify(
            "Failed to start Pendulum LSP server: " .. opts.lsp_binary,
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
        active = tostring(is_active),
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
    if not is_active and flag then
        flag = false
        log_activity(true, last_active_time)
    elseif is_active and not flag then
        flag = true
    end
    log_activity(is_active)
end

function M.setup(opts)
    opts.lsp_binary = opts.lsp_binary or "pendulum-lsp"
    update_activity()

    vim.api.nvim_create_augroup("Pendulum", { clear = true })

    vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
        group = "Pendulum",
        callback = update_activity,
    })

    vim.api.nvim_create_autocmd({ "BufEnter", "VimLeave" }, {
        group = "Pendulum",
        callback = function()
            log_activity(true)
        end,
    })

    vim.defer_fn(function()
        init_lsp_client(opts)
        vim.fn.timer_start(opts.timer_len * 1000, function()
            vim.schedule(function()
                check_active_status(opts)
            end)
        end, { ["repeat"] = -1 })
    end, 100)
end

return M
