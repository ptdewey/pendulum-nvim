local M = {}

local options = {}

-- Function to create a buffer with the content received from the LSP
local function create_buffer(content, filetype)
    -- Create a new scratch buffer
    local buf = vim.api.nvim_create_buf(false, true)

    -- Set buffer options
    vim.api.nvim_buf_set_option(buf, "filetype", filetype or "markdown")

    -- Process content based on type
    local lines = {}

    if type(content) == "table" then
        -- If content is an array
        for _, section in ipairs(content) do
            if type(section) == "string" then
                -- If it's a string, split it on newlines and add each line
                for line in section:gmatch("[^\r\n]+") do
                    table.insert(lines, line)
                end
            elseif type(section) == "table" then
                -- If it's an array of strings
                for _, line in ipairs(section) do
                    if type(line) == "string" then
                        table.insert(lines, line)
                    end
                end
            end
        end
    elseif type(content) == "string" then
        -- For a single string
        for line in content:gmatch("[^\r\n]+") do
            table.insert(lines, line)
        end
    end

    -- Set buffer content line by line
    vim.api.nvim_buf_set_lines(buf, 0, -1, false, lines)

    -- Set buffer keymap for close
    vim.api.nvim_buf_set_keymap(
        buf,
        "n",
        "q",
        "<cmd>close!<CR>",
        { silent = true }
    )

    return buf
end

-- Function to create a popup window with a buffer
local function create_popup_window(buf)
    -- Get screen dimensions
    local screen_width = vim.api.nvim_get_option("columns")
    local screen_height = vim.api.nvim_get_option("lines")

    -- Calculate popup dimensions
    local popup_width = math.floor(screen_width * 0.85)
    local popup_height = math.floor(screen_height * 0.85)

    -- Create window config
    local win_config = {
        relative = "editor",
        row = math.floor((screen_height - popup_height) / 2) - 1,
        col = math.floor((screen_width - popup_width) / 2),
        width = popup_width,
        height = popup_height,
        style = "minimal",
        border = "rounded",
        zindex = 50,
    }

    -- Open window with buffer
    local win = vim.api.nvim_open_win(buf, true, win_config)

    return win
end

-- Function to handle LSP report generation responses
local function handle_lsp_response(err, result, context)
    if err then
        vim.notify(
            "Error generating report: " .. vim.inspect(err),
            vim.log.levels.ERROR
        )
        return
    end

    if not result then
        vim.notify("No data returned from LSP", vim.log.levels.WARN)
        return
    end

    -- Create buffer with the content
    local buf = create_buffer(result, "markdown")

    -- Create popup window
    create_popup_window(buf)
end

-- Setup pendulum commands for report generation
local function setup_pendulum_commands(lsp_client)
    vim.api.nvim_create_user_command("Pendulum", function(args)
        if not lsp_client or lsp_client.is_stopped() then
            vim.notify(
                "Pendulum LSP client not available",
                vim.log.levels.ERROR
            )
            return
        end

        options.time_range = args.args or "all"
        options.view = "metrics"

        -- Send request to LSP for metrics report
        lsp_client.request("workspace/executeCommand", {
            command = "pendulum.generateMetricsReport",
            arguments = { options },
        }, handle_lsp_response)
    end, { nargs = "?" })

    vim.api.nvim_create_user_command("PendulumHours", function()
        if not lsp_client or lsp_client.is_stopped() then
            vim.notify(
                "Pendulum LSP client not available",
                vim.log.levels.ERROR
            )
            return
        end

        options.view = "hours"

        -- Send request to LSP for hours report
        lsp_client.request("workspace/executeCommand", {
            command = "pendulum.generateHoursReport",
            arguments = { options },
        }, handle_lsp_response)
    end, { nargs = 0 })
end

-- Report generation setup using LSP
function M.setup(opts)
    options = opts

    -- Get LSP client from handlers module
    local handlers = require("pendulum.handlers")
    local lsp_client = handlers.get_lsp_client()

    -- Set up commands if client is available or wait for it
    if lsp_client then
        setup_pendulum_commands(lsp_client)
    else
        -- If client not available yet, wait for it to be initialized
        vim.defer_fn(function()
            local client = handlers.get_lsp_client()
            if client then
                setup_pendulum_commands(client)
            else
                vim.notify(
                    "Unable to get Pendulum LSP client",
                    vim.log.levels.WARN
                )
            end
        end, 1000) -- Wait 1 second for LSP to initialize
    end
end

return M

