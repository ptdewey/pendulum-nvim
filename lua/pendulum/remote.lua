local chan

local function ensure_job()
    if chan then
        return chan
    end

    -- Start the job and ensure it is running
    chan = vim.fn.jobstart({ "remote/pendulum-nvim" }, {
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

    print("Job started with channel ID: " .. chan)
    return chan
end

vim.api.nvim_create_user_command("Pendulum", function(args)
    chan = ensure_job()
    if not chan or chan == 0 then
        print("Error: Invalid channel")
        return
    end

    local success, result =
        pcall(vim.fn.rpcrequest, chan, "pendulum", args.fargs)
    if not success then
        print("RPC request failed: " .. result)
    end
end, { nargs = "*" })
