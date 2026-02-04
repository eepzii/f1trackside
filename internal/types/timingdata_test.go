package types

import (
	"encoding/json"
	"testing"
)

// TODO: refactor test table
func TestParseTimingData(t *testing.T) {

	jsonInput := []byte(`{
        "TimingData":{
            "Lines":{
                "55":{
                    "RacingNumber":"55",
                    "Status":0,
                    "InPit":true,
                    "Sectors":[{
                        "Value":"38.404",
					    "Status":1,
                        "Segments": [
                            {"Status":2},
                            {"Status":3}]
					    }
                    ],
				    "Speeds":{
                        "I1":{
                            "Value":"444",
                            "Status":4,
                            "OverallFastest":true,
                            "PersonalFastest":true
                        }
                    },
				    "LastLapTime":{
                        "Value":"",
					    "Status":5
                    }
			    },
                "81":{
                    "GapToLeader":"+7.649",
                    "IntervalToPositionAhead":{
                        "Value":"+1.659"
                    },
					"InPit":false,
					"BestLapTime":{
                        "Value":"1:38.299",
						"Lap":4
                    }
                },
                "23":{
                    "Sectors":{
                        "1":{
                            "Segments":{
                                "1":{
                                    "Status":6
                                }
                            }
                        }
                    }
                },
                "87":{
                    "NumberOfLaps":5,
                    "Sectors":{
                        "2":{
                            "Value":"38.150"
                        }
                    },
                    "Speeds":{
                        "FL":{
                            "Value":"322",
                            "OverallFastest":false
                        }
                    },
					"BestLapTime":{
                        "Value":""
                    },
                    "LastLapTime":{
                        "Value":"1:39.360",
                        "PersonalFastest":false
                    }
                }
            }
        }
    }`)

	var response TimingDataResponse
	if err := json.Unmarshal(jsonInput, &response); err != nil {
		t.Fatalf("failed to unmarshal json: %v", err)
	}

	if response.TimingData.Lines == nil {
		t.Fatalf(".Lines is nil")
	}

	var zero = 0
	var one = 1
	var two = 2
	var three = 3
	var four = 4
	var five = 5
	var six = 6
	var truePtr = true
	var falsePtr = false

	tests := []struct {
		DriverNumber          string
		ExpectedRacingNumber  string
		ExpectedStatus        *int
		ExpectedInPit         *bool
		ExpectedSectorMapSize int
		ExpectedSectors       map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedSegmentMapSize int
			ExpectedSegments       map[string]struct {
				ExpectedStatus *int
			}
		}
		ExpectedSpeedMapSize int
		ExpectedSpeeds       map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedOverallFastest *bool
		}
		ExpectedBestLapTimeValue           string
		ExpectedBestLapTimeLap             int
		ExpectedLastLapTimeValue           string
		ExpectedLastLapTimeStatus          *int
		ExpectedLastLapTimePersonalFastest *bool
	}{
		{"55", "55", &zero, &truePtr, 1, map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedSegmentMapSize int
			ExpectedSegments       map[string]struct{ ExpectedStatus *int }
		}{
			"0": {"38.404", &one, 2, map[string]struct {
				ExpectedStatus *int
			}{
				"0": {&two},
				"1": {&three},
			}},
		}, 1, map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedOverallFastest *bool
		}{"I1": {"444", &four, &truePtr}},
			"", 0, "", &five, nil,
		},
		{"81", "", nil, &falsePtr, 0, nil, 0, nil, "1:38.299", 4, "", nil, nil},
		{"23", "", nil, nil, 1, map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedSegmentMapSize int
			ExpectedSegments       map[string]struct{ ExpectedStatus *int }
		}{
			"1": {"", nil, 1, map[string]struct {
				ExpectedStatus *int
			}{
				"1": {&six},
			}},
		},
			0, nil, "", 0, "", nil, nil,
		},
		{"87", "", nil, nil, 1, map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedSegmentMapSize int
			ExpectedSegments       map[string]struct{ ExpectedStatus *int }
		}{
			"2": {"38.150", nil, 0, nil},
		}, 1, map[string]struct {
			ExpectedValue          string
			ExpectedStatus         *int
			ExpectedOverallFastest *bool
		}{"FL": {"322", nil, &falsePtr}},
			"", 0, "1:39.360", nil, &falsePtr,
		},
	}

	for _, test := range tests {

		driver, exists := response.TimingData.Lines[test.DriverNumber]
		if !exists {
			t.Errorf("missing key on .Lines: %s", test.DriverNumber)
			continue
		}

		err := validateValues(driver.RacingNumber, test.ExpectedRacingNumber)
		if err != nil {
			t.Errorf(".Lines[%s].RacingNumber -> %v", test.DriverNumber, err)
		}

		err = validatePointers(driver.Status, test.ExpectedStatus)
		if err != nil {
			t.Errorf(".Lines[%s].Status -> %v", test.DriverNumber, err)
		}

		err = validatePointers(driver.InPit, test.ExpectedInPit)
		if err != nil {
			t.Errorf(".Lines[%s].InPit -> %v", test.DriverNumber, err)
		}

		err = validateValues(len(driver.Sectors), test.ExpectedSectorMapSize)
		if err != nil {
			t.Errorf(".Lines[%s].Sectors map size -> %v", test.DriverNumber, err)
		}

		for key, sector := range driver.Sectors {

			expectedSector, ok := test.ExpectedSectors[key]
			if !ok {
				t.Errorf("missing key on .Lines[%s].Sectors: %s", test.DriverNumber, key)
				continue
			}

			err = validateValues(sector.Value, expectedSector.ExpectedValue)
			if err != nil {
				t.Errorf(".Lines[%s].Sectors[%s].Value -> %v",
					test.DriverNumber, key, err)
			}

			err = validatePointers(sector.Status, expectedSector.ExpectedStatus)
			if err != nil {
				t.Errorf(".Lines[%s].Sectors[%s].Status -> %v",
					test.DriverNumber, key, err)
			}

			err = validateValues(len(sector.Segments), expectedSector.ExpectedSegmentMapSize)
			if err != nil {
				t.Errorf(".Lines[%s].Sectors[%s].Segments map size -> %v",
					test.DriverNumber, key, err)
			}

			for segmentMapKey, segment := range sector.Segments {

				expectedSegment, ok := expectedSector.ExpectedSegments[segmentMapKey]
				if !ok {
					t.Errorf("missing key on .Lines[%s].Sectors[%s].Segments: %s",
						test.DriverNumber, key, segmentMapKey)
					continue
				}

				err = validatePointers(segment.Status, expectedSegment.ExpectedStatus)
				if err != nil {
					t.Errorf(".Lines[%s].Sectors[%s].Segments[%s].Status -> %v",
						test.DriverNumber, key, segmentMapKey, err)
				}

			}

		}

		err = validateValues(len(driver.Speeds), test.ExpectedSpeedMapSize)
		if err != nil {
			t.Errorf(".Lines[%s].Speeds map size -> %v", test.DriverNumber, err)
		}

		for key, speed := range driver.Speeds {

			expectedSpeed, ok := test.ExpectedSpeeds[key]
			if !ok {
				t.Errorf("missing key on .Lines[%s].Speeds: %s", test.DriverNumber, key)
				continue
			}

			err = validateValues(speed.Value, expectedSpeed.ExpectedValue)
			if err != nil {
				t.Errorf(".Lines[%s].Speeds[%s].Value -> %v",
					test.DriverNumber, key, err)
			}

			err = validatePointers(speed.Status, expectedSpeed.ExpectedStatus)
			if err != nil {
				t.Errorf(".Lines[%s].Speeds[%s].Status -> %v",
					test.DriverNumber, key, err)
			}

			err = validatePointers(speed.OverallFastest, expectedSpeed.ExpectedOverallFastest)
			if err != nil {
				t.Errorf(".Lines[%s].Speeds[%s].OverallFastest -> %v",
					test.DriverNumber, key, err)
			}

		}

		err = validateValues(driver.BestLapTime.Value, test.ExpectedBestLapTimeValue)
		if err != nil {
			t.Errorf(".Lines[%s].BestLapTime.Value -> %v", test.DriverNumber, err)
		}

		err = validateValues(driver.BestLapTime.Lap, test.ExpectedBestLapTimeLap)
		if err != nil {
			t.Errorf(".Lines[%s].BestLapTime.Lap -> %v", test.DriverNumber, err)
		}

		err = validateValues(driver.LastLapTime.Value, test.ExpectedLastLapTimeValue)
		if err != nil {
			t.Errorf(".Lines[%s].LastLapTime.Value -> %v", test.DriverNumber, err)
		}

		err = validatePointers(driver.LastLapTime.Status, test.ExpectedLastLapTimeStatus)
		if err != nil {
			t.Errorf(".Lines[%s].LastLapTime.Status -> %v", test.DriverNumber, err)
		}

		err = validatePointers(driver.LastLapTime.PersonalFastest, test.ExpectedLastLapTimePersonalFastest)
		if err != nil {
			t.Errorf(".Lines[%s].LastLapTime.PersonalFastest -> %v", test.DriverNumber, err)
		}

	}

}
