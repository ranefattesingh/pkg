package config

import (
	"os"
	"strings"

	"github.com/MakeNowJust/heredoc/v2"
	"github.com/jeremywohl/flatten"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type loader struct {
	fileName          string
	path              string
	envFileName       string
	envPath           string
	enableFallback    bool
	prioritizeEnvFile bool
	envPrefix         string
}

func (l loader) load(conf interface{}) error {
	v, err := l.loadFromFile()
	if err != nil {
		if !l.enableFallback {
			return err
		}

		if l.prioritizeEnvFile {
			v, err = l.loadFromEnvFile()
			if err != nil {
				v, err = l.loadFromEnv(conf)
				if err != nil {
					return err
				}
			}
		} else {
			v, err = l.loadFromEnv(conf)
			if err != nil {
				return err
			}
		}
	}

	err = v.Unmarshal(conf)
	if err != nil {
		return err
	}

	return nil
}

func (l loader) loadFromFile() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(l.fileName)
	v.AddConfigPath(l.path)
	v.SetConfigType("yaml")

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (l loader) loadFromEnvFile() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName(l.envFileName)
	v.SetConfigType("env")
	v.AddConfigPath(l.envPath)

	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (l loader) loadFromEnv(conf interface{}) (*viper.Viper, error) {
	v := viper.New()
	v.SetEnvPrefix(l.envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	keys, err := getFlattenedStructKeys(conf)
	if err != nil {
		return nil, err
	}

	for key := range keys {
		err := v.BindEnv(strings.ToUpper(keys[key]))
		if err != nil {
			return nil, err
		}
	}

	return v, nil
}

func getFlattenedStructKeys(conf interface{}) ([]string, error) {
	var structMap map[string]interface{}

	err := mapstructure.Decode(conf, &structMap)
	if err != nil {
		return nil, err
	}

	flat, err := flatten.Flatten(structMap, "", flatten.DotStyle)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}

	return keys, nil
}

func DefaultLoader() loader {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	loader := loader{
		fileName:          "config.yaml",
		path:              path,
		envPrefix:         "",
		enableFallback:    true,
		envFileName:       ".env",
		envPath:           path,
		prioritizeEnvFile: true,
	}

	return loader
}

func (l loader) NewCommand(config interface{}) *cobra.Command {
	var filePath string
	var configType string
	var envPrefix string

	command := &cobra.Command{
		Use: "app_name <command> [flags]",
		Long: heredoc.Doc(`
			app description will come here
		`),
		Example: heredoc.Doc(`
			$ ./app --path path/to/config/config.yaml
			$ ./app -p path/to/config/config.yaml
			$ ./app --type env
			$ ./app -t env
		`),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileName := ""
			if filePath != "" {
				dirs := strings.Split(filePath, "/")
				filePath = strings.Join(dirs[:len(dirs)-1], "/")
				fileName = dirs[len(dirs)-1]
			}

			if configType == "yaml" {
				l.enableFallback = false
				l.prioritizeEnvFile = false
				l.fileName = fileName
				l.path = filePath
				l.envFileName = ""
				l.envPath = ""
			}

			if configType == "env" {
				l.enableFallback = true
				l.prioritizeEnvFile = false
				l.fileName = ""
				l.path = ""
				l.envFileName = ""
				l.envPath = ""
			}

			if configType == ".env" {
				l.enableFallback = true
				l.prioritizeEnvFile = true
				l.fileName = ""
				l.path = ""
				l.envFileName = ".env"
				l.envPath = filePath
			}

			l.envPrefix = envPrefix

			return l.load(config)
		},
	}

	command.Flags().StringVarP(&filePath, "config", "c", "", "override default config file")
	command.Flags().StringVarP(&configType, "type", "t", "", "override default config type")
	command.Flags().StringVarP(&envPrefix, "prefix", "p", "", "override default config type")

	return command
}
