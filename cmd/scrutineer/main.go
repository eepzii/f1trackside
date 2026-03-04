package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/goccy/go-yaml"

	"github.com/eepzii/f1trackside/internal/scrutineer"
)

func main() {

	configPath := flag.String("config", "", "Path to config")
	useConcurrency := flag.Bool("concurrent", false, "Run the program in concurrent mode (WARNING: cpu and memory intensive)")
	flag.Parse()

	if *configPath == "" {
		log.Fatal("please set path to config file")
	}

	yamlData, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("could not read config: %v\n", err)
	}

	var configs []scrutineer.YamlConfig
	if err := yaml.Unmarshal(yamlData, &configs); err != nil {
		log.Fatalf("unable to unmarshal %s contents: %v\n", *configPath, err)
	}

	fmt.Println("Scrutineer Started")
	fmt.Println("========================")
	fmt.Println()

	if *useConcurrency {
		runConcurrent(configs)
		return
	}

	runSynchronous(configs)
}
