package main

import (
	"os"

	"github.com/ranefattesingh/pkg/config"
)

func main() {
	os.Setenv("TEST", "1")

	type Conf struct {
		Test string `yaml:"test"`
	}

	loader := config.NewConfigLoaderBuilder().UseEnv().Build()

	t := Conf{}
	loader.Load(&t)
}
