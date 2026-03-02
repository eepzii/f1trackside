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

	for _, config := range configs {
		fmt.Printf("check: %s (on type: %s)\n", config.Name, config.TypeName)

		inspector := scrutineer.New(config.TypeName)

		typeTemplate, exists := scrutineer.TYPE_REGISTRY[inspector.Root.Name]
		if !exists {
			fmt.Printf("    unknown type %q -> skipping...\n", config.TypeName)
			continue
		}

		for _, path := range config.Paths {
			if err := inspector.InspectFile(path, typeTemplate); err != nil {
				log.Printf("    error processing %s: %v\n", path, err)
			}
		}

		inspector.PrintTree()
	}

}
