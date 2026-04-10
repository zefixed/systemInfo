package systeminfo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultMemInfoPath = "/proc/meminfo"

// Mem holds memory statistics parsed from /proc/meminfo.
// All values are stored in the current unit, which defaults to bytes (B).
// Use Format to convert values to a different unit.
type Mem struct {
	currentUnit string
	Total       float64 `json:"total"`
	Free        float64 `json:"free"`
	Buffers     float64 `json:"buffers"`
	Cached      float64 `json:"cached"`
	Used        float64 `json:"used"`
}

// NewMem creates a new Mem instance and populates it with current memory stats.
// Returns an error if /proc/meminfo cannot be read.
func NewMem() (*Mem, error) {
	m := &Mem{
		currentUnit: "KB",
	}
	if err := m.loadFrom(defaultMemInfoPath); err != nil {
		return nil, err
	}
	return m, nil
}

// CurrentUnit returns the unit in which memory values are currently expressed.
func (m *Mem) CurrentUnit() string {
	return m.currentUnit
}

// Refresh re-reads /proc/meminfo and updates all memory values.
// Resets the current unit back to kilobytes (KB).
func (m *Mem) Refresh() error {
	if err := m.loadFrom(defaultMemInfoPath); err != nil {
		return err
	}
	m.currentUnit = "KB"
	return nil
}

// loadFrom reads and parses the meminfo file at the given path.
// It is separated from Refresh to make the logic testable with custom files.
func (m *Mem) loadFrom(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return ErrReadFile
	}

	info := make(map[string]uint64)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		val, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		info[key] = val
	}

	m.Total = float64(info["MemTotal"])
	m.Free = float64(info["MemFree"])
	m.Buffers = float64(info["Buffers"])
	m.Cached = float64(info["Cached"])
	m.Used = float64(info["MemTotal"] - info["MemAvailable"])
	return nil
}

// Format converts all memory values from the current unit to the specified unit.
// Supported units: "b" (bits), "B" (bytes), "KB", "MB", "GB", "TB".
// Returns ErrUnknownUnit if the given unit is not supported.
//
// Example:
//
//	m, _ := NewMem()
//	m.Format("MB")
//	fmt.Printf("Total: %.2f MB\n", m.Total)
func (m *Mem) Format(unit string) error {
	units := map[string]float64{
		"b":  0.125,
		"B":  1,
		"KB": 1 << 10,
		"MB": 1 << 20,
		"GB": 1 << 30,
		"TB": 1 << 40,
	}

	target, ok := units[unit]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownUnit, unit)
	}
	current, ok := units[m.currentUnit]
	if !ok {
		return fmt.Errorf("%w: %s", ErrUnknownUnit, m.currentUnit)
	}

	divisor := target / current
	m.Total /= divisor
	m.Free /= divisor
	m.Buffers /= divisor
	m.Cached /= divisor
	m.Used /= divisor

	m.currentUnit = unit
	return nil
}
