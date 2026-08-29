# Documentation avancée

Cette page regroupe les détails techniques, la liste des réglages et les instructions de développement de Call of Duty Settings Patcher.

## Détection des configurations

L’outil recherche les configurations sous :

```text
%LOCALAPPDATA%\Activision\Call of Duty
```

Il détecte les dossiers suivants :

- `players` pour les installations de jeu complet ;
- `playersBeta` pour les versions bêta.

Les fichiers directement présents dans ces dossiers avec les extensions `.txt`, `.txt0` et `.txt1` sont analysés. Les fichiers de sauvegarde contenant `.backup-` dans leur nom sont ignorés.

## Sécurité des écritures

Avant toute écriture, l’application :

1. calcule un plan de changements en mémoire ;
2. affiche le jeu, son dossier, les fichiers et chaque transition `ancienne valeur → nouvelle valeur` ;
3. attend une confirmation explicite avec `y` ;
4. relit les fichiers pour vérifier qu’ils n’ont pas changé depuis l’aperçu ;
5. crée une sauvegarde horodatée de chaque fichier ;
6. écrit via un fichier temporaire situé dans le même répertoire, puis le remplace de manière atomique ;
7. relit le fichier pour vérifier le contenu écrit ;
8. tente un rollback depuis les sauvegardes si une transaction multi-fichiers échoue.

Les sauvegardes sont nommées ainsi :

```text
<nom-du-fichier>.backup-YYYYMMDD-HHMMSS.nanosecondes
```

Le moteur travaille directement sur les octets des fichiers afin de préserver les commentaires `//`, les espaces et tabulations de fin de ligne, ainsi que les fins de ligne `LF`, `CRLF` et `CR`.

## RendererWorkerCount

`RendererWorkerCount@` est calculé à partir de la topologie CPU Windows, sans utiliser les threads logiques :

1. la topologie Windows est interrogée pour identifier les P-cores lorsqu’ils sont disponibles ;
2. sinon, le nombre de cœurs physiques est utilisé ;
3. les cœurs logiques, SMT et Hyper-Threading ne sont jamais utilisés ;
4. la valeur appliquée est `ceil(80 % du nombre de cœurs retenus)` ;
5. elle ne dépasse jamais `nombre_de_cœurs - 1`.

Exemples :

| Cœurs retenus | Calcul | Valeur |
|---:|---:|---:|
| 12 cœurs physiques | `ceil(12 × 0,80)` | `10` |
| 8 P-cores | `ceil(8 × 0,80)`, maximum `8 - 1` | `7` |
| 6 cœurs physiques | `ceil(6 × 0,80)` | `5` |
| 20 P-cores | `ceil(20 × 0,80)` | `16` |

## Réglages appliqués

Les clés sont recherchées sous leur forme `NomDeCle@... = valeur`. La portion après `@`, l’espacement autour de `=`, les commentaires éventuels et les fins de ligne sont conservés.

| Clé | Valeur appliquée |
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
| `RendererWorkerCount@` | `ceil(80 % des P-cores ou cœurs physiques), maximum cores - 1` |

## Compiler depuis les sources

### Prérequis

- Go 1.26 ou plus récent ;
- Windows pour l’exécution du patcher ;
- [mise](https://mise.jdx.dev/) est recommandé pour l’outillage et les tâches locales.

### Compilation locale

```powershell
go mod download
go build -trimpath -ldflags="-s -w" -o cod-settings-patcher.exe .
```

### Cross-compilation Windows AMD64

La tâche mise génère un exécutable Windows AMD64 sans CGO :

```bash
mise run build-windows-amd64
```

Le binaire est produit dans :

```text
build/windows-amd64/cod-settings-patcher.exe
```

## Vérifications qualité

```bash
gofmt -w .
go mod tidy
go mod verify
go vet ./...
go test -race -count=1 ./...
golangci-lint run --timeout=3m
```

Pour contrôler explicitement la cible Windows sans exécuter le binaire :

```bash
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go vet ./...
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o build/windows-amd64/cod-settings-patcher.exe .
```

Le détecteur de race ne peut pas être exécuté lors d’une cross-compilation macOS/Linux vers Windows, car le binaire de test Windows ne peut pas être lancé sur l’hôte.

## Releases

Un tag au format `vX.Y.Z` déclenche le workflow GitHub Actions de release. GoReleaser :

- compile `cod-settings-patcher.exe` pour `windows/amd64` avec `CGO_ENABLED=0` ;
- crée une archive ZIP versionnée contenant le binaire et le README ;
- génère `checksums.txt` en SHA-256 ;
- publie ou met à jour la GitHub Release correspondante.

Exemple :

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

## Limites actuelles

- L’outil ne cherche que les fichiers placés directement dans `players` ou `playersBeta`.
- Tous les fichiers `.txt`, `.txt0` et `.txt1` trouvés dans ces dossiers sont analysés, mais seules les clés listées ci-dessus peuvent être modifiées.
- Les règles sont compilées dans `settings.go` et ne sont pas encore configurables depuis un fichier externe.
- SmartScreen peut avertir pour les binaires non signés ; utilisez uniquement les artefacts publiés dans les Releases officielles et vérifiez `checksums.txt`.
