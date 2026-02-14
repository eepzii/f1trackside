package scrutineer

import (
	"encoding/json"
	"reflect"
)

func NewReport(onType string) *Report {
	return &Report{
		Root: &TypeStats{
			Name:     onType,
			Children: make(map[string]*TypeStats),
		},
	}
}

func (r *Report) ReadLine(data []byte, responseTypeTemplate any) error {
	var minimumJSON any
	if err := json.Unmarshal(data, &minimumJSON); err != nil {
		return err
	}

	var responseType = reflect.TypeOf(responseTypeTemplate)
	responseVal := reflect.New(responseType)
	if err := json.Unmarshal(data, responseVal.Interface()); err != nil {
		return err
	}

	r.Root.Analyze(minimumJSON, responseVal.Elem().Interface())

	return nil
}

func (r *Report) Show() string {
	return r.Root.Tree()
}
