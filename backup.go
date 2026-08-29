// Package main provides a terminal application for safely patching Call of Duty settings files.
package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func applyPlan(plan ChangePlan) ([]string, error) {
	if !plan.HasChanges() {
		return nil, nil
	}

	for _, filePlan := range plan.Files {
		current, err := os.ReadFile(filePlan.Path) // #nosec G304 -- path originates from discovered game settings files.
		if err != nil {
			return nil, fmt.Errorf("relecture de %q : %w", filePlan.Path, err)
		}
		if !bytes.Equal(current, filePlan.Before) {
			return nil, fmt.Errorf("%q a changé depuis l'aperçu ; annulation pour éviter un écrasement", filePlan.Path)
		}
	}

	backups := make([]string, 0, len(plan.Files))
	for _, filePlan := range plan.Files {
		info, err := os.Stat(filePlan.Path)
		if err != nil {
			return backups, fmt.Errorf("stat de %q : %w", filePlan.Path, err)
		}

		backup, err := createTimestampedBackup(filePlan.Path, filePlan.Before, info.Mode())
		if err != nil {
			return backups, err
		}
		backups = append(backups, backup)
	}

	for index, filePlan := range plan.Files {
		info, err := os.Stat(filePlan.Path)
		if err != nil {
			return backups, rollbackApplied(plan.Files[:index], backups[:index], err)
		}
		if err := atomicWrite(filePlan.Path, filePlan.After, info.Mode()); err != nil {
			return backups, rollbackApplied(plan.Files[:index], backups[:index], err)
		}
		if err := verifyWrittenFile(filePlan.Path, filePlan.After); err != nil {
			return backups, rollbackApplied(plan.Files[:index+1], backups[:index+1], err)
		}
	}

	return backups, nil
}

func createTimestampedBackup(source string, content []byte, mode fs.FileMode) (string, error) {
	timestamp := time.Now().Format("20060102-150405.000000000")
	backupPath := source + ".backup-" + timestamp

	if err := writeNewFile(backupPath, content, mode.Perm()); err != nil {
		return "", fmt.Errorf("création du backup %q : %w", backupPath, err)
	}

	return backupPath, nil
}

func writeNewFile(path string, content []byte, permissions fs.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, permissions) // #nosec G304 -- destination is derived from a discovered settings file and cannot overwrite an existing file.
	if err != nil {
		return err
	}

	defer func() { _ = file.Close() }()

	if _, err := file.Write(content); err != nil {
		return err
	}
	return file.Sync()
}

func atomicWrite(path string, content []byte, mode fs.FileMode) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), ".cod-settings-patcher-*")
	if err != nil {
		return fmt.Errorf("création du fichier temporaire : %w", err)
	}

	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()

	if err := temporary.Chmod(mode.Perm()); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("chmod du fichier temporaire : %w", err)
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("écriture temporaire : %w", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("synchronisation temporaire : %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("fermeture temporaire : %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("remplacement de %q : %w", path, err)
	}

	committed = true
	return nil
}

func verifyWrittenFile(path string, expected []byte) error {
	actual, err := os.ReadFile(path) // #nosec G304 -- path originates from a discovered settings file written by atomicWrite.
	if err != nil {
		return fmt.Errorf("vérification de %q : %w", path, err)
	}
	if sha256.Sum256(actual) != sha256.Sum256(expected) {
		return fmt.Errorf("vérification échouée pour %q", path)
	}
	return nil
}

func rollbackApplied(files []FileChangePlan, backups []string, originalErr error) error {
	var rollbackErrs []error
	for i := len(files) - 1; i >= 0; i-- {
		backupContent, err := os.ReadFile(backups[i]) // #nosec G304 -- backup path was created by createTimestampedBackup in this transaction.
		if err != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("lecture du backup %q : %w", backups[i], err))
			continue
		}
		info, err := os.Stat(files[i].Path)
		if err != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("stat pendant rollback %q : %w", files[i].Path, err))
			continue
		}
		if err := atomicWrite(files[i].Path, backupContent, info.Mode()); err != nil {
			rollbackErrs = append(rollbackErrs, fmt.Errorf("rollback de %q : %w", files[i].Path, err))
		}
	}
	if len(rollbackErrs) > 0 {
		return fmt.Errorf("%w ; rollback incomplet : %v", originalErr, rollbackErrs)
	}
	return fmt.Errorf("%w ; les fichiers déjà modifiés ont été restaurés", originalErr)
}
