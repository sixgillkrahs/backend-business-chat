package config

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type JWT struct {
	SecretKey string `mapstructure:"secretKey"`
	Duration  int64  `mapstructure:"duration"`
}

type ServerConfig struct {
	Port int        `mapstructure:"port" validate:"required,min=1,max=65535"`
	Mode string     `mapstructure:"mode" validate:"required,oneof=debug release test"`
	Env  string     `mapstructure:"env" validate:"required,oneof=develop production test"`
	Cors CorsConfig `mapstructure:"cors"`
	Jwt  JWT        `mapstructure:"jwt"`
}

type CorsConfig struct {
	AllowOrigins string `mapstructure:"allowOrigins" validate:"required"`
}

type PostgresConfig struct {
	URI      string `mapstructure:"uri" validate:"required"`
	Database string `mapstructure:"database" validate:"required"`
}

type RedisConfig struct {
	Addr          string   `mapstructure:"addr" validate:"required"`
	Password      string   `mapstructure:"password"`
	DB            int      `mapstructure:"db" validate:"min=0"`
	UseSentinel   bool     `mapstructure:"useSentinel"`
	MasterName    string   `mapstructure:"masterName"`
	SentinelAddrs []string `mapstructure:"sentinelAddrs"`
}
