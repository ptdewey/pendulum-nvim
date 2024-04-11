-- init.lua
local M = {}

local handlers = require("pendulum.handlers")

-- default plugin options
local default_opts = {
    log_file =  vim.env.HOME .. "/pendulum-log.csv",
    timeout_len = 180,
    timer_len = 120,
}

---set up plugin autocommands with user options
---@param opts table?
function M.setup(opts)
    opts = vim.tbl_deep_extend("force", default_opts, opts)
    handlers.setup(opts)
end

return M
