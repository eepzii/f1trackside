package types

import "fmt"

func validateValues[T comparable](got, want T) error {
	if got != want {
		return fmt.Errorf("expected: %v, got: %v", want, got)
	}
	return nil
}

func validatePointers[T comparable](got, want *T) error {

	if got == nil && want == nil {
		return nil
	}

	if (got != nil && want != nil) && *got == *want {
		return nil
	}

	var wantStr = "nil"
	if want != nil {
		wantStr = fmt.Sprintf("%v", *want)
	}

	var gotStr = "nil"
	if got != nil {
		gotStr = fmt.Sprintf("%v", *got)
	}

	return fmt.Errorf("expected: %s, got: %s", wantStr, gotStr)

}
