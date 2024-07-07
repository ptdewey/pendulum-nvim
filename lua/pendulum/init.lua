local M = {}

local handlers = require("pendulum.handlers")
require("pendulum.report")

-- default plugin options
local default_opts = {
    log_file = vim.env.HOME .. "/pendulum-log.csv",
    timeout_len = 180,
    timer_len = 120,
    -- option for go remote?
}

---set up plugin autocommands with user options
---@param opts table?
function M.setup(opts)
    opts = vim.tbl_deep_extend("force", default_opts, opts or {})
    handlers.setup(opts)
    -- process call to build go binary if it doesnt exist
end

return M
