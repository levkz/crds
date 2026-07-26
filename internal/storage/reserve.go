package storage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CreateReserve creates a compressed backup of the database, state file, and
// all deck YAML files into ~/.local/share/crds/reserve-copies/.
//
// The backup file is named crds-rsv-{increment}-{ddMMyyyy}-{mmss}.tar.gz.
// Increment is auto-derived from existing backups.
func (s *Store) CreateReserve(sharedDir string) error {
	ctx := context.Background()

	// Flush WAL for a clean DB snapshot
	_, err := s.conn.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
	if err != nil {
		return fmt.Errorf("reserve: checkpoint: %w", err)
	}

	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	if err := os.MkdirAll(reserveDir, 0755); err != nil {
		return fmt.Errorf("reserve: mkdir: %w", err)
	}

	increment, err := nextReserveIncrement(reserveDir)
	if err != nil {
		return fmt.Errorf("reserve: increment: %w", err)
	}

	now := time.Now()
	filename := fmt.Sprintf("crds-rsv-%03d-%s-%s.tar.gz",
		increment,
		now.Format("02012006"),
		now.Format("150405"),
	)
	reservePath := filepath.Join(reserveDir, filename)

	f, err := os.Create(reservePath)
	if err != nil {
		return fmt.Errorf("reserve: create %s: %w", reservePath, err)
	}
	defer f.Close()

	gzw := gzip.NewWriter(f)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	for _, name := range []string{"state.yaml", "crds.db"} {
		if err := addFileToTar(tw, sharedDir, name); err != nil {
			return fmt.Errorf("reserve: add %s: %w", name, err)
		}
	}

	if err := addDeckDirToTar(tw, sharedDir); err != nil {
		return fmt.Errorf("reserve: add decks: %w", err)
	}

	return nil
}

func addFileToTar(tw *tar.Writer, baseDir, relPath string) error {
	fullPath := filepath.Join(baseDir, relPath)

	fi, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return err
	}
	header.Name = filepath.ToSlash(relPath)
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if !fi.Mode().IsRegular() {
		return nil
	}

	f, err := os.Open(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(tw, f)
	return err
}

func addDeckDirToTar(tw *tar.Writer, sharedDir string) error {
	decksDir := filepath.Join(sharedDir, "decks")

	entries, err := os.ReadDir(decksDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".yaml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		relPath := filepath.Join("decks", name)
		if err := addFileToTar(tw, sharedDir, relPath); err != nil {
			return err
		}
	}

	return nil
}

func nextReserveIncrement(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return 1, nil
	}
	if err != nil {
		return 0, err
	}

	maxInc := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasPrefix(e.Name(), "crds-rsv-") {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".tar.gz")
		parts := strings.Split(name, "-")
		if len(parts) < 3 {
			continue
		}
		inc, err := strconv.Atoi(parts[2])
		if err != nil {
			continue
		}
		if inc > maxInc {
			maxInc = inc
		}
	}

	return maxInc + 1, nil
}
