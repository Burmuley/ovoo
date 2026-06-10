-- Set default values from environment
local default_domain_name = os.getenv("OVOO_OPENDKIM_DEFAULT_DOMAIN") or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_DOMAIN"
local domain = default_domain_name

-- OpenDKIM requests sender via the the global 'query' variable
if query ~= nil then
    domain = query:match("@(.+)$") or query
end

return domain
