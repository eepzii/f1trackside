# Scrutineer

A Go reflection-based tool to validate Go structs against incoming JSON data. It is specifically designed for the archive `.jsonStream` files from F1, which you can get by navigating the root directory at: https://livetiming.formula1.com/static/2025/Index.json. (replace 2025 with another year to get another season's archive)

## Features
- **Missing Field Detection**: Flags keys present in the JSON input that are not defined as fields in your Go struct.
- **Unsafe Default Validation**: Detects when a primitive field (string, int, bool) receives an explicit zero-value from JSON, making it indistinguishable from a missing field => it will suggest using a pointer type (*string, *int, *bool).
- **Dynamic Array Support**: Automatically handles fields that have multiple value types like an array and a map. It will convert the array as well as the map in a string indexed map with any value.
- **Concurrency (USE WITH CAUTION):** Can be executed concurrently to speed up processing. However, this massively increases CPU usage and requires free memory up to 5x the size of your testing data. To enable this, run the tool with the `-concurrent` flag. 

## Usage
Run the tool by passing a YAML configuration file that defines the target Go type and the JSON paths to validate against.

Currently supported `onType` list:
- **DriverList**
- **Heartbeat**
- **LapCount**
- **SessionInfo**
- **SessionStatus**
- **TimingAppData**
- **TimingData**
- **TrackStatus**


```yaml
# scrutineer.yml

- check:
  onType: "DriverList"
  withPaths:
    - "/absolute/path"
    
- check: "optional text"
  onType: "TimingData"
  withPaths:
    - "/absolute/path"
```

```bash
go run . -config="/path/to/scrutineer.yml"
```