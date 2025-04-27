package configloader

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/joho/godotenv"
	"github.com/mcuadros/go-defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var (
	ErrUnableToDetermineConfigFileFormat = errors.New("unable to determine config file format")
	ErrTargetMustBeStructPtr             = errors.New("target must be struct ptr")
)

type viperLoader struct {
	viper               *viper.Viper
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

func NewDefaultLoader() *viperLoader {
	return &viperLoader{
		viper:               viper.New(),
		confName:            "config",
		confFilePath:        "../",
		confFileType:        "yaml",
		useEnv:              true,
		useDefaults:         true,
		useSnakeCaseEnvVars: true,
		enableFallbacking:   true,
	}
}

type configDef struct {
	Key     string `json:"key"`
	Doc     string `json:"doc"`
	Default any    `json:"default"`
}

func (vl *viperLoader) Load(ctx context.Context, target any) (err error) {
	rv := reflect.ValueOf(target)
	if err = ensureStructPtr(rv); err != nil {
		return err
	}

	if vl.useDefaults {
		defaults.SetDefaults(target)
	}

	if vl.enableFallbacking {
		err = vl.loadWithFallbacking(ctx, rv, target)
	} else {
		err = vl.loadWithoutFallbacking(ctx, rv, target)
	}

	if err != nil {
		return err
	}

	if err := vl.viper.Unmarshal(target, viper.DecodeHook(configDecoder)); err != nil {
		return err
	}

	vl.viper.WatchConfig()
	vl.viper.OnConfigChange(func(event fsnotify.Event) {
		if event.Op == fsnotify.Write {
			if err := vl.viper.Unmarshal(&target); err != nil {
				panic("config watcher: " + err.Error())
			}
		}
	})

	return nil
}

func (vl *viperLoader) loadWithoutFallbacking(ctx context.Context, rv reflect.Value, target any) (err error) {
	useEnv := vl.useEnv

	if strings.HasSuffix(vl.confFile, ".env") || vl.confFileType == "env" {
		confFile := vl.confFile
		if confFile == "" {
			confFile = filepath.Join(vl.confFilePath, vl.confName+"."+vl.confFileType)
		}

		err := godotenv.Load(confFile)
		if err != nil {
			return err
		}

		useEnv = true
	}

	if useEnv {
		err = vl.loadFromEnv(rv)
		if err == nil {
			go watchEnvVars(ctx, vl.viper, target)
		}
	} else {
		err = vl.loadFromFile()
	}

	return err
}

func (vl *viperLoader) loadWithFallbacking(ctx context.Context, rv reflect.Value, target any) (err error) {
	err = vl.loadFromFile()
	if err != nil {
		if strings.HasSuffix(vl.confFile, ".env") || vl.confFileType == "env" {
			confFile := vl.confFile
			if confFile == "" {
				confFile = filepath.Join(vl.confFilePath, vl.confName+"."+vl.confFileType)
			}

			err := godotenv.Load(confFile)
			if err != nil {
				return err
			}
		}

		err = vl.loadFromEnv(rv)
		if err == nil {
			go watchEnvVars(ctx, vl.viper, target)
		}
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

func (vl *viperLoader) loadFromEnv(rv reflect.Value) error {
	if vl.useSnakeCaseEnvVars {
		vl.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_", " ", "_"))
	} else {
		vl.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "", "-", "", " ", ""))
	}

	cfgs := readRecursive(deref(rv), "")

	vl.viper.SetEnvPrefix(vl.envPrefix)
	vl.viper.AutomaticEnv()

	for _, cfg := range cfgs {
		val := cfg.Key
		if vl.useSnakeCaseEnvVars {
			val = convertCamelCaseToSnakeCase(val)
		} else {
			val = strings.ReplaceAll(val, "_", "")
		}

		err := vl.viper.BindEnv(val)
		if err != nil {
			return err
		}
	}

	return nil
}

func readRecursive(rv reflect.Value, root string) []configDef {
	result := []configDef{}
	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		fv := deref(rv.Field(i))
		ft := rt.Field(i)

		name, exists := ft.Tag.Lookup("mapstructure")
		if !exists {
			continue
		}

		if fv.Kind() == reflect.Struct {
			nestedConfigs := readRecursive(fv, name)
			result = append(result, nestedConfigs...)

		} else {
			key := name
			if root != "" {
				key = root + "." + name
			}
			result = append(result, configDef{
				Key:     key,
				Doc:     ft.Tag.Get("doc"),
				Default: fv.Interface(),
			})
		}
	}

	return result
}

func deref(rv reflect.Value) reflect.Value {
	if rv.Kind() == reflect.Ptr {
		rv = reflect.Indirect(rv)
	}

	return rv
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
	for i, ch := range input {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			sb.WriteByte('_')
		}

		sb.WriteRune(ch)
	}

	return strings.ToLower(sb.String())
}

func configDecoder(dc *mapstructure.DecoderConfig) {
	dc.MatchName = func(mapKey, fieldName string) bool {
		equalFold := strings.EqualFold(mapKey, fieldName)
		snakeCaseEnvVars := convertCamelCaseToSnakeCase(mapKey) == convertCamelCaseToSnakeCase(fieldName)
		camelCaseEnvVars := strings.ReplaceAll(mapKey, "_", "") == strings.ReplaceAll(fieldName, "_", "")

		return snakeCaseEnvVars || equalFold || camelCaseEnvVars
	}
}

func watchEnvVars(ctx context.Context, v *viper.Viper, target any) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := v.Unmarshal(target); err != nil {
				panic("config watcher: " + err.Error())
			}
		}

		time.Sleep(10 * time.Second)
	}
}
