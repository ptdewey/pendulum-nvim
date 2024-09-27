local M = {}

local csv = require("pendulum.csv")

---initialize last active time
local last_active_time = os.time()

local flag = true

---update last active time
local function update_activity()
    last_active_time = os.time()
end

---get the name of the git project
---@return string
local function git_project()
    -- TODO: possibly change cwd to file path (to capture its git project while in a different working directory)
    local project_name = vim.system(
        { "git", "config", "--local", "remote.origin.url" },
        { text = true, cwd = vim.loop.cwd() }
    )
        :wait().stdout

    if project_name then
        project_name = project_name:gsub("%s+$", ""):match(".*/([^.]+)%.git$")
    end

    return project_name or "unknown_project"
end

---get name of current git branch
---@return string
local function git_branch()
    -- TODO: possibly change cwd to file path (to capture its git project while in a different working directory)
    local branch_name = vim.system(
        { "git", "branch", "--show-current" },
        { text = true, cwd = vim.loop.cwd() }
    )
        :wait().stdout

    if not branch_name or branch_name == "" or branch_name:match("^fatal:") then
        return "unknown_branch"
    end

    return branch_name:gsub("%s+$", "") or "unknown_branch"
end

---get table of tracked metrics
---@param is_active boolean
---@param active_time integer?
---@return table
local function log_activity(is_active, opts, active_time)
    -- TODO: allow adding a specific time, (last active, but in actual datetime format)
    -- https://stackoverflow.com/questions/32022898/subtracting-hours-from-os-date
    local _ = active_time
    local ft = vim.bo.filetype
    if ft == "" then
        ft = "unknown_filetype"
    end
    local data = {
        time = vim.fn.strftime("%Y-%m-%d %H:%M:%S"),
        active = tostring(is_active),
        -- file = vim.fn.expand("%:t+"), -- only file name
        file = vim.fn.expand("%:p"), -- file name with path
        -- TODO: file path - filename -> handoff to git to get file names
        -- - change cwd to file path without filename
        filetype = ft,
        cwd = vim.loop.cwd(),
        project = git_project(),
        branch = git_branch(),
    }
    if data.file ~= "" then
        csv.write_table_to_csv(opts.log_file, { data }, true)
    end

    return data
end

---Check if the user is currently active
---@param opts table
local function check_active_status(opts)
    local is_active = os.time() - last_active_time < opts.timeout_len

    -- for first non-active entry, log last active time
    if not is_active and flag then
        flag = false
        log_activity(true, opts, last_active_time)
    elseif is_active and not flag then
        flag = true
    end

    log_activity(is_active, opts)
end

---Setup periodic activity checks
---@param opts table
function M.setup(opts)
    update_activity()

    -- create autocommand group
    vim.api.nvim_create_augroup("Pendulum", { clear = true })

    -- define autocmd to update last active time
    vim.api.nvim_create_autocmd({ "CursorMoved", "CursorMovedI" }, {
        group = "Pendulum",
        callback = function()
            update_activity()
        end,
    })

    -- define autocmd for logging events
    vim.api.nvim_create_autocmd({ "BufEnter", "VimLeave" }, {
        group = "Pendulum",
        callback = function()
            log_activity(true, opts)
        end,
    })

    -- logging timer
    vim.fn.timer_start(opts.timer_len * 1000, function()
        vim.schedule(function()
            check_active_status(opts)
        end)
    end, { ["repeat"] = -1 })
end

return M
