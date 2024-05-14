package config_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ranefattesingh/pkg/config"
	"gopkg.in/yaml.v3"
)

type AppConfigSnakeCase struct {
	LogLevel       string         `yaml:"log_level" mapstructure:"log_level" default:"info"`
	ServerConfig   HTTPConfig     `yaml:"server_config" mapstructure:"server_config"`
	DatabaseConfig DatabaseConfig `yaml:"db_config" mapstructure:"db_config"`
}

type AppConfigCamelCase struct {
	LogLevel       string         `yaml:"logLevel" mapstructure:"logLevel" default:"info"`
	ServerConfig   HTTPConfig     `yaml:"serverConfig" mapstructure:"serverConfig"`
	DatabaseConfig DatabaseConfig `yaml:"dbConfig" mapstructure:"dbConfig"`
}

type HTTPConfig struct {
	Host string `yaml:"host" mapstructure:"host" default:"0.0.0.0"`
	Port int    `yaml:"port" mapstructure:"port" default:"8080"`
}

type DatabaseConfig struct {
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	Host     string `yaml:"host" mapstructure:"host" default:"0.0.0.0"`
	Port     int    `yaml:"port" mapstructure:"port" default:"5432"`
}

func TestLoad(t *testing.T) {
	t.Parallel()

	testTable := map[string]func(t *testing.T){
		"test load from file with config name and type":             testLoadFromFileWithConfigNameAndType,
		"test load from file with config file":                      testLoadFromFileWithConfigFile,
		"test load from file with config file and defaults enabled": testLoadFromFileWithConfigFileAndDefaultsEnabled,
	}

	for name, function := range testTable {
		t.Run(name, func(t *testing.T) {
			f := function

			f(t)
		})
	}
}

func testLoadFromFileWithConfigNameAndType(t *testing.T) {
	t.Helper()

	testTable := map[string]interface{}{
		"should load config defined in snake case": AppConfigSnakeCase{
			LogLevel: "warn",
			ServerConfig: HTTPConfig{
				Host: "localhost",
				Port: 8000,
			},
			DatabaseConfig: DatabaseConfig{
				User:     "test",
				Password: "password",
				Host:     "localhost",
				Port:     3000,
			},
		},
		"should load config defined in camel case": AppConfigCamelCase{
			LogLevel: "warn",
			ServerConfig: HTTPConfig{
				Host: "localhost",
				Port: 8000,
			},
			DatabaseConfig: DatabaseConfig{
				User:     "test",
				Password: "password",
				Host:     "localhost",
				Port:     3000,
			},
		},
	}

	for scenario, testdata := range testTable {
		file, err := createTestFile(testdata, "", "config*.yaml")
		if err != nil {
			errorF(t, t.Name(), "err", nil, err)

			return
		}

		defer os.Remove(file.Name())

		pathIndex := strings.LastIndex(file.Name(), "/")
		extensionIndex := strings.LastIndex(file.Name(), ".")
		fullName := file.Name()
		configName := fullName[pathIndex+1 : extensionIndex]
		configPath := fullName[:pathIndex]
		configType := fullName[extensionIndex+1:]

		loader := config.NewConfigLoaderBuilder().
			WithName(configName).
			WithFilePath(configPath).
			WithFileType(configType).
			Build()

		var result interface{}

		if strings.Contains(scenario, "snake") {
			actualConfig := AppConfigSnakeCase{}
			err = loader.Load(&actualConfig)
			if err != nil {
				errorF(t, t.Name(), "err", nil, err)

				return
			}

			result = actualConfig
		} else {
			actualConfig := AppConfigCamelCase{}
			err = loader.Load(&actualConfig)
			if err != nil {
				errorF(t, t.Name(), "err", nil, err)

				return
			}

			result = actualConfig

		}

		isEqual := reflect.DeepEqual(testdata, result)
		if !isEqual {
			errorF(t, t.Name()+"/"+scenario, "config", testdata, result)
		}
	}
}

func testLoadFromFileWithConfigFile(t *testing.T) {
	t.Helper()

	testTable := map[string]interface{}{
		"should load config defined in snake case": AppConfigSnakeCase{
			LogLevel: "warn",
			ServerConfig: HTTPConfig{
				Host: "localhost",
				Port: 8000,
			},
			DatabaseConfig: DatabaseConfig{
				User:     "test",
				Password: "password",
				Host:     "localhost",
				Port:     3000,
			},
		},
		"should load config defined in camel case": AppConfigCamelCase{
			LogLevel: "warn",
			ServerConfig: HTTPConfig{
				Host: "localhost",
				Port: 8000,
			},
			DatabaseConfig: DatabaseConfig{
				User:     "test",
				Password: "password",
				Host:     "localhost",
				Port:     3000,
			},
		},
	}

	for scenario, testdata := range testTable {
		file, err := createTestFile(testdata, "", "config*.yaml")
		if err != nil {
			errorF(t, t.Name()+"/"+scenario, "err", nil, err)

			return
		}

		loader := config.NewConfigLoaderBuilder().
			WithFile(file.Name()).
			Build()

		var result interface{}

		if strings.Contains(scenario, "snake") {
			actualConfig := AppConfigSnakeCase{}

			err = loader.Load(&actualConfig)
			if err != nil {
				errorF(t, t.Name()+"/"+scenario, "err", nil, err)

				return
			}

			result = actualConfig
		} else {
			actualConfig := AppConfigCamelCase{}

			err = loader.Load(&actualConfig)
			if err != nil {
				errorF(t, t.Name()+"/"+scenario, "err", nil, err)

				return
			}

			result = actualConfig

		}

		defer os.Remove(file.Name())

		isEqual := reflect.DeepEqual(testdata, result)
		if !isEqual {
			errorF(t, t.Name()+"/"+scenario, "config", testdata, result)
		}
	}
}

func testLoadFromFileWithConfigFileAndDefaultsEnabled(t *testing.T) {
	t.Helper()

	type PartialDatabaseConfig struct {
		User     string `yaml:"user" mapstructure:"user"`
		Password string `yaml:"password" mapstructure:"password"`
	}

	type PartialAppSnakeCaseConfig struct {
		DatabaseConfig PartialDatabaseConfig `yaml:"db_config" mapstructure:"db_config"`
	}

	type PartialAppCamelCaseConfig struct {
		DatabaseConfig PartialDatabaseConfig `yaml:"dbConfig" mapstructure:"dbConfig"`
	}

	testTable := map[string]struct {
		expected interface{}
		input    interface{}
	}{
		"should load config defined in snake case": {
			expected: AppConfigSnakeCase{
				LogLevel: "info",
				ServerConfig: HTTPConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				DatabaseConfig: DatabaseConfig{
					User:     "test",
					Password: "password",
					Host:     "0.0.0.0",
					Port:     5432,
				},
			},
			input: PartialAppSnakeCaseConfig{
				DatabaseConfig: PartialDatabaseConfig{
					User:     "test",
					Password: "password",
				},
			},
		},
		"should load config defined in camel case": {
			expected: AppConfigCamelCase{
				LogLevel: "info",
				ServerConfig: HTTPConfig{
					Host: "0.0.0.0",
					Port: 8080,
				},
				DatabaseConfig: DatabaseConfig{
					User:     "test",
					Password: "password",
					Host:     "0.0.0.0",
					Port:     5432,
				},
			},
			input: PartialAppCamelCaseConfig{
				DatabaseConfig: PartialDatabaseConfig{
					User:     "test",
					Password: "password",
				},
			},
		},
	}

	for scenario, testdata := range testTable {
		file, err := createTestFile(testdata.input, "", "config*.yaml")
		if err != nil {
			errorF(t, t.Name(), "err", nil, err)

			return
		}

		loader := config.NewConfigLoaderBuilder().
			WithFile(file.Name()).
			UseDefaults().
			Build()

		var result interface{}

		if strings.Contains(scenario, "snake") {
			actualConfig := AppConfigSnakeCase{}

			err = loader.Load(&actualConfig)
			if err != nil {
				errorF(t, t.Name()+"/"+scenario, "err", nil, err)

				return
			}

			result = actualConfig
		} else {
			actualConfig := AppConfigCamelCase{}

			err = loader.Load(&actualConfig)
			if err != nil {
				errorF(t, t.Name()+"/"+scenario, "err", nil, err)

				return
			}

			result = actualConfig

		}

		defer os.Remove(file.Name())

		isEqual := reflect.DeepEqual(testdata.expected, result)
		if !isEqual {
			errorF(t, t.Name()+"/"+scenario, "config", testdata.expected, result)
		}
	}
}

func errorF(t *testing.T, sceanrio string, param, expectedVal, actualVal interface{}) {
	t.Errorf("[scenario: %v] expected [%v: %v] but actual [%v: %v].", sceanrio, param, expectedVal, param, actualVal)
}

func createTestFile(input interface{}, filePath, fileName string) (*os.File, error) {
	data, err := yaml.Marshal(input)
	if err != nil {
		return nil, err
	}

	file, err := os.CreateTemp(filePath, fileName)
	if err != nil {
		return nil, err
	}

	_, err = file.Write(data)
	if err != nil {
		defer os.Remove(file.Name())

		return nil, err
	}

	return file, nil
}
