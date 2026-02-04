package types

import (
	"encoding/json"
	"testing"
)

// TODO: refactor test table
func TestParseTimingAppData(t *testing.T) {

	jsonInput := []byte(`{
        "TimingAppData":{
            "Lines":{
                "44":{
                    "RacingNumber":"44",
					"Line":3,
                    "GridPos":"3",
                    "Stints":[{
					    "LapTime":"1:40.526",
						"LapNumber":3,
						"LapFlags":0,
                        "Compound":"UNKNOWN",
						"New":"false",
						"TyresNotChanged":"0",
                        "TotalLaps":3,
						"StartLaps":0
                    }]
                },
                "16":{
                    "RacingNumber":"16",
                    "Stints":[{
                        "Compound":"MEDIUM",
					    "New":"true",
					    "TyresNotChanged":"0",
                        "TotalLaps":22
                    }, {
                        "Compound":"HARD",
					    "TyresNotChanged":"1",
                        "TotalLaps":0,
                        "New":"true"
                    }]
                },
				"22":{
				    "Stints":{
					    "0":{
						    "TotalLaps":1
						}
					}
				},
				"43":{
				    "Stints":{
					    "0":{
						    "Compound":"SOFT"
						}
					}
				},
				"18":{ }
            }
        }
    }`)

	var response TimingAppDataResponse
	if err := json.Unmarshal(jsonInput, &response); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	if response.TimingAppData.Lines == nil {
		t.Fatalf(".Lines is nil")
	}

	var zero = 0
	var one = 1
	var three = 3
	var twentytwo = 22

	tests := []struct {
		DriverNumber         string
		ExpectedRacingNumber string
		ExpectedStintMapSize int
		ExpectedStints       map[string]struct {
			ExpectedTyresNotChanged string
			ExpectedCompound        string
			ExpectedNew             string
			ExpectedTotalLaps       *int
		}
	}{
		{"44", "44", 1, map[string]struct {
			ExpectedTyresNotChanged string
			ExpectedCompound        string
			ExpectedNew             string
			ExpectedTotalLaps       *int
		}{
			"0": {"0", "UNKNOWN", "false", &three},
		}},
		{"16", "16", 2, map[string]struct {
			ExpectedTyresNotChanged string
			ExpectedCompound        string
			ExpectedNew             string
			ExpectedTotalLaps       *int
		}{
			"0": {"0", "MEDIUM", "true", &twentytwo},
			"1": {"1", "HARD", "true", &zero},
		}},
		{"22", "", 1, map[string]struct {
			ExpectedTyresNotChanged string
			ExpectedCompound        string
			ExpectedNew             string
			ExpectedTotalLaps       *int
		}{
			"0": {"", "", "", &one},
		}},
		{"43", "", 1, map[string]struct {
			ExpectedTyresNotChanged string
			ExpectedCompound        string
			ExpectedNew             string
			ExpectedTotalLaps       *int
		}{
			"0": {"", "SOFT", "", nil},
		}},
		{"18", "", 0, nil},
	}

	for _, test := range tests {

		driver, exists := response.TimingAppData.Lines[test.DriverNumber]
		if !exists {
			t.Errorf("missing key on .Lines: %s", test.DriverNumber)
			continue
		}

		err := validateValues(driver.RacingNumber, test.ExpectedRacingNumber)
		if err != nil {
			t.Errorf(".Lines[%s].RacingNumber -> %v", test.DriverNumber, err)
		}

		err = validateValues(len(driver.Stints), test.ExpectedStintMapSize)
		if err != nil {
			t.Errorf(".Lines[%s].Stints map size -> %v", test.DriverNumber, err)
		}

		for key, stint := range driver.Stints {

			expectedStint, ok := test.ExpectedStints[key]
			if !ok {
				t.Errorf("missing key on .Lines[%s].Stints: %s", test.DriverNumber, key)
				continue
			}

			err = validateValues(stint.TyresNotChanged, expectedStint.ExpectedTyresNotChanged)
			if err != nil {
				t.Errorf(".Lines[%s].Stints[%s].TyresNotChanged -> %v",
					test.DriverNumber, key, err)
			}

			err = validateValues(stint.Compound, expectedStint.ExpectedCompound)
			if err != nil {
				t.Errorf(".Lines[%s].Stints[%s].Compound -> %v",
					test.DriverNumber, key, err)
			}

			err = validateValues(stint.New, expectedStint.ExpectedNew)
			if err != nil {
				t.Errorf(".Lines[%s].Stints[%s].New -> %v",
					test.DriverNumber, key, err)
			}

			err = validatePointers(stint.TotalLaps, expectedStint.ExpectedTotalLaps)
			if err != nil {
				t.Errorf(".Lines[%s].Stints[%s].TotalLaps -> %v",
					test.DriverNumber, key, err)
			}

		}
	}

}
