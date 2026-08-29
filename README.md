# Call of Duty Settings Patcher

Application Windows en Go avec interface terminal **Bubble Tea** pour détecter des fichiers de configuration Call of Duty et appliquer un ensemble de réglages de manière contrôlée.

## Garanties

- Détection sous `%LOCALAPPDATA%\\Activision\\Call of Duty`.
- Prise en charge des dossiers `players` (jeu complet) et `playersBeta` (bêta).
- Détection des fichiers `.txt`, `.txt0` et `.txt1` dans ces dossiers.
- Sélection explicite de l'installation dans la TUI.
- Aperçu obligatoire du jeu, du dossier, des fichiers et de chaque changement `ancienne valeur → nouvelle valeur`.
- Confirmation explicite (`y`) avant toute écriture.
- Backup unique et horodaté de chaque fichier modifié : `fichier.backup-YYYYMMDD-HHMMSS.nanosecondes`.
- Vérification que le fichier n'a pas été modifié entre l'aperçu et la confirmation.
- Écriture atomique dans le même répertoire, avec vérification après écriture.
- Tentative de rollback depuis les backups si une écriture multi-fichiers échoue.
- Préservation byte à byte des commentaires `//`, des espaces et tabulations de fin, et des fins de lignes `LF`, `CRLF` ou `CR`.

## Settings appliqués

| Clé | Valeur |
|---|---|
| `NvidiaReflex@` | `Enabled` |
| `BloodLimit@` | `true` |
| `BloodLimitInterval@` | `2000` |
| `ShowBlood@` | `false` |
| `ShowBrass@` | `false` |
| `CorpseLimit@` | `0` |
| `GPUUploadHeaps@` | `false` |
| `PersistentDamageLayer@` | `false` |
| `SubDivisionLevel@` | `0` |
| `WaterCausticMode@` | `Off` |
| `WaterWaveWetness@` | `false` |
| `WeatherGridVolumesQuality@` | `Off` |
| `BulletImpacts@` | `false` |
| `CorpsesCullingThreshold@` | `0.500000` |
| `TerrainQuality@` | `Very Low` |
| `Tesselation@` | `0_Off` |
| `RendererWorkerCount@` | `10` |

Les parties variables après `@` sont préservées. Exemple :

```text
RendererWorkerCount@a1b2c3 = 12   // configuration CPU
```

est modifié en :

```text
RendererWorkerCount@a1b2c3 = 10   // configuration CPU
```

## Prérequis

- Go 1.23 ou plus récent.
- Windows avec `%LOCALAPPDATA%` défini.
- Le jeu doit être fermé pendant l'application des changements.

## Construire et lancer

```powershell
go mod download
go test -race ./...
go vet ./...
go build -trimpath -ldflags="-s -w" -o cod-settings-patcher.exe .
.\cod-settings-patcher.exe
```

## Utilisation

1. Lancez l'exécutable dans un terminal Windows.
2. Sélectionnez l'installation détectée avec `↑`/`↓` ou `j`/`k`.
3. Appuyez sur `Entrée` pour calculer l'aperçu.
4. Vérifiez le dossier, les fichiers et toutes les transitions de valeurs.
5. Appuyez sur `Entrée` ou `y` pour accéder à la confirmation.
6. Appuyez sur `y` pour créer les backups et appliquer les changements.

La touche `n`, `b` ou `Échap` annule/revient en arrière selon l'écran. `q` quitte l'application.

## Validation qualité

```powershell
gofmt -w .
go vet ./...
go test -race -count=1 ./...
```

`golangci-lint` est configuré dans `.golangci.yml` et exécuté dans GitHub Actions.

## Limites actuelles

- L'outil ne détecte que les fichiers présents directement dans `players` ou `playersBeta`.
- Tous les fichiers avec extension `.txt`, `.txt0` ou `.txt1` de ces dossiers sont analysés ; seuls les settings listés ci-dessus peuvent être modifiés.
- Les règles de mapping sont compilées dans `settings.go` et ne sont pas encore configurables par fichier externe.
