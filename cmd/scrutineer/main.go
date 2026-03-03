package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"

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

	var configWg sync.WaitGroup
	var printMu sync.Mutex

	for i, config := range configs {
		fmt.Printf("checking %s\n", config.TypeName)
		if i == len(configs)-1 {
			fmt.Println()
			fmt.Println("========================")
			fmt.Println()
		}

		configWg.Go(func() {
			inspector := scrutineer.New(config.TypeName)
			results := make(chan *scrutineer.Scrutineer, len(config.Paths))

			typeTemplate, exists := scrutineer.TYPE_REGISTRY[inspector.Root.Name]
			if !exists {
				printMu.Lock()
				fmt.Printf("    unknown type %q -> skipping...\n", config.TypeName)
				printMu.Unlock()
				return
			}

			go func() {
				defer close(results)
				var fileWg sync.WaitGroup

				for _, path := range config.Paths {
					fileWg.Go(func() {
						tmpInspector := scrutineer.New(config.TypeName)

						if err := tmpInspector.InspectFile(path, typeTemplate); err != nil {
							printMu.Lock()
							fmt.Printf("    error processing %s: %v\n", path, err)
							printMu.Unlock()
						}

						results <- tmpInspector
					})
				}

				fileWg.Wait()
			}()

			for inspectedFile := range results {
				inspector.Root.Merge(inspectedFile.Root)
			}

			printMu.Lock()
			inspector.PrintTree()
			printMu.Unlock()
		})
	}

	configWg.Wait()
}
