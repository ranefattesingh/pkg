package config_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/ranefattesingh/pkg/config"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	LogLevel       string         `yaml:"log_level" mapstructure:"log_level" default:"info"`
	ServerConfig   HTTPConfig     `yaml:"http" mapstructure:"http"`
	DatabaseConfig DatabaseConfig `yaml:"db" mapstructure:"db"`
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

	expectedConfig := AppConfig{
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
	}

	file, err := createTestFile(expectedConfig, "", "config*.yaml")
	if err != nil {
		errorF(t, t.Name(), "err", nil, err)

		return
	}

	actualConfig := AppConfig{}
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

	err = loader.Load(&actualConfig)
	if err != nil {
		errorF(t, t.Name(), "err", nil, err)

		return
	}

	defer os.Remove(file.Name())

	isEqual := reflect.DeepEqual(expectedConfig, actualConfig)
	if !isEqual {
		errorF(t, t.Name(), "config", expectedConfig, actualConfig)
	}
}

func testLoadFromFileWithConfigFile(t *testing.T) {
	t.Helper()

	expectedConfig := AppConfig{
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
	}

	file, err := createTestFile(expectedConfig, "", "config*.yaml")
	if err != nil {
		errorF(t, t.Name(), "err", nil, err)

		return
	}

	actualConfig := AppConfig{}

	loader := config.NewConfigLoaderBuilder().
		WithFile(file.Name()).
		Build()

	err = loader.Load(&actualConfig)
	if err != nil {
		errorF(t, t.Name(), "err", nil, err)

		return
	}

	defer os.Remove(file.Name())

	isEqual := reflect.DeepEqual(expectedConfig, actualConfig)
	if !isEqual {
		errorF(t, t.Name(), "config", expectedConfig, actualConfig)
	}
}

func testLoadFromFileWithConfigFileAndDefaultsEnabled(t *testing.T) {
	t.Helper()

	expectedConfig := AppConfig{
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
	}

	type PartialDatabaseConfig struct {
		User     string
		Password string
	}

	partialAppConfig := struct {
		DatabaseConfig PartialDatabaseConfig `yaml:"db"`
	}{
		DatabaseConfig: PartialDatabaseConfig{
			User:     "test",
			Password: "password",
		},
	}

	file, err := createTestFile(partialAppConfig, "", "config*.yaml")
	if err != nil {
		errorF(t, t.Name(), "err", nil, err)

		return
	}

	actualConfig := AppConfig{}

	loader := config.NewConfigLoaderBuilder().
		WithFile(file.Name()).
		UseDefaults().
		Build()

	err = loader.Load(&actualConfig)
	if err != nil {
		errorF(t, t.Name(), "err", nil, err)

		return
	}

	defer os.Remove(file.Name())

	isEqual := reflect.DeepEqual(expectedConfig, actualConfig)
	if !isEqual {
		errorF(t, t.Name(), "config", expectedConfig, actualConfig)
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
