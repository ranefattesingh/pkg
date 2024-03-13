package config

type MainConfig struct {
	MyConfigVar string `yaml:"my_config_var" mapstructure:"my_config_var"`
}
