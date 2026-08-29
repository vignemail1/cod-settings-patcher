package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
)

var desiredValues = map[string]string{
	"NvidiaReflex@":                     "Enabled",
	"BloodLimit@":                       "true",
	"BloodLimitInterval@":               "2000",
	"ShowBlood@":                        "false",
	"ShowBrass@":                        "false",
	"CorpseLimit@":                      "0",
	"GPUUploadHeaps@":                   "false",
	"PersistentDamageLayer@":            "false",
	"SubdivisionLevel@":                 "0",
	"WaterCausticsMode@":                "Off",
	"WaterWaveWetness@":                 "false",
	"WeatherGridVolumesQuality@":        "Off",
	"BulletImpacts@":                    "false",
	"CorpsesCullingThreshold@":          "0.500000",
	"TerrainQuality@":                   "Very Low",
	"Tessellation@":                     "0_Off",
	"DepthOfField@":                     "false",
	"DepthOfFieldQuality@":              "Low",
	"StaticSunshadowClipmapResolution@": "0",
	"EnableVelocityBasedBlur@":          "false",
	"SkipIntro@":                        "true",
	"SkipSeasonIntroVideo@":             "true",
	"SkipSeasonVideo@":                  "true",
	"ViewedSplashScreen@":               "true",
	"ShaderQuality@":                    "Low",
	"ReflectionProbeRelighting@":        "1",
	"ScreenSpaceShadowQuality@":         "Off",
	"SSRQuality@":                       "Off",
	"WorldStreamingQuality@":            "Low",
	"AmbientLightingQuality@":           "Off",
	"ModelQuality@":                     "Low Quality",
	"ParticleQuality@":                  "very low",
	"ShadowQuality@":                    "Very_Low",
	"VolumetricQuality@":                "QUALITY_LOW",
}

const rendererWorkerCountKey = "RendererWorkerCount@"

// SettingChange describes one value replacement planned for a settings line.
type SettingChange struct {
	Key      string
	OldValue string
	NewValue string
	Line     int
}

// FileChangePlan contains the original and planned contents for one settings file.
type FileChangePlan struct {
	Path    string
	Changes []SettingChange
	Before  []byte
	After   []byte
}

// ChangePlan contains all file changes proposed for one game installation.
type ChangePlan struct {
	Game                GameInstallation
	Files               []FileChangePlan
	RendererWorkerCount int
	CPU                 CPUCoreInfo
}

// HasChanges reports whether the plan contains at least one modified file.
func (p ChangePlan) HasChanges() bool {
	return len(p.Files) > 0
}

// ChangedFileCount returns the number of files that would be modified.
func (p ChangePlan) ChangedFileCount() int {
	return len(p.Files)
}

// ChangedSettingCount returns the total number of setting values that would change.
func (p ChangePlan) ChangedSettingCount() int {
	count := 0
	for _, file := range p.Files {
		count += len(file.Changes)
	}
	return count
}

func buildPlan(game GameInstallation) (ChangePlan, error) {
	workerCount, cpu, err := detectRendererWorkerCount()
	if err != nil {
		return ChangePlan{}, err
	}

	values := make(map[string]string, len(desiredValues)+1)
	for key, value := range desiredValues {
		values[key] = value
	}
	values[rendererWorkerCountKey] = strconv.Itoa(workerCount)

	plan := ChangePlan{
		Game:                game,
		RendererWorkerCount: workerCount,
		CPU:                 cpu,
	}

	for _, path := range game.Files {
		original, err := os.ReadFile(path) // #nosec G304 -- path originates from a discovered Call of Duty players directory.
		if err != nil {
			return ChangePlan{}, fmt.Errorf("lecture de %q : %w", path, err)
		}

		updated, changes := applyRules(original, values)
		if len(changes) == 0 {
			continue
		}

		plan.Files = append(plan.Files, FileChangePlan{
			Path:    path,
			Changes: changes,
			Before:  original,
			After:   updated,
		})
	}

	return plan, nil
}

func applyRules(data []byte, values map[string]string) ([]byte, []SettingChange) {
	var output bytes.Buffer
	output.Grow(len(data))

	var changes []SettingChange
	lineNumber := 1

	for len(data) > 0 {
		line, eol, remaining := splitPhysicalLine(data)
		updatedLine, changed, change := updateLineWithChange(line, lineNumber, values)
		if changed {
			changes = append(changes, change)
		}

		output.Write(updatedLine)
		output.Write(eol)
		data = remaining
		lineNumber++
	}

	return output.Bytes(), changes
}

func splitPhysicalLine(data []byte) (line, eol, rest []byte) {
	for i, b := range data {
		switch b {
		case '\n':
			if i > 0 && data[i-1] == '\r' {
				return data[:i-1], data[i-1 : i+1], data[i+1:]
			}
			return data[:i], data[i : i+1], data[i+1:]
		case '\r':
			if i+1 < len(data) && data[i+1] == '\n' {
				return data[:i], data[i : i+2], data[i+2:]
			}
			return data[:i], data[i : i+1], data[i+1:]
		}
	}
	return data, nil, nil
}

func updateLineWithChange(line []byte, lineNumber int, values map[string]string) ([]byte, bool, SettingChange) {
	if isCommentOnlyLine(line) {
		return line, false, SettingChange{}
	}

	equalIndex := bytes.IndexByte(line, '=')
	if equalIndex < 0 {
		return line, false, SettingChange{}
	}

	keyBytes := extractKey(line[:equalIndex])
	if len(keyBytes) == 0 {
		return line, false, SettingChange{}
	}

	key := string(keyBytes)
	newValue, managed := values[key]
	if !managed {
		return line, false, SettingChange{}
	}

	right := line[equalIndex+1:]
	valueStart := 0
	for valueStart < len(right) && isHorizontalWhitespace(right[valueStart]) {
		valueStart++
	}

	commentStart := findCommentStart(right[valueStart:])
	if commentStart >= 0 {
		commentStart += valueStart
	} else {
		commentStart = len(right)
	}

	valueEnd := commentStart
	for valueEnd > valueStart && isHorizontalWhitespace(right[valueEnd-1]) {
		valueEnd--
	}

	oldValue := string(right[valueStart:valueEnd])
	if oldValue == newValue {
		return line, false, SettingChange{}
	}

	updated := make([]byte, 0, len(line)-len(oldValue)+len(newValue))
	updated = append(updated, line[:equalIndex+1]...)
	updated = append(updated, right[:valueStart]...)
	updated = append(updated, newValue...)
	updated = append(updated, right[valueEnd:]...)

	return updated, true, SettingChange{
		Key:      key,
		OldValue: oldValue,
		NewValue: newValue,
		Line:     lineNumber,
	}
}

func extractKey(left []byte) []byte {
	trimmed := bytes.TrimLeft(left, " \t")
	at := bytes.IndexByte(trimmed, '@')
	if at <= 0 {
		return nil
	}
	return trimmed[:at+1]
}

func isCommentOnlyLine(line []byte) bool {
	return bytes.HasPrefix(bytes.TrimLeft(line, " \t"), []byte("//"))
}

func findCommentStart(data []byte) int {
	return bytes.Index(data, []byte("//"))
}

func isHorizontalWhitespace(b byte) bool {
	return b == ' ' || b == '\t'
}
