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
	_, err := s.CreateReserveTo(sharedDir, "", "")
	return err
}

// CreateReserveTo creates a reserve archive. outputDir defaults to
// sharedDir/reserve-copies/. name defaults to an auto-generated filename.
// If name does not end in .tar.gz the extension is appended.
// Returns the full path of the created archive.
func (s *Store) CreateReserveTo(sharedDir, outputDir, name string) (string, error) {
	if outputDir == "" {
		outputDir = filepath.Join(sharedDir, "reserve-copies")
	}
	if name == "" {
		inc, err := nextReserveIncrement(outputDir)
		if err != nil {
			return "", fmt.Errorf("reserve: increment: %w", err)
		}
		now := time.Now()
		name = fmt.Sprintf("crds-rsv-%03d-%s-%s.tar.gz",
			inc,
			now.Format("02012006"),
			now.Format("150405"),
		)
	} else if !strings.HasSuffix(name, ".tar.gz") {
		name += ".tar.gz"
	}

	reservePath := filepath.Join(outputDir, name)
	if err := s.createReserveArchive(sharedDir, reservePath); err != nil {
		return "", err
	}
	return reservePath, nil
}

// createReserveArchive checkpoints the WAL and writes a compressed tar archive
// at reservePath containing state.yaml, crds.db, and decks/*.yaml.
func (s *Store) createReserveArchive(sharedDir, reservePath string) error {
	ctx := context.Background()
	if _, err := s.conn.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("reserve: checkpoint: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(reservePath), 0755); err != nil {
		return fmt.Errorf("reserve: mkdir: %w", err)
	}

	f, err := os.Create(reservePath)
	if err != nil {
		return fmt.Errorf("reserve: create %s: %w", reservePath, err)
	}
	defer func() { _ = f.Close() }()

	gzw := gzip.NewWriter(f)
	defer func() { _ = gzw.Close() }()

	tw := tar.NewWriter(gzw)
	defer func() { _ = tw.Close() }()

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

// ListReserves returns full paths of reserve archives in the default reserve
// directory, sorted newest-first (by filename which embeds an increment).
func ListReserves(sharedDir string) ([]string, error) {
	reserveDir := filepath.Join(sharedDir, "reserve-copies")
	entries, err := os.ReadDir(reserveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".tar.gz") && strings.HasPrefix(e.Name(), "crds-rsv-") {
			files = append(files, filepath.Join(reserveDir, e.Name()))
		}
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))
	return files, nil
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
	defer func() { _ = f.Close() }()

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
