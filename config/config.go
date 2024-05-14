package config

import (
	"errors"
	"reflect"
	"strings"

	"github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
)

var (
	ErrUnableToDetermineConfigFileFormat = errors.New("unable to determine config file format")
	ErrTargetMustBeStructPtr             = errors.New("target must be struct ptr")
)

type viperLoader struct {
	viper        *viper.Viper
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
	rv := reflect.ValueOf(target)
	if err = ensureStructPtr(rv); err != nil {
		return err
	}

	if vl.useDefaults {
		defaults.SetDefaults(target)
	}

	if vl.useEnv {
		err = vl.loadFromEnv(extractConfigDefs(rv))
	} else {
		err = vl.loadFromFile()
	}

	if err != nil {
		return err
	}

	if err := vl.viper.Unmarshal(target); err != nil {
		return err
	}

	return nil
}

func (vl *viperLoader) loadFromFile() error {
	if vl.confFile != "" {
		vl.viper.SetConfigFile(vl.confFile)
	} else {
		vl.viper.SetConfigName(vl.confName)
		vl.viper.SetConfigType(vl.confFileType)

		if vl.confFilePath != "" {
			vl.viper.AddConfigPath(vl.confFilePath)
		}

		vl.viper.AddConfigPath(".")
	}

	return vl.viper.ReadInConfig()
}

func (vl *viperLoader) loadFromEnv(cfgs []configDef) error {
	vl.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_", " ", "_"))
	vl.viper.SetEnvPrefix(vl.envPrefix)
	vl.viper.AutomaticEnv()

	for _, cfg := range cfgs {
		vl.viper.SetDefault(cfg.Key, cfg.Default)

		err := vl.viper.BindEnv(cfg.Key)
		if err != nil {
			return err
		}
	}

	return nil
}

func ensureStructPtr(rv reflect.Value) error {
	rvKind := rv.Kind()
	rvIndirectKind := reflect.Indirect(rv).Kind()

	if rvKind != reflect.Ptr || rvIndirectKind != reflect.Struct {
		return ErrTargetMustBeStructPtr
	}

	return nil
}

func convertCamelCaseToSnakeCase(input string) string {
	sb := strings.Builder{}
	for _, ch := range input {
		if ch >= 'A' && ch <= 'Z' {
			sb.WriteByte('_')
		}

		sb.WriteRune(ch)
	}

	return sb.String()
}

func deref(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}
	return rv
}

func extractConfigDefs(rv reflect.Value) []configDef {
	return readRecursive(rv, "")
}

func readRecursive(rv reflect.Value, rootKey string) []configDef {
	rt := rv.Type()

	configs := make([]configDef, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		ft := rt.Field(i)
		fv := deref(rv.Field(i))
		key := convertCamelCaseToSnakeCase(ft.Name)

		if rootKey != "" {
			key = rootKey + "." + key
		}

		if fv.Kind() == reflect.Struct {
			nestedConfigs := readRecursive(fv, key)
			configs = append(configs, nestedConfigs...)
		} else {
			configs = append(configs, configDef{
				Key:     key,
				Doc:     ft.Tag.Get("doc"),
				Default: fv.Interface(),
			})
		}
	}

	return configs
}
