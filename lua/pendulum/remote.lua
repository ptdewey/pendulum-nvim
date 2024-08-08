local M = {}

local chan
local bin_path
local plugin_path

local options = {}

---report generation setup (requires go)
---@param opts table
function M.setup(opts)
    options.log_file = opts.log_file
    options.timer_len = opts.timer_len
    options.top_n = opts.top_n or 5

    local uname = vim.loop.os_uname().sysname
    local path_separator = (uname == "Windows_NT") and "\\" or "/"
    plugin_path = debug.getinfo(1).source:sub(2):match("(.*/).*/.*/")
    -- FIX: windows binary name is different, figure out how to use it

    bin_path = plugin_path .. "remote" .. path_separator .. "pendulum-nvim" .. (uname == "Windows_NT" and ".exe" or "")

    -- check if go binary exists
    local uv = vim.loop
    local handle = uv.fs_open(bin_path, "r", 438)
    if handle then
        uv.fs_close(handle)
        return
    end

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

vim.api.nvim_create_user_command("Pendulum", function()
    chan = ensure_job()
    if not chan or chan == 0 then
        print("Error: Invalid channel")
        return
    end
    local args =
        { options.log_file, "" .. options.timer_len, "" .. options.top_n }
    local success, result = pcall(vim.fn.rpcrequest, chan, "pendulum", args)
    if not success then
        print("RPC request failed: " .. result)
    end
end, { nargs = 0 })

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

return M
