# Call of Duty Settings Patcher

A Windows terminal tool for easily applying selected Call of Duty settings to your configuration files.

The application detects your configurations, shows every change before it is applied, and automatically creates a backup of each modified file.

## Download

Download the latest Windows archive from the [Releases](../../releases) page, extract it to any folder, then run `cod-settings-patcher.exe`.

## Usage

1. **Close Call of Duty** before running the tool.
2. Open a terminal in the folder containing the executable.
3. Run:

   ```powershell
   .\cod-settings-patcher.exe
   ```

4. Select the detected installation with `↑`/`↓` or `j`/`k`.
5. Press `Enter` to review the proposed changes.
6. Check the game, directory, files, and values that will be changed.
7. Confirm with `y` to apply the changes, or cancel with `n`/`Esc`.

The tool does not create new settings: it only changes keys that are already present in your configuration files.

## Backups

Before making any changes, the tool creates a timestamped copy of each file next to the original. If anything goes wrong, simply restore the matching `.backup-...` file.

## SmartScreen warning

Windows may show a **Microsoft Defender SmartScreen** warning for an executable downloaded from the Internet that is unsigned or has not yet established a reputation.

Only bypass this warning if you downloaded the archive from this repository's official [Releases](../../releases) page and, ideally, verified its SHA-256 checksum using `checksums.txt`.

## Advanced documentation

For the complete list of settings, details about `RendererWorkerCount`, backups, building from source, and technical checks, see [ADVANCED.md](ADVANCED.md).
