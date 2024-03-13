package main

import (
	"config/config"
	"fmt"
)

func main() {
	var configuration config.MainConfig
	err := config.DefaultLoader().NewCommand(&configuration).Execute()
	if err != nil {
		panic(err)
	}

	fmt.Println(configuration)
}
