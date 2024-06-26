package configloader

import "github.com/spf13/viper"

type configBuilder struct {
	confName            string
	confFile            string
	confFilePath        string
	confFileType        string
	useEnv              bool
	envPrefix           string
	useDefaults         bool
	useSnakeCaseEnvVars bool
	enableFallbacking   bool
}

func NewConfigLoaderBuilder() *configBuilder {
	cb := new(configBuilder)
	cb.useSnakeCaseEnvVars = true

	return cb
}

func (cb *configBuilder) WithName(name string) *configBuilder {
	cb.confName = name

	return cb
}

func (cb *configBuilder) WithFile(file string) *configBuilder {
	cb.confFile = file

	return cb
}

func (cb *configBuilder) WithFileType(fileType string) *configBuilder {
	cb.confFileType = fileType

	return cb
}

func (cb *configBuilder) WithFilePath(filePath string) *configBuilder {
	cb.confFilePath = filePath

	return cb
}

func (cb *configBuilder) UseEnv() *configBuilder {
	cb.useEnv = true

	return cb
}

func (cb *configBuilder) DoNotUseEnv() *configBuilder {
	cb.useEnv = false

	return cb
}

func (cb *configBuilder) WithEnvPrefix(prefix string) *configBuilder {
	cb.envPrefix = prefix

	return cb
}

func (cb *configBuilder) UseDefaults() *configBuilder {
	cb.useDefaults = true

	return cb
}

func (cb *configBuilder) DoNotUseSnakeCaseEnvironmentVariableNamingConvention() *configBuilder {
	cb.useSnakeCaseEnvVars = false

	return cb
}

func (cb *configBuilder) DoNotUseDefaults() *configBuilder {
	cb.useDefaults = false

	return cb
}

func (cb *configBuilder) EnableFallbacking() *configBuilder {
	cb.enableFallbacking = true

	return cb
}

func (cb *configBuilder) Build() *viperLoader {
	return &viperLoader{
		viper:               viper.New(),
		confName:            cb.confName,
		confFile:            cb.confFile,
		confFilePath:        cb.confFilePath,
		confFileType:        cb.confFileType,
		useEnv:              cb.useEnv,
		envPrefix:           cb.envPrefix,
		useDefaults:         cb.useDefaults,
		useSnakeCaseEnvVars: cb.useSnakeCaseEnvVars,
		enableFallbacking:   cb.enableFallbacking,
	}
}
