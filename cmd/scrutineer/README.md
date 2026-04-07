# Scrutineer

A Go reflection-based tool to validate Go structs against incoming JSON data. It is specifically designed for the archive `.jsonStream` files from F1, which you can get by navigating the root directory at: https://livetiming.formula1.com/static/2025/Index.json. (replace 2025 with another year to get another season's archive)<br><br>**Scrutineer** can also be used for live data, but on some supported types, you might encounter issues that require a custom unmarshaler. In these cases, the tool reaches its limits and cannot definitively confirm whether your custom unmarshaler works as intended.

## Features
- **Missing Field Detection**: Flags keys present in the JSON input that are not defined as fields in your Go struct.
- **Unsafe Default Validation**: Detects when a primitive field (string, int, bool) receives an explicit zero-value from JSON, making it indistinguishable from a missing field => it will suggest using a pointer type (*string, *int, *bool).
- **Dynamic Array Support**: Automatically handles fields that have multiple value types like an array and a map. It will convert the array as well as the map into a string-indexed map with any value.
- **Concurrency (USE WITH CAUTION):** Can be executed concurrently to speed up processing. However, this massively increases CPU usage and requires free memory up to 5x the size of your testing data. To enable this, run the tool with the `-concurrent` flag. 

## Usage
> **Note:** Scrutineer is designed as an internal workspace rather than a standalone importable package. To use it, you should clone/fork the repository, add your target structs directly to the internal `registry.go` file, and run the tool locally.

Run the tool by passing a YAML configuration file that defines the target Go type and the JSON paths to validate against.

Currently supported `onType` list: (for archive data)
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
  onType: "Heartbeat"
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

## Workflow
When profiling a new F1 API endpoint, do not attempt to guess the structure. Use Scrutineer iteratively to "unpack" the JSON step-by-step.

### 1. Start Empty
Begin by defining your target object as an empty struct.
```go
type Field struct{}
```
This newly defined struct needs to be added to the [registry](https://github.com/eepzii/f1trackside/blob/main/internal/scrutineer/registry.go):
```go
var TypeRegistry = map[string]any{
	"Field":     Field{},
}
```
Also don't forget to add this newly added type to the `onType` in the `config.yml`.

### 2. The First Run (`[+ MISSING]`)
Run the tool. You will see a tree of missing fields and type suggestions.
```
Field
├── A                      [+ MISSING]   -> suggest []any
├── B                      [+ MISSING]   -> suggest string
├── C                      [+ MISSING]   -> suggest struct/ map[string]any
├── D                      [+ MISSING]   -> suggest dynamic JSON array (struct/ map[string]any, []any)
├── E                      [+ MISSING]   -> suggest int
└── f                      [+ MISSING]   -> suggest bool
```
Add these fields to your Go struct with their respective tag for `json`.
```go
type Field struct {
	A struct{}              `json:"A"` // use struct{} on an array to get the data type of the elements inside
	B string                `json:"B"`
	C struct{}              `json:"C"` // use struct{} as a probe to look inside objects
	D DynamicJSON[struct{}] `json:"D"` // use DynamicJSON for dynamic json arrays
	E bool                  `json:"E"` // intentionally wrong type
	f bool                  `json:"f"` // intentionally unexported
}
```
> **Note:** We intentionally make a mistake on `E` and leave `f` unexported (lowercase) to demonstrate the next step.

### 3. Another Run (`[? CONFLICT]`)
Run the tool again. The output evolves:
```
Field
├── A                      [? CONFLICT]  -> suggest []any
│     ↳ array has item(s) with types of: "string"
│     ↳ => consider []string instead of []any
├── B                      [✓ OK]
├── C                      [? CONFLICT]  -> suggest struct/ map[string]any
│   ├── 16                 [+ MISSING]   -> suggest struct/ map[string]any
│   └── 81                 [+ MISSING]   -> suggest struct/ map[string]any
├── D                      [? CONFLICT]  -> suggest dynamic JSON array ([]any, struct/ map[string]any)
│     ↳ want: int, got: struct {}
├── E                      [? CONFLICT]  -> suggest int
│     ↳ want: int, got: bool
└── f                      [+ MISSING]   -> suggest bool
```
#### How to interpret this:
- `A` **array element type:** Because we used the `struct{}` probe on an array, the tool tells us the types of the elements inside. You can decide how to handle it, but it's highly recommended to use the actual type instead of `any`. In this case, use `[]string`.
- `B` **is** `[✓ OK]`**:** Perfect, move on.
- `C` **is a Map:** Because our `struct{}` probe revealed numbered keys (`16`, `81`), `C` is probably something more dynamic (a `map`), not a static object. Change to `map[string]struct{}`.
- `D` **DynamicJSON:** Here we got an error since we used `struct{}`. If the inner data had been a struct, the tool would have displayed those inner fields. However, in this case, the data inside is only an `int`, so it throws a mismatch error.    
- `E` **Type Mismatch:** The tool caught our intentional error. `E` expects an `int`, not a `bool`.
- `f` **is Missing:** Go's `json` package ignores unexported lowercase fields. Capitalize it to `F`.

```go
type Field struct {
	A []string            `json:"A"`
	B string              `json:"B"`
	C map[string]struct{} `json:"C"`
	D DynamicJSON[int]    `json:"D"`
	E int                 `json:"E"`
	F bool                `json:"f"`
}
```

### 4. More Fields, More Flags
```
Field
├── A                      [✓ OK]
├── B                      [✓ OK]
├── C                      [✓ OK]
│   ├── CA                 [+ MISSING]   -> suggest struct/ map[string]any
│   ├── CB                 [+ MISSING]   -> suggest dynamic JSON array (struct/ map[string]any, []any)
│   └── CC                 [+ MISSING]   -> suggest bool
├── D                      [! POINTER]   (9 defaults) -> []*any
├── E                      [! POINTER]   (3 defaults) -> *int
└── f                      [✓ OK]
```
We have three more `[✓ OK]` flags, but the output has evolved again. Here is exactly what is happening and how to fix it:
- **Unlocking Nested Fields (**`C`**):** Notice the new fields (`CA`, `CB`, `CC`)? Because we previously changed `C` to `map[string]struct{}`, the tool is now looking inside those map elements. We need to create a new `CField` struct to handle these newly discovered fields.
- **The Pointer Rule (**`[! POINTER]`**):** Why do `D` and `E` suddenly want pointers (`*`)? The F1 API often sends partial payloads where fields are omitted entirely. If you use a standard `int`, Go's unmarshaler defaults it to `0`. This makes it impossible to know if the API actually sent a `0`, or if it sent nothing at all. By changing the type to a pointer (`*int`), omitted fields safely default to `nil`.
- **The Lowercase** `f`**:** Why does the output still say `f` when we capitalized it to `F` in our Go code? The Scrutineer displays the JSON key name, not the Go field name. Because it says `[✓ OK]`, your Go struct is totally fine.

```go
type Field struct {
	A []string          `json:"A"`
	B string            `json:"B"`
	C map[string]CField `json:"C"`
	D DynamicJSON[*int] `json:"D"`
	E *int              `json:"E"`
	F bool              `json:"f"`
}

type CField struct {
	CA struct{}              `json:"CA"`
	CB DynamicJSON[struct{}] `json:"CB"`
	CC bool                  `json:"CC"`
}
```

### 5. Repeat
```
Field
├── A                      [✓ OK]
├── B                      [✓ OK]
├── C                      [✓ OK]
│   ├── CA                 [? CONFLICT]  -> suggest struct/ map[string]any
│   │   ├── CAA            [+ MISSING]   -> suggest int
│   │   ├── CAB            [+ MISSING]   -> suggest string
│   │   └── CAC            [+ MISSING]   -> suggest []any
│   ├── CB                 [✓ OK]
│   │   ├── CBA            [+ MISSING]   -> suggest int
│   │   └── CBB            [+ MISSING]   -> suggest string
│   └── CC                 [! POINTER]   (72 defaults) -> *bool
├── D                      [✓ OK]
├── E                      [✓ OK]
└── f                      [✓ OK]
```

Although we are not done with the whole implementation of our `Field` struct, this should be more than enough to move forward on your own and implement all this by yourself.
> **NOTE:** In case you get stuck, you can use the section below to look at some specific cases on their own and how you can resolve them.

## Things To Note
- **The Probing Strategy:** If the tool suggests `[]any` or `struct/ map[string]any`, always start by using `struct{}` and **not** `[]struct{}` or `map[string]struct{}` directly out of the box. When using `struct{}` and you get numbers or possible abbreviations with all the same type suggestion, you are likely dealing with a `map`.
- **Missing Composite Type:** If you have some field which you determined should be implemented as `map[string]Field` or `[]Field` but you assign `Field` without the composite type (map/slice), then the tool will spit out something like this:
  ```
  │   ├── Field               [✓ OK]
  │   │   ├── 16              [+ MISSING]   -> suggest struct/ map[string]any
  │   │   ├── 31              [+ MISSING]   -> suggest struct/ map[string]any
  │   │   ├── 33              [+ MISSING]   -> suggest struct/ map[string]any
  │   │   ├── Name            [- UNUSED]    defined in Go struct but not found in JSON
  │   │   ├── Age             [- UNUSED]    defined in Go struct but not found in JSON
  │   │   ├── 44              [+ MISSING]   -> suggest struct/ map[string]any
  │   │   ├── Height          [- UNUSED]    defined in Go struct but not found in JSON
  │   │   └── Weight          [- UNUSED]    defined in Go struct but not found in JSON
  ```
**Fix:** This heavily implies you need to wrap your type in a slice or a map (`map[string]`) or just use `struct{}` again to see what the tool suggests.

- **The** `[- UNUSED]` **Flag & False Positives:**<br>
This flag will appear when a field you defined in your Go struct is entirely missing from the JSON input data.<br><br>**Warning:** Do not immediately delete unused fields. Only remove them after a reasonable amount of time has passed where you think that the endpoint will not send this field in the foreseeable future, and also only if you have tested against a large dataset. You might want to test against an entire season, or even multiple years of data, to be absolutely sure. Don't delete prematurely!