local M = {}

local chan
local bin_path
local plugin_path

local options = {}

--- job runner for pendulum remote binary
---@return integer?
local function ensure_job()
    if chan then
        return chan
    end

    if not bin_path then
        print("Error: Pendulum binary not found.")
        return
    end

    chan = vim.fn.jobstart({ bin_path }, {
        rpc = true,
        on_exit = function(_, code, _)
            if code ~= 0 then
                print("Error: Pendulum job exited with code " .. code)
                chan = nil
            end
        end,
        on_stderr = function(_, data, _)
            for _, line in ipairs(data) do
                if line ~= "" then
                    print("stderr: " .. line)
                end
            end
        end,
        on_stdout = function(_, data, _)
            for _, line in ipairs(data) do
                if line ~= "" then
                    print("stdout: " .. line)
                end
            end
        end,
    })

    if not chan or chan == 0 then
        error("Failed to start pendulum-nvim job")
    end

    return chan
end

--- create plugin user commands to build binary and show report
local function setup_pendulum_commands()
    vim.api.nvim_create_user_command("Pendulum", function(args)
        chan = ensure_job()
        if not chan or chan == 0 then
            print("Error: Invalid channel")
            return
        end

        -- local time_range = args.args or "all"
        local view = args.args

        local command_args = {
            log_file = options.log_file,
            timer_len = options.timer_len,
            top_n = options.top_n,
            hours_n = options.hours_n,
            -- time_range = time_range,
            view = view,
            time_range = "all",
            report_excludes = options.report_excludes,
            report_section_excludes = options.report_section_excludes,
            time_format = options.time_format,
        }

        local success, result =
            pcall(vim.fn.rpcrequest, chan, "pendulum", command_args)
        if not success then
            print("RPC request failed: " .. result)
        end
    end, { nargs = "?" })

    vim.api.nvim_create_user_command("PendulumRebuild", function()
        print("Rebuilding Pendulum binary with Go...")
        local result =
            os.execute("cd " .. plugin_path .. "remote" .. " && go build")
        if result == 0 then
            print("Go binary compiled successfully.")
            if chan then
                vim.fn.jobstop(chan)
                chan = nil
            end
        else
            print("Failed to compile Go binary.")
        end
    end, { nargs = 0 })
end

--- report generation setup (requires go)
---@param opts table
function M.setup(opts)
    options = opts

    -- get plugin install path
    plugin_path = debug.getinfo(1).source:sub(2):match("(.*/).*/.*/")

    -- check os to switch separators and binary extension if necessary
    local uname = vim.loop.os_uname().sysname
    local path_separator = (uname == "Windows_NT") and "\\" or "/"
    bin_path = plugin_path
        .. "remote"
        .. path_separator
        .. "pendulum-nvim"
        .. (uname == "Windows_NT" and ".exe" or "")

    setup_pendulum_commands()

    -- check if binary exists
    local uv = vim.loop
    local handle = uv.fs_open(bin_path, "r", 438)
    if handle then
        uv.fs_close(handle)
        return
    end

    -- TODO: check if go is installed and is correct version

    -- compile binary if it doesn't exist
    print(
        "Pendulum binary not found at "
            .. bin_path
            .. ", attempting to compile with Go..."
    )

    local result =
        os.execute("cd " .. plugin_path .. "remote" .. " && go build")
    if result == 0 then
        print("Go binary compiled successfully.")
    else
        print("Failed to compile Go binary." .. uv.cwd())
    end
end

return M
