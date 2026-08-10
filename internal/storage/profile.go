package storage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"crds/internal/storage/db"

	"github.com/pressly/goose/v3"
)

// CreateProfile packs everything needed for device migration into a .tar.gz
// archive: crds.db, state.yaml, decks/*.yaml, config.yaml, keymaps.yaml,
// themes/*.yaml, and mappings/*.yaml from the config directory.
//
// The archive is written to outputDir with the base name crds-profile.tar.gz.
// If that file already exists, -1, -2, etc. are tried. Pass a non-empty name
// to override the base name (.tar.gz appended if missing).
// Returns the full path of the created archive.
func (s *Store) CreateProfile(sharedDir, configDir, outputDir, name string) (string, error) {
	if name == "" {
		name = "crds-profile.tar.gz"
	} else if !strings.HasSuffix(name, ".tar.gz") {
		name += ".tar.gz"
	}

	profilePath := filepath.Join(outputDir, name)
	if _, err := os.Stat(profilePath); err == nil {
		base := strings.TrimSuffix(name, ".tar.gz")
		for i := 1; ; i++ {
			candidate := filepath.Join(outputDir, base+"-"+strconv.Itoa(i)+".tar.gz")
			if _, err := os.Stat(candidate); os.IsNotExist(err) {
				profilePath = candidate
				break
			}
		}
	}

	if err := s.createProfileArchive(sharedDir, configDir, profilePath); err != nil {
		return "", err
	}
	return profilePath, nil
}

// createProfileArchive checkpoints the WAL and writes a compressed tar archive.
func (s *Store) createProfileArchive(sharedDir, configDir, profilePath string) error {
	ctx := context.Background()
	if _, err := s.conn.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		return fmt.Errorf("profile: checkpoint: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(profilePath), 0755); err != nil {
		return fmt.Errorf("profile: mkdir: %w", err)
	}

	f, err := os.Create(profilePath)
	if err != nil {
		return fmt.Errorf("profile: create %s: %w", profilePath, err)
	}
	defer func() { _ = f.Close() }()

	gzw := gzip.NewWriter(f)
	defer func() { _ = gzw.Close() }()

	tw := tar.NewWriter(gzw)
	defer func() { _ = tw.Close() }()

	for _, name := range []string{"state.yaml", "crds.db"} {
		if err := addFileToTar(tw, sharedDir, name); err != nil {
			return fmt.Errorf("profile: add %s: %w", name, err)
		}
	}

	if err := addDeckDirToTar(tw, sharedDir); err != nil {
		return fmt.Errorf("profile: add decks: %w", err)
	}

	if err := addConfigDirToTar(tw, configDir); err != nil {
		return fmt.Errorf("profile: add config: %w", err)
	}

	return nil
}

// addConfigDirToTar adds config.yaml, keymaps.yaml, themes/*.yaml, and
// mappings/*.yaml from configDir into the tar under a config/ prefix.
func addConfigDirToTar(tw *tar.Writer, configDir string) error {
	// config.yaml
	if err := addFileWithTarName(tw, filepath.Join(configDir, "config.yaml"), "config/config.yaml"); err != nil {
		return err
	}
	// keymaps.yaml
	if err := addFileWithTarName(tw, filepath.Join(configDir, "keymaps.yaml"), "config/keymaps.yaml"); err != nil {
		return err
	}
	// themes/*.yaml
	themesDir := filepath.Join(configDir, "themes")
	entries, err := os.ReadDir(themesDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
			if err := addFileWithTarName(tw, filepath.Join(themesDir, e.Name()), filepath.Join("config", "themes", e.Name())); err != nil {
				return err
			}
		}
	}

	// mappings/*.yaml (per-language input mappings)
	mappingsDir := filepath.Join(configDir, "mappings")
	mentries, err := os.ReadDir(mappingsDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	for _, e := range mentries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".yaml") || strings.HasSuffix(e.Name(), ".yml") {
			if err := addFileWithTarName(tw, filepath.Join(mappingsDir, e.Name()), filepath.Join("config", "mappings", e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

// addFileWithTarName adds a single file to the tar writer with an explicit
// tar-internal path. Silently skips if the file does not exist.
func addFileWithTarName(tw *tar.Writer, filePath, tarName string) error {
	fi, err := os.Stat(filePath)
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
	header.Name = filepath.ToSlash(tarName)
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if !fi.Mode().IsRegular() {
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	_, err = io.Copy(tw, f)
	return err
}

// ImportProfile restores application state from a profile archive. It
// first creates an automatic pre-restore backup, closes the current database
// connection, extracts the archive (sharedDir files go to sharedDir, config
// files go to configDir), re-opens the database, runs pending migrations,
// and re-syncs YAML decks.
func (s *Store) ImportProfile(sharedDir, configDir, profilePath string) error {
	if err := validateProfileArchive(profilePath); err != nil {
		return fmt.Errorf("profile-import: invalid archive: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "crds-profile-*")
	if err != nil {
		return fmt.Errorf("profile-import: temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	extracted, err := extractTarGz(profilePath, tmpDir)
	if err != nil {
		return fmt.Errorf("profile-import: extract: %w", err)
	}

	if err := s.CreateReserve(sharedDir); err != nil {
		return fmt.Errorf("profile-import: pre-backup: %w", err)
	}

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("profile-import: close db: %w", err)
	}

	for _, relPath := range extracted {
		src := filepath.Join(tmpDir, filepath.FromSlash(relPath))

		var dst string
		switch {
		case strings.HasPrefix(relPath, "config/"):
			sub := strings.TrimPrefix(relPath, "config/")
			dst = filepath.Join(configDir, filepath.FromSlash(sub))
		default:
			dst = filepath.Join(sharedDir, filepath.FromSlash(relPath))
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("profile-import: mkdir %s: %w", filepath.Dir(dst), err)
		}

		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("profile-import: read %s: %w", relPath, err)
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("profile-import: write %s: %w", relPath, err)
		}
	}

	conn, err := sql.Open("sqlite", s.dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("profile-import: open db: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("profile-import: ping db: %w", err)
	}

	goose.SetBaseFS(migrationsFS)
	_ = goose.SetDialect("sqlite")
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("profile-import: migrations: %w", err)
	}

	s.conn = conn
	s.queries = db.New(conn)
	s.mu = sync.Mutex{}
	s.currentSession = 0
	s.currentSessionAt = time.Time{}

	decksDir := filepath.Join(sharedDir, "decks")
	if err := s.SyncDecks(decksDir); err != nil {
		return fmt.Errorf("profile-import: sync: %w", err)
	}

	return nil
}

// validateProfileArchive checks that profilePath is a readable gzipped tar
// containing at least a crds.db entry.
func validateProfileArchive(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	hasDB := false
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		if hdr.Name == "crds.db" {
			hasDB = true
			break
		}
	}

	if !hasDB {
		return fmt.Errorf("archive does not contain crds.db")
	}
	return nil
}
