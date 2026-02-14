package scrutineer

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
)

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Pointer, reflect.Interface:
		return v.IsNil()
	default:
		return v.IsZero()
	}
}

func StreamFile(path string, report *Report, template any) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

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

		if err := report.ReadLine(line[startIndex:], template); err != nil {
			fmt.Printf("found bad line in %s: %v", path, err)
		}
	}

	return scanner.Err()
}
