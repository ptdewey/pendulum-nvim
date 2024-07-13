local M = {}

local handlers = require("pendulum.handlers")
local remote = require("pendulum.remote")

-- default plugin options
local default_opts = {
    log_file = vim.env.HOME .. "/pendulum-log.csv",
    timeout_len = 180,
    timer_len = 120,
    gen_reports = nil,
    top_n = 5,
}

---set up plugin autocommands with user options
---@param opts table?
function M.setup(opts)
    opts = vim.tbl_deep_extend("force", default_opts, opts or {})
    handlers.setup(opts)

    if opts.gen_reports then
        remote.setup(opts)
    end
end

return M
