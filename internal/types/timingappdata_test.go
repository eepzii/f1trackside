package types

import (
	"encoding/json"
	"testing"
)

// TODO: apply refactor changes of TestParseTimingData here as well.
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
			"1": {"", "", "", nil},
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
			t.Errorf("missing %s driver from map", test.DriverNumber)
			continue
		}

		if driver.RacingNumber != test.ExpectedRacingNumber {
			t.Errorf(`expected .Lines.RacingNumber "%s", got: "%s"`,
				test.ExpectedRacingNumber, driver.RacingNumber)
		}

		if len(driver.Stints) != test.ExpectedStintMapSize {
			t.Errorf(`expected .Lines.Stints length "%d", got: "%d"`,
				test.ExpectedStintMapSize, len(driver.Stints))
		}

		for key, stint := range driver.Stints {

			expectedStint, ok := test.ExpectedStints[key]
			if !ok {
				t.Errorf(`unexpected stint map key "%s" on driver number "%s"`, key, test.DriverNumber)
				continue
			}

			if stint.TyresNotChanged != expectedStint.ExpectedTyresNotChanged {
				t.Errorf(`expected .Lines.Stints.TyresNotChanged "%s", got: "%s"`,
					expectedStint.ExpectedTyresNotChanged, stint.TyresNotChanged)
			}

			if stint.Compound != expectedStint.ExpectedCompound {
				t.Errorf(`expected .Lines.Stints.Compound "%s", got: "%s"`,
					expectedStint.ExpectedCompound, stint.Compound)
			}

			if stint.New != expectedStint.ExpectedNew {
				t.Errorf(`expected .Lines.Stints.New "%s", got: "%s"`,
					expectedStint.ExpectedNew, stint.New)
			}

			if stint.TotalLaps == nil && expectedStint.ExpectedTotalLaps == nil {
				// ok
			} else if stint.TotalLaps == nil && expectedStint.ExpectedTotalLaps != nil {
				t.Errorf(`expected .Lines.Stints.TotalLaps "%d", got: "%v"`,
					*expectedStint.ExpectedTotalLaps, stint.TotalLaps)
			} else if stint.TotalLaps != nil && expectedStint.ExpectedTotalLaps == nil {
				t.Errorf(`expected .Lines.Stints.TotalLaps "%v", got: "%d"`,
					expectedStint.ExpectedTotalLaps, *stint.TotalLaps)
			} else if *stint.TotalLaps != *expectedStint.ExpectedTotalLaps {
				t.Errorf(`expected .Lines.Stints.TotalLaps "%d", got: "%d"`,
					*expectedStint.ExpectedTotalLaps, *stint.TotalLaps)
			}

		}
	}

}
