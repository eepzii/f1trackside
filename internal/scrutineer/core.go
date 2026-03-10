package scrutineer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/eepzii/f1trackside/internal/scrutineer/schema"
)

func New(onType string) *Scrutineer {
	return &Scrutineer{
		Root: &schema.Field{
			Name:     onType,
			Errors:   make(map[string]int),
			Children: make(map[string]*schema.Field),
		},
	}
}

func (s *Scrutineer) PrintTree() {
	var sb strings.Builder

	fmt.Fprintf(&sb, "%s\n", s.Root.Name)

	s.Root.WriteTree(&sb, "")

	fmt.Println(sb.String())
}

func (s *Scrutineer) InspectFile(path string, template any, useConcurrency bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	workers := 1
	queue := make(chan []byte, 100)
	results := make(chan *schema.Field, workers)

	if useConcurrency {
		workers = runtime.NumCPU()
	}

	for range workers {
		go func() {
			worker := New(s.Root.Name)
			for data := range queue {
				worker.readLine(data, template)
			}
			results <- worker.Root
		}()
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, cap(buf))

	for scanner.Scan() {
		line := scanner.Bytes()

		startIndex := -1
		for i, b := range line {
			if b == '{' {
				startIndex = i
				break
			}
		}
		if startIndex == -1 {
			continue
		}

		// make copy of []byte otherwise the workers will overwrite each other's data
		// due to the nature of scanner.Bytes() which reuses the same memory array for every line
		data := make([]byte, len(line)-startIndex)
		copy(data, line[startIndex:])

		queue <- data
	}

	close(queue)

	for range workers {
		fieldRoot := <-results
		s.Root.Merge(fieldRoot)
	}

	return scanner.Err()
}

func (s *Scrutineer) readLine(data []byte, responseTypeTemplate any) {
	var minimumJSON any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	if err := decoder.Decode(&minimumJSON); err != nil {
		// return on syntax error relating to the JSON input
		return
	}

	responseType := reflect.TypeOf(responseTypeTemplate)
	responseVal := reflect.New(responseType)

	if err := json.Unmarshal(data, responseVal.Interface()); err != nil {
		var typeErr *json.UnmarshalTypeError
		if errors.As(err, &typeErr) {
			path := strings.Split(typeErr.Field, ".")
			node := s.Root.AddByPath(path)

			var numericTypeHint string
			if typeErr.Value == "number" {
				numericTypeHint = " (try \"int\" first)"
			}

			msg := fmt.Sprintf("want: %s%s, got: %s",
				typeErr.Value, numericTypeHint, typeErr.Type)

			if node.Errors == nil {
				node.Errors = make(map[string]int)
			}
			node.Errors[msg]++
		}
	}

	s.Root.Compare(minimumJSON, responseVal.Elem().Interface())
}
