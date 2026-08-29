package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// GameInstallation describes one detected Call of Duty configuration directory.
type GameInstallation struct {
	Name       string
	Root       string
	PlayersDir string
	Variant    string
	Files      []string
}

var acceptedConfigExtensions = map[string]struct{}{
	".txt":  {},
	".txt0": {},
	".txt1": {},
}

func findCODRoot() (string, error) {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return "", errors.New("variable d'environnement LOCALAPPDATA absente")
	}

	root := filepath.Join(localAppData, "Activision", "Call of Duty")
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("dossier Call of Duty introuvable : %q", root)
		}
		return "", fmt.Errorf("stat de %q : %w", root, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%q n'est pas un répertoire", root)
	}

	return root, nil
}

func discoverInstallations(root string) ([]GameInstallation, error) {
	var installations []GameInstallation

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			return nil
		}

		name := entry.Name()
		if !strings.EqualFold(name, "players") && !strings.EqualFold(name, "playersBeta") {
			return nil
		}

		files, err := findSettingsFiles(path)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return filepath.SkipDir
		}

		variant := "Jeu complet"
		if strings.EqualFold(name, "playersBeta") {
			variant = "Bêta"
		}

		gameRoot := filepath.Dir(path)
		installations = append(installations, GameInstallation{
			Name:       filepath.Base(gameRoot),
			Root:       gameRoot,
			PlayersDir: path,
			Variant:    variant,
			Files:      files,
		})

		return filepath.SkipDir
	})
	if err != nil {
		return nil, fmt.Errorf("scan de %q : %w", root, err)
	}

	sort.Slice(installations, func(i, j int) bool {
		return strings.ToLower(installations[i].Name) < strings.ToLower(installations[j].Name)
	})

	return installations, nil
}

func findSettingsFiles(playersDir string) ([]string, error) {
	entries, err := os.ReadDir(playersDir)
	if err != nil {
		return nil, fmt.Errorf("lecture de %q : %w", playersDir, err)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.Contains(strings.ToLower(name), ".backup-") {
			continue
		}

		if _, ok := acceptedConfigExtensions[strings.ToLower(filepath.Ext(name))]; !ok {
			continue
		}

		files = append(files, filepath.Join(playersDir, name))
	}

	sort.Strings(files)
	return files, nil
}
