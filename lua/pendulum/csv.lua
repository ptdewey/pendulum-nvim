-- csv.lua
local M = {}

---@param field any
---@return string
local function escape_csv_field(field)
    if type(field) == "string" and (field:find('[,"]') or field:find("\n")) then
        field = '"' .. field:gsub('"', '""') .. '"'
    end
    return tostring(field)
end

---convert lua table to csv style table
---@param t table
---@return string, table
local function table_to_csv(t)
    if #t == 0 then
        return "", {}
    end

    local csv_data = {}
    local headers = {}

    for key, _ in pairs(t[1]) do
        table.insert(headers, key)
    end
    table.sort(headers)

    for _, row in ipairs(t) do
        local temp = {}
        for _, field_key in ipairs(headers) do
            table.insert(temp, escape_csv_field(row[field_key]))
        end
        table.insert(csv_data, table.concat(temp, ",") .. "\n")
    end

    return table.concat(csv_data), headers
end

---write lua table to csv file
---@param filepath string
---@param data_table table
---@param include_header boolean
function M.write_table_to_csv(filepath, data_table, include_header)
    local f = io.open(filepath, "a+")
    if not f then
        error("Error opening file: " .. filepath)
    end

    local csv_content, headers = table_to_csv(data_table)
    if f:seek("end") == 0 and include_header then
        f:write(table.concat(headers, ",") .. "\n")
    end

    if csv_content ~= "" then
        f:write(csv_content)
    else
        print("No data to write.")
    end

    f:close()
end
return M
