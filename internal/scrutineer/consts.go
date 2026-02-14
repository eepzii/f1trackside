package scrutineer

import "github.com/eepzii/f1trackside/internal/types"

var TYPE_REGISTRY = map[string]any{
	"DriverList":    map[string]types.Driver{},
	"Heartbeat":     types.Heartbeat{},
	"LapCount":      types.LapCount{},
	"SessionInfo":   types.SessionInfo{},
	"SessionStatus": types.SessionStatus{},
	"TimingAppData": types.TimingAppData{},
	"TimingData":    types.TimingData{},
	"TrackStatus":   types.TrackStatus{},
}
