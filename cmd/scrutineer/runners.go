package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/eepzii/f1trackside/internal/scrutineer"
)

func runSynchronous(configs []scrutineer.YamlConfig) {
	for _, config := range configs {
		fmt.Printf("checking %s\n", config.TypeName)

		inspector := scrutineer.New(config.TypeName)
		typeTemplate, exists := scrutineer.TypeRegistry[inspector.Root.Name]
		if !exists {
			fmt.Printf("    unknown type %q -> skipping...\n", config.TypeName)
			continue
		}

		for _, path := range config.Paths {
			if err := inspector.InspectFile(path, typeTemplate, false); err != nil {
				log.Printf("    error processing %s: %v\n", path, err)
			}
		}

		inspector.PrintTree()
	}
}

func runConcurrent(configs []scrutineer.YamlConfig) {
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

			typeTemplate, exists := scrutineer.TypeRegistry[inspector.Root.Name]
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

						if err := tmpInspector.InspectFile(path, typeTemplate, true); err != nil {
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
				if inspectedFile.HasSyntaxError {
					inspector.HasSyntaxError = true
				}
			}

			printMu.Lock()
			inspector.PrintTree()
			printMu.Unlock()
		})
	}

	configWg.Wait()
}
