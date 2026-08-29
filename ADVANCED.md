# Advanced Documentation

This page covers the technical details, settings list, and development instructions for Call of Duty Settings Patcher.

## Configuration discovery

The tool looks for configurations under:

```text
%LOCALAPPDATA%\Activision\Call of Duty
```

It detects the following directories:

- `players` for full game installations;
- `playersBeta` for beta versions.

Files located directly in these directories with the `.txt`, `.txt0`, and `.txt1` extensions are analyzed. Backup files with `.backup-` in their names are ignored.

## Write safety

Before writing anything, the application:

1. builds an in-memory change plan;
2. shows the game, its directory, the files, and every `old value → new value` transition;
3. waits for explicit confirmation with `y`;
4. reads the files again to ensure they have not changed since the preview;
5. creates a timestamped backup of every file;
6. writes through a temporary file in the same directory, then replaces the original atomically;
7. reads the file again to verify the written content;
8. attempts a rollback from backups if a multi-file transaction fails.

Backups are named as follows:

```text
<file-name>.backup-YYYYMMDD-HHMMSS.nanoseconds
```

The engine works directly on file bytes to preserve `//` comments, trailing spaces and tabs, and `LF`, `CRLF`, and `CR` line endings.

## RendererWorkerCount

`RendererWorkerCount@` is calculated from the Windows CPU topology without using logical threads:

1. Windows topology is queried to identify P-cores when they are available;
2. otherwise, the physical core count is used;
3. logical cores, SMT, and Hyper-Threading are never used;
4. the applied value is `ceil(80% of the selected core count)`;
5. it never exceeds `core count - 1`.

Examples:

| Selected cores | Calculation | Value |
|---:|---:|---:|
| 12 physical cores | `ceil(12 × 0.80)` | `10` |
| 8 P-cores | `ceil(8 × 0.80)`, maximum `8 - 1` | `7` |
| 6 physical cores | `ceil(6 × 0.80)` | `5` |
| 20 P-cores | `ceil(20 × 0.80)` | `16` |

## Applied settings

Keys are matched in the `KeyName@... = value` form. The portion after `@`, spacing around `=`, optional comments, and line endings are preserved.

| Key | Applied value |
|---|---|
| `NvidiaReflex@` | `Enabled` |
| `BloodLimit@` | `true` |
| `BloodLimitInterval@` | `2000` |
| `ShowBlood@` | `false` |
| `ShowBrass@` | `false` |
| `CorpseLimit@` | `0` |
| `GPUUploadHeaps@` | `false` |
| `PersistentDamageLayer@` | `false` |
| `SubdivisionLevel@` | `0` |
| `Tessellation@` | `0_Off` |
| `TerrainQuality@` | `Very Low` |
| `ShaderQuality@` | `Low` |
| `ModelQuality@` | `Low Quality` |
| `ParticleQuality@` | `very low` |
| `ShadowQuality@` | `Very_Low` |
| `VolumetricQuality@` | `QUALITY_LOW` |
| `AmbientLightingQuality@` | `Off` |
| `ScreenSpaceShadowQuality@` | `Off` |
| `SSRQuality@` | `Off` |
| `ReflectionProbeRelighting@` | `1` |
| `WorldStreamingQuality@` | `Low` |
| `WaterCausticsMode@` | `Off` |
| `WaterWaveWetness@` | `false` |
| `WeatherGridVolumesQuality@` | `Off` |
| `StaticSunshadowClipmapResolution@` | `0` |
| `DepthOfField@` | `false` |
| `DepthOfFieldQuality@` | `Low` |
| `EnableVelocityBasedBlur@` | `false` |
| `BulletImpacts@` | `false` |
| `CorpsesCullingThreshold@` | `0.500000` |
| `SkipIntro@` | `true` |
| `SkipSeasonIntroVideo@` | `true` |
| `SkipSeasonVideo@` | `true` |
| `ViewedSplashScreen@` | `true` |
| `RendererWorkerCount@` | `ceil(80% of P-cores or physical cores), maximum cores - 1` |

## Build from source

### Prerequisites

- Go 1.26 or newer;
- Windows to run the patcher;
- [mise](https://mise.jdx.dev/) is recommended for tooling and local tasks.

### Local build

```powershell
go mod download
go build -trimpath -ldflags="-s -w" -o cod-settings-patcher.exe .
```

### Windows AMD64 cross-compilation

The mise task builds a Windows AMD64 executable without CGO:

```bash
mise run build-windows-amd64
```

The binary is produced at:

```text
build/windows-amd64/cod-settings-patcher.exe
```

## Quality checks

```bash
gofmt -w .
go mod tidy
go mod verify
go vet ./...
go test -race -count=1 ./...
golangci-lint run --timeout=3m
```

To explicitly check the Windows target without running the binary:

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o build/windows-amd64/cod-settings-patcher.exe .
```

The race detector cannot run during macOS/Linux-to-Windows cross-compilation because the Windows test binary cannot run on the host.

## Releases

A tag in the `vX.Y.Z` format triggers the GitHub Actions release workflow. GoReleaser:

- builds `cod-settings-patcher.exe` for `windows/amd64` with `CGO_ENABLED=0`;
- creates a versioned ZIP archive containing the binary and README;
- generates SHA-256 `checksums.txt`;
- publishes or updates the matching GitHub Release.

Example:

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Current limitations

- The tool only searches for files located directly in `players` or `playersBeta`.
- Every `.txt`, `.txt0`, and `.txt1` file found in these directories is analyzed, but only the keys listed above can be modified.
- Rules are compiled into `settings.go` and are not yet configurable from an external file.
- SmartScreen may warn about unsigned binaries; only use artifacts published in the official Releases and verify `checksums.txt`.
