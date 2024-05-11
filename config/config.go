package config

import (
	"github.com/spf13/viper"
)

type viperLoader struct {
	viper        *viper.Viper
	target       interface{}
	confName     string
	confFile     string
	confFilePath string
	confFileType string
	useEnv       bool
	envPrefix    string
	useDefaults  bool
}

type configDef struct {
	Key     string      `json:"key"`
	Doc     string      `json:"doc"`
	Default interface{} `json:"default"`
}

func (vl *viperLoader) Load(target interface{}) (err error) {
	return nil
}
