package config

// Ovoo API configuration

type APIConfig struct {
	Cache        *ConfigCache          `koanf:"cache"`
	Database     ConfigDB              `koanf:"database"`
	DefaultAdmin *ConfigDefaultAdmin   `koanf:"default_admin"`
	Domains      []string              `koanf:"domains"`
	ListenAddr   string                `koanf:"listen_addr"`
	Log          ConfigLogging         `koanf:"logging"`
	OIDC         map[string]ConfigOIDC `koanf:"oidc"`
	TLS          ConfigTLS             `koanf:"tls"`
	Version      SystemVersion
}

type SystemVersion struct {
	Version   string
	GitCommit string
	BuiltAt   string
}

type ConfigLogging struct {
	Destination string `koanf:"destination"`
	Level       string `koanf:"level"`
}

type ConfigTLS struct {
	Cert string `koanf:"cert"`
	Key  string `koanf:"key"`
}

type ConfigDefaultAdmin struct {
	FirstName string `koanf:"first_name"`
	LastName  string `koanf:"last_name"`
	Login     string `koanf:"login"`
	Password  string `koanf:"password"`
}

type ConfigOIDC struct {
	ClientId       string            `koanf:"client_id"`
	ClientSecret   string            `koanf:"client_secret"`
	Issuer         string            `koanf:"issuer"`
	ExtraScopes    []string          `koanf:"extra_scopes"`     // extra scopes to include in request
	ExtraURLParams map[string]string `koanf:"extra_url_params"` // extra parameters to include in authorization URL
}

type ConfigCache struct {
	CacheDriver   string            `koanf:"driver"`
	Config        ConfigCacheDriver `koanf:"config"`
	ListTTL       int               `koanf:"list_ttl"`
	SingleItemTTL int               `koanf:"single_item_ttl"`
}

type ConfigCacheDriver struct {
	Redis *ConfigCacheDriverRedis `koanf:"redis"`
}

type ConfigCacheDriverRedis struct {
	Addr     *string `koanf:"address"`
	DB       int     `koanf:"db"`
	Password *string `koanf:"password"`
	Protocol int     `koanf:"protocol"`
	Username *string `koanf:"username"`
}

type ConfigDB struct {
	Config   ConfigDBDriver `koanf:"config"`
	Driver   string         `koanf:"driver"`
	LogLevel string         `koanf:"log_level"`
}

type ConfigDBDriver struct {
	GORM ConfigDBDriverGORM `koanf:"gorm"`
}

type ConfigDBDriverGORM struct {
	ConnectionString string `koanf:"connection_string"`
	Driver           string `koanf:"driver"`
}

// Ovoo Milter configuration

type MilterConfig struct {
	Api             ConfigMilterAPIConn `koanf:"api"`
	Domains         []string            `koanf:"domains"`
	ListenAddr      string              `koanf:"listen_addr"`
	Log             ConfigLogging       `koanf:"log"`
	MailDisplayName string              `koanf:"mail_display_name"`
}

type ConfigMilterAPIConn struct {
	Addr          string `koanf:"addr"`
	AuthToken     string `koanf:"auth_token"`
	TLSSkipVerify bool   `koanf:"tls_skip_verify"`
	Timeout       int    `koanf:"client_timeout"`
}
