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

## RendererWorkerCount

`RendererWorkerCount@` est calculé au lancement à partir de la topologie CPU Windows, sans utiliser le nombre de threads logiques :

1. Le programme interroge PowerShell/CIM (`Win32_Processor`).
2. Lorsque Windows expose `NumberOfPerformanceCores`, ce nombre de P-cores est utilisé.
3. Sinon, il utilise `NumberOfCores`, le nombre de cores physiques.
4. La valeur appliquée est l'arrondi mathématique de `80 %` de ce nombre, avec un minimum de `1`.

Exemples :

| Cores utilisés | Calcul | `RendererWorkerCount` |
|---:|---:|---:|
| 12 cores physiques | `round(12 × 0,80)` | `10` |
| 8 P-cores | `round(8 × 0,80)` | `6` |
| 6 cores physiques | `round(6 × 0,80)` | `5` |
| 20 P-cores | `round(20 × 0,80)` | `16` |

Le récapitulatif TUI affiche la valeur effectivement planifiée. Les processeurs sans cœurs hybrides utilisent automatiquement le fallback `NumberOfCores` ; les cœurs logiques/SMT/Hyper-Threading ne sont jamais pris en compte.

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
| `WaterCausticMode@` | `Off` |
| `WaterWaveWetness@` | `false` |
| `WeatherGridVolumesQuality@` | `Off` |
| `BulletImpacts@` | `false` |
| `CorpsesCullingThreshold@` | `0.500000` |
| `TerrainQuality@` | `Very Low` |
| `Tesselation@` | `0_Off` |
| `RendererWorkerCount@` | `round(80 % des P-cores, ou des cores physiques)` |

Les parties variables après `@` sont préservées. Exemple :

```text
RendererWorkerCount@a1b2c3 = 12   // configuration CPU
```

est modifié, sur un processeur de 12 cores physiques, en :

```text
RendererWorkerCount@a1b2c3 = 10   // configuration CPU
```

## Prérequis

- Go 1.23 ou plus récent.
- Windows avec `%LOCALAPPDATA%` défini.
- PowerShell et la classe CIM `Win32_Processor` disponibles.
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
