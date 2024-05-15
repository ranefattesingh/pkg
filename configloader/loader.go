package configloader

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"

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
	stopWatching        chan struct{}
	startWatching       chan struct{}
	target              interface{}
}

type configDef struct {
	Key     string      `json:"key"`
	Doc     string      `json:"doc"`
	Default interface{} `json:"default"`
}

func (vl *viperLoader) Load(target any) (err error) {
	rv := reflect.ValueOf(target)
	if err = ensureStructPtr(rv); err != nil {
		return err
	}

	if vl.useDefaults {
		defaults.SetDefaults(target)
	}

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
	} else {
		err = vl.loadFromFile()
	}

	if err != nil {
		return err
	}

	if err := vl.viper.Unmarshal(target, configDecoder); err != nil {
		return err
	}

	vl.target = target

	return nil
}

func (vl *viperLoader) EnableLiveReload(ctx context.Context) {
	defer close(vl.startWatching)

	go func() {
		for {
			select {
			case <-ctx.Done():
				close(vl.stopWatching)

				return
			case <-vl.stopWatching:
				return

			case <-vl.startWatching:
				vl.viper.OnConfigChange(func(e fsnotify.Event) {
					vl.Load(vl.target)
				})

				vl.viper.WatchConfig()
			}
		}
	}()
}

func (vl *viperLoader) StopLiveReload() {
	vl.stopWatching <- struct{}{}
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

	// for _, cfg := range cfgs {
	// 	vl.viper.SetDefault(cfg.Key, cfg.Default)
	// }

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
