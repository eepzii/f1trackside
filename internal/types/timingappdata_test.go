package types

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseTimingAppData(t *testing.T) {

	tests := []struct {
		name  string
		input []byte
		want  TimingAppDataResponse
	}{
		{
			name: "general .TimingAppData and with .TimingAppData.Lines[44].Stints json arrays",
			input: []byte(`{
                "TimingAppData":{
                    "Lines":{
                        "44":{
                            "RacingNumber":"44",
                            "Line":3,
                            "GridPos":"3",
                            "Stints":[
                                {
                                    "LapTime":"1:40.526",
                                    "LapNumber":3,
                                    "LapFlags":0,
                                    "Compound":"UNKNOWN",
                                    "New":"false",
                                    "TyresNotChanged":"0",
                                    "TotalLaps":3,
                                    "StartLaps":0
                                }
                            ]
                        }
                    }
                }
            }`),
			want: TimingAppDataResponse{
				TimingAppData: TimingAppData{
					Lines: map[string]TimingAppDataLine{
						"44": {
							RacingNumber: "44",
							Line:         3,
							GridPos:      "3",
							Stints: map[string]Stint{
								"0": {
									LapTime:         "1:40.526",
									LapNumber:       3,
									LapFlags:        valToPtr(0),
									Compound:        "UNKNOWN",
									New:             "false",
									TyresNotChanged: "0",
									TotalLaps:       valToPtr(3),
									StartLaps:       valToPtr(0),
								},
							},
						},
					},
				},
			},
		},
		{
			name: ".TimingAppData.Lines[16].Stints with multiple elements in json array",
			input: []byte(`{
                "TimingAppData":{
                    "Lines":{
                        "16":{
                            "RacingNumber":"16",
                            "Stints":[
                                {
                                    "Compound":"MEDIUM",
                                    "New":"true",
                                    "TyresNotChanged":"0",
                                    "TotalLaps":22
                                },
                                {
                                    "Compound":"HARD",
                                    "TyresNotChanged":"1",
                                    "TotalLaps":0,
                                    "New":"true"
                                }
                            ]
                        }
                    }
                }
            }`),
			want: TimingAppDataResponse{
				TimingAppData: TimingAppData{
					Lines: map[string]TimingAppDataLine{
						"16": {
							RacingNumber: "16",
							Stints: map[string]Stint{
								"0": {
									Compound:        "MEDIUM",
									New:             "true",
									TyresNotChanged: "0",
									TotalLaps:       valToPtr(22),
								},
								"1": {
									Compound:        "HARD",
									TyresNotChanged: "1",
									TotalLaps:       valToPtr(0),
									New:             "true",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "multiple .TimingAppData.Lines with json maps on .TimingAppData.Lines[key].Stints",
			input: []byte(`{
                "TimingAppData":{
                    "Lines":{
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
                        }
                    }
                }
            }`),
			want: TimingAppDataResponse{
				TimingAppData: TimingAppData{
					Lines: map[string]TimingAppDataLine{
						"22": {
							Stints: map[string]Stint{
								"0": {
									TotalLaps: valToPtr(1),
								},
							},
						},
						"43": {
							Stints: map[string]Stint{
								"0": {
									Compound: "SOFT",
								},
							},
						},
					},
				},
			},
		},
		{
			name: ".TimingAppData.Lines[18]",
			input: []byte(`{
                "TimingAppData":{
                    "Lines":{
                        "18":{}
                    }
                }
            }`),
			want: TimingAppDataResponse{
				TimingAppData: TimingAppData{
					Lines: map[string]TimingAppDataLine{
						"18": {},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var got TimingAppDataResponse
			if err := json.Unmarshal(test.input, &got); err != nil {
				t.Fatalf("failed to unmarshal json: %v", err)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("mismatch testing %s (-want +got):\n%s", test.name, diff)
			}
		})
	}

}
