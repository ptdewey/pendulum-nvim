local M = {}

---set up plugin autocommands with user options
---@param opts table?
function M.setup(opts)
    local handlers = require("pendulum.handlers")
    local remote = require("pendulum.remote")

    -- default plugin options
    local default_opts = {
        log_file = vim.env.HOME .. "/pendulum-log.csv",
        timeout_len = 180,
        timer_len = 120,
        top_n = 5,
        hours_n = 10,
        time_format = "12h",
        time_zone = "UTC", -- Format "America/New_York"
        report_excludes = {
            branch = {},
            directory = {},
            file = {},
            filetype = {},
            project = {},
        },
        report_section_excludes = {},
        lsp_binary = nil,
    }

    opts = vim.tbl_deep_extend("force", default_opts, opts or {})
    handlers.setup(opts)
    remote.setup(opts)
end

return M
