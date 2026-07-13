package config

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
}

type ServerConfig struct {
	Port int        `mapstructure:"port" validate:"required,min=1,max=65535"`
	Mode string     `mapstructure:"mode" validate:"required,oneof=debug release test"`
	Env  string     `mapstructure:"env" validate:"required,oneof=develop production test"`
	Cors CorsConfig `mapstructure:"cors"`
}

type CorsConfig struct {
	AllowOrigins string `mapstructure:"allowOrigins" validate:"required"`
}

type PostgresConfig struct {
	URI      string `mapstructure:"uri" validate:"required"`
	Database string `mapstructure:"database" validate:"required"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr" validate:"required"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db" validate:"min=0"`
}
