package types

import (
	"encoding/json"
	"testing"
)

// TODO: refactor duplicate pointer validation logic into a helper function to reduce verbose code.
// and check over fields again that might need to be a pointer value to handle empty vs. nil updates
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
			t.Errorf("missing %s driver from map", test.DriverNumber)
			continue
		}

		if driver.RacingNumber != test.ExpectedRacingNumber {
			t.Errorf(`expected .Lines.RacingNumber "%s", got: "%s"`,
				test.ExpectedRacingNumber, driver.RacingNumber)
		}

		if driver.Status == nil && test.ExpectedStatus == nil {
			// ok
		} else if driver.Status == nil && test.ExpectedStatus != nil {
			t.Errorf(`expected .Lines.Status "%d", got: "%v"`,
				*test.ExpectedStatus, driver.Status)
		} else if driver.Status != nil && test.ExpectedStatus == nil {
			t.Errorf(`expected .Lines.Status "%v", got: "%d"`,
				test.ExpectedStatus, *driver.Status)
		} else if *driver.Status != *test.ExpectedStatus {
			t.Errorf(`expected .Lines.Status "%d", got: "%d"`,
				*test.ExpectedStatus, *driver.Status)
		}

		if driver.InPit == nil && test.ExpectedInPit == nil {
			// ok
		} else if driver.InPit == nil && test.ExpectedInPit != nil {
			t.Errorf(`expected .Lines.InPit "%v", got: "%v"`,
				*test.ExpectedInPit, driver.InPit)
		} else if driver.InPit != nil && test.ExpectedInPit == nil {
			t.Errorf(`expected .Lines.InPit "%v", got: "%v"`,
				test.ExpectedInPit, *driver.InPit)
		} else if *driver.InPit != *test.ExpectedInPit {
			t.Errorf(`expected .Lines.InPit "%v", got: "%v"`,
				*test.ExpectedInPit, *driver.InPit)
		}

		if len(driver.Sectors) != test.ExpectedSectorMapSize {
			t.Errorf(`expected .Lines.Sectors length "%d", got: "%d"`,
				test.ExpectedSectorMapSize, len(driver.Sectors))
		}

		for key, sector := range driver.Sectors {

			expectedSector, ok := test.ExpectedSectors[key]
			if !ok {
				t.Errorf(`unexpected sector map key "%s" on driver number "%s"`, key, test.DriverNumber)
				continue
			}

			if sector.Value != expectedSector.ExpectedValue {
				t.Errorf(`expected .Lines.Sectors.Value "%s", got: "%s"`,
					expectedSector.ExpectedValue, sector.Value)
			}

			if sector.Status == nil && expectedSector.ExpectedStatus == nil {
				// ok
			} else if sector.Status == nil && expectedSector.ExpectedStatus != nil {
				t.Errorf(`expected .Lines.Sectors.Status "%d", got: "%v"`,
					*expectedSector.ExpectedStatus, sector.Status)
				continue
			} else if sector.Status != nil && expectedSector.ExpectedStatus == nil {
				t.Errorf(`expected .Lines.Sectors.Status "%v", got: "%d"`,
					expectedSector.ExpectedStatus, *sector.Status)
				continue
			} else if *sector.Status != *expectedSector.ExpectedStatus {
				t.Errorf(`expected .Lines.Sectors.Status "%d", got: "%d"`,
					*expectedSector.ExpectedStatus, *sector.Status)
			}

			if len(sector.Segments) != expectedSector.ExpectedSegmentMapSize {
				t.Errorf(`expected .Lines.Sectors.Segments length "%d", got: "%d"`,
					expectedSector.ExpectedSegmentMapSize, len(sector.Segments))
			}

			for segmentMapKey, segment := range sector.Segments {

				expectedSegment, ok := expectedSector.ExpectedSegments[segmentMapKey]
				if !ok {
					t.Errorf(`unexpected segment map key "%s" in sector "%s"`, segmentMapKey, key)
					continue
				}

				if segment.Status == nil && expectedSegment.ExpectedStatus == nil {
					// ok
				} else if segment.Status == nil && expectedSegment.ExpectedStatus != nil {
					t.Errorf(`expected .Lines.Sectors.Segments.Status "%d", got: "%v"`,
						*expectedSegment.ExpectedStatus, segment.Status)
				} else if segment.Status != nil && expectedSegment.ExpectedStatus == nil {
					t.Errorf(`expected .Lines.Sectors.Segments.Status "%v", got: "%d"`,
						expectedSegment.ExpectedStatus, *segment.Status)
				} else if *segment.Status != *expectedSegment.ExpectedStatus {
					t.Errorf(`expected .Lines.Sectors.Segments.Status "%d", got: "%d"`,
						*expectedSegment.ExpectedStatus, *segment.Status)
				}

			}

		}

		if len(driver.Speeds) != test.ExpectedSpeedMapSize {
			t.Errorf(`expected .Lines.Speeds length "%d", got: "%d"`,
				test.ExpectedSpeedMapSize, len(driver.Speeds))
		}

		for key, speed := range driver.Speeds {

			expectedSpeed, ok := test.ExpectedSpeeds[key]
			if !ok {
				t.Errorf(`unexpected speed map key "%s" on driver number "%s"`, key, test.DriverNumber)
				continue
			}

			if speed.Value != expectedSpeed.ExpectedValue {
				t.Errorf(`expected .Lines.Speeds.Value "%s", got: "%s"`,
					expectedSpeed.ExpectedValue, speed.Value)
			}

			if speed.Status == nil && expectedSpeed.ExpectedStatus == nil {
				// ok
			} else if speed.Status == nil && expectedSpeed.ExpectedStatus != nil {
				t.Errorf(`expected .Lines.Speeds.Status "%d", got: "%v"`,
					*expectedSpeed.ExpectedStatus, speed.Status)
				continue
			} else if speed.Status != nil && expectedSpeed.ExpectedStatus == nil {
				t.Errorf(`expected .Lines.Speeds.Status "%v", got: "%d"`,
					expectedSpeed.ExpectedStatus, *speed.Status)
				continue
			} else if *speed.Status != *expectedSpeed.ExpectedStatus {
				t.Errorf(`expected .Lines.Speeds.Status "%d", got: "%d"`,
					*expectedSpeed.ExpectedStatus, *speed.Status)
			}

			if speed.OverallFastest == nil && expectedSpeed.ExpectedOverallFastest == nil {
				// ok
			} else if speed.OverallFastest == nil && expectedSpeed.ExpectedOverallFastest != nil {
				t.Errorf(`expected .Lines.Speeds.OverallFastest "%v", got: "%v"`,
					*expectedSpeed.ExpectedOverallFastest, speed.OverallFastest)
			} else if speed.OverallFastest != nil && expectedSpeed.ExpectedOverallFastest == nil {
				t.Errorf(`expected .Lines.Speeds.OverallFastest "%v", got: "%v"`,
					expectedSpeed.ExpectedOverallFastest, *speed.OverallFastest)
			} else if *speed.OverallFastest != *expectedSpeed.ExpectedOverallFastest {
				t.Errorf(`expected .Lines.Speeds.OverallFastest "%v", got: "%v"`,
					*expectedSpeed.ExpectedOverallFastest, *speed.OverallFastest)
			}

		}

		if driver.BestLapTime.Value != test.ExpectedBestLapTimeValue {
			t.Errorf(`expected .Lines.BestLapTime.Value "%s", got: "%s"`,
				test.ExpectedBestLapTimeValue, driver.BestLapTime.Value)
		}

		if driver.BestLapTime.Lap != test.ExpectedBestLapTimeLap {
			t.Errorf(`expected .Lines.BestLapTime.Lap "%d", got: "%d"`,
				test.ExpectedBestLapTimeLap, driver.BestLapTime.Lap)
		}

		if driver.LastLapTime.Value != test.ExpectedLastLapTimeValue {
			t.Errorf(`expected .Lines.LastLapTime.Value "%s", got: "%s"`,
				test.ExpectedLastLapTimeValue, driver.LastLapTime.Value)
		}

		if driver.LastLapTime.Status == nil && test.ExpectedLastLapTimeStatus == nil {
			// ok
		} else if driver.LastLapTime.Status == nil && test.ExpectedLastLapTimeStatus != nil {
			t.Errorf(`expected .Lines.LastLapTime.Status "%d", got: "%v"`,
				*test.ExpectedLastLapTimeStatus, driver.LastLapTime.Status)
		} else if driver.LastLapTime.Status != nil && test.ExpectedLastLapTimeStatus == nil {
			t.Errorf(`expected .Lines.LastLapTime.Status "%v", got: "%d"`,
				test.ExpectedLastLapTimeStatus, *driver.LastLapTime.Status)
		} else if *driver.LastLapTime.Status != *test.ExpectedLastLapTimeStatus {
			t.Errorf(`expected .Lines.LastLapTime.Status "%d", got: "%d"`,
				*test.ExpectedLastLapTimeStatus, *driver.LastLapTime.Status)
		}

		if driver.LastLapTime.PersonalFastest == nil && test.ExpectedLastLapTimePersonalFastest == nil {
			// ok
		} else if driver.LastLapTime.PersonalFastest == nil && test.ExpectedLastLapTimePersonalFastest != nil {
			t.Errorf(`expected .Lines.LastLapTime.PersonalFastest "%v", got: "%v"`,
				*test.ExpectedLastLapTimePersonalFastest, driver.LastLapTime.PersonalFastest)
		} else if driver.LastLapTime.PersonalFastest != nil && test.ExpectedLastLapTimePersonalFastest == nil {
			t.Errorf(`expected .Lines.LastLapTime.PersonalFastest "%v", got: "%v"`,
				test.ExpectedLastLapTimePersonalFastest, *driver.LastLapTime.PersonalFastest)
		} else if *driver.LastLapTime.PersonalFastest != *test.ExpectedLastLapTimePersonalFastest {
			t.Errorf(`expected .Lines.LastLapTime.PersonalFastest "%v", got: "%v"`,
				*test.ExpectedLastLapTimePersonalFastest, *driver.LastLapTime.PersonalFastest)
		}

	}

}
