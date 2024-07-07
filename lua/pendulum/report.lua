local csv = require("pendulum.csv")

local M = {}

local function parse_log_file(log_file)
    local data = csv.read_csv(log_file)
    return data
end

--- TODO:
---@param data table
---@return table
local function aggregate_metrics(data)
    -- TODO: add param for time frame
    local files = {}
    local projects = {}
    local branches = {}
    local workdirs = {}
    local fts = {}

    for _, v in pairs(data) do
        -- TODO: create running total, average for metrics
        -- - time spent per project, file, language
        -- - language should also be collected for outside of specific project
        -- - allow option to only show result if currently in git dir?
        if v.active then
            -- file names
            if files[v.file] == nil then
                files[v.file] = { count = 1, time = 0 }
            else
                files[v.file].count = files[v.file].count + 1
                -- TODO: time calculations (make function outside of this to use with other metrics as well)
            end

            -- file types
            -- TODO: turn this into a function since it will be common across metrics
            -- - take in table and v.metric name
            if fts[v.ft] == nil then
                fts[v.ft] = { count = 1, time = 0 }
            else
                fts[v.ft].count = fts[v.ft].count + 1
            end
        else
            -- idle
            -- TODO: total idle time, most idle file type (spent googling/gpting?)
        end
    end

    print(vim.inspect(files))

    return { files, fts, projects, branches, workdirs }
end

vim.api.nvim_create_user_command("TestPendulum", function()
    local data =
        parse_log_file(vim.fn.expand("$HOME/projects/pendulum-log.csv"))
    -- print(vim.inspect(data))
    aggregate_metrics(data)
end, {})

return M
