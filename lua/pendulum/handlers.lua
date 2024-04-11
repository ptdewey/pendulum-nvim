-- handlers.lua
local M = {}

local csv = require("pendulum.csv")

-- initialize last active time
local last_active_time = os.time()

---update last active time
local function update_activity()
    last_active_time = os.time()
end

---get the name of the git project
---@return string
local function git_project()
    local project_name = vim.fn.system("git config --local remote.origin.url")
    project_name = project_name:gsub("%s+$", ""):match(".*/([^.]+)%.git$")
    return project_name or "unknown_project"
end

---get name of current git branch
---@return string
local function git_branch()
    local branch_name = vim.fn.system("git branch --show-current")

    -- Check for errors or empty output
    if branch_name == "" or branch_name:match("^fatal:") then
        return "unknown_branch"
    end

    return branch_name:gsub("%s+$", "") or "unknown_branch"
end

---get table of tracked metrics
---@param is_active boolean
---@return table
local function log_activity(is_active, opts)
    local data = {
        time = vim.fn.strftime("%Y-%m-%d %H:%M:%S"),
        active = tostring(is_active),
        file = vim.fn.expand("%:t+"),
        filetype = vim.bo.filetype,
        cwd = vim.loop.cwd(),
        project = git_project(),
        branch = git_branch(),
    }

    csv.write_table_to_csv(opts.log_file, {data}, true)

    return data
end

---Check if the user is currently active
---@param opts table
local function check_active_status(opts)
    log_activity(os.time() - last_active_time < opts.timeout_len, opts)
end

-- Setup periodic activity checks
---@param opts table
function M.setup(opts)
    update_activity()

    -- create autocommand group
    vim.api.nvim_create_augroup("ActivityTracker", { clear = true })

    -- define autocmd to update last active time
    vim.api.nvim_create_autocmd({"CursorMoved", "CursorMovedI" }, {
        group = "ActivityTracker",
        callback = function()
            update_activity()
        end,
    })

    -- define autocmd for logging events
    vim.api.nvim_create_autocmd({ "BufEnter", "VimLeave" }, {
        group = "ActivityTracker",
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
