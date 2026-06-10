-- Set default values from environment
local default_domain = os.getenv("OVOO_OPENDKIM_DEFAULT_DOMAIN") or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_DOMAIN"
local default_key = os.getenv("OVOO_OPENDKIM_DEFAULT_KEY") or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_KEY"
local default_selector = os.getenv("OVOO_OPENDKIM_DEFAULT_SELECTOR") or
    "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_SELECTOR"

local domain = default_domain
local sign_key = default_key
local selector = default_selector

-- OpenDKIM requests key name via the the global 'query' variable
if query ~= nil then
    domain = query
end

return domain, selector, sign_key
