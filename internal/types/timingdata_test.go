package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseTimingData(t *testing.T) {

	tests := []struct {
		name  string
		input []byte
		want  TimingDataResponse
	}{
		{
			name: ".TimingData with json arrays",
			input: []byte(`{
                "TimingData":{
                    "Lines":{
                        "55":{
                            "RacingNumber":"55",
                            "Status":0,
                            "InPit":true,
                            "Sectors":[
                                {
                                    "Value":"38.404",
                                    "Status":1,
                                    "Segments":[
                                        {
                                            "Status":2
                                        },
                                        {
                                            "Status":3
                                        }
                                    ]
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
                        }
                    }
                }
            }`),
			want: TimingDataResponse{
				TimingData: TimingData{
					Lines: map[string]TimingDataLine{
						"55": {
							RacingNumber: "55",
							Status:       valToPtr(0),
							InPit:        valToPtr(true),
							Sectors: map[string]Sector{
								"0": {
									Value:  valToPtr("38.404"),
									Status: valToPtr(1),
									Segments: map[string]Segment{
										"0": {valToPtr(2)},
										"1": {valToPtr(3)},
									},
								},
							},
							Speeds: map[string]Speed{
								"I1": {
									Value:           valToPtr("444"),
									Status:          valToPtr(4),
									OverallFastest:  valToPtr(true),
									PersonalFastest: valToPtr(true),
								},
							},
							LastLapTime: LastLapTime{
								Status: valToPtr(5),
								Value:  valToPtr(""),
							},
						},
					},
				},
			},
		},
		{
			name: ".TimingData with json maps",
			input: []byte(`{
    			"TimingData":{
                    "Lines":{
                        "23":{
                            "IntervalToPositionAhead":{
                                "Value":"+1.659"
                            },
                            "Sectors":{
                                "0":{
                                    "Segments":{
                                        "7":{
                                            "Status":6
                                        }
                                    }
                                },
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
                            "LastLapTime":{
                                "Value":"1:39.360",
                                "PersonalFastest":false
                            }
                        }
                    }
                }
            }`),
			want: TimingDataResponse{
				TimingData: TimingData{
					Lines: map[string]TimingDataLine{
						"23": {
							IntervalToPositionAhead: IntervalToPositionAhead{
								Value: valToPtr("+1.659"),
							},
							Sectors: map[string]Sector{
								"0": {
									Segments: map[string]Segment{
										"7": {valToPtr(6)},
									},
								},
								"2": {
									Value: valToPtr("38.150"),
								},
							},
							Speeds: map[string]Speed{
								"FL": {
									Value:          valToPtr("322"),
									OverallFastest: valToPtr(false),
								},
							},
							LastLapTime: LastLapTime{
								Value:           valToPtr("1:39.360"),
								PersonalFastest: valToPtr(false),
							},
						},
					},
				},
			},
		},
		{
			name:  "default values (for pointers = nil)",
			input: []byte(`{"TimingData":{"Lines":{"81":{}}}}`),
			want: TimingDataResponse{
				TimingData: TimingData{
					Lines: map[string]TimingDataLine{
						"81": {},
					},
				},
			},
		},
		{
			name: "empty .TimingData input",
			// json: []byte(`{"TimingData":{"Lines":{}}}`),
			input: []byte(`{"TimingData":{}}`),
			want:  TimingDataResponse{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var got TimingDataResponse
			if err := json.Unmarshal(test.input, &got); err != nil {
				t.Fatalf("failed to unmarshal json: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("mismatch testing %s (-want +got):\n%s", test.name, diff)
			}
		})
	}

}
