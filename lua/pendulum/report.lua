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
    for key, value in pairs(data) do
        -- TODO: create running total, average for metrics
        -- - time spent per project, file, language
        -- - language should also be collected for outside of specific project
        -- - allow option to only show result if currently in git dir?
        print(key, value)
    end

    return {}
end

return M
