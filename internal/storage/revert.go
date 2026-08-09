package storage

import (
	"archive/tar"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"crds/internal/storage/db"

	"github.com/pressly/goose/v3"
)

// RevertReserve restores application state from a reserve copy archive. It
// first creates an automatic pre-revert backup, closes the current database
// connection, extracts the archive over the shared directory, re-opens the
// database, runs any pending migrations on the restored DB, and re-syncs
// YAML decks.
//
// The caller should verify the archive was created by CreateReserve (tar.gz
// containing at least crds.db).
func (s *Store) RevertReserve(sharedDir, reservePath string) error {
	if err := validateReserveArchive(reservePath); err != nil {
		return fmt.Errorf("revert: invalid archive: %w", err)
	}

	tmpDir, err := os.MkdirTemp("", "crds-revert-*")
	if err != nil {
		return fmt.Errorf("revert: temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	extracted, err := extractTarGz(reservePath, tmpDir)
	if err != nil {
		return fmt.Errorf("revert: extract: %w", err)
	}

	if err := s.CreateReserve(sharedDir); err != nil {
		return fmt.Errorf("revert: pre-backup: %w", err)
	}

	if err := s.conn.Close(); err != nil {
		return fmt.Errorf("revert: close db: %w", err)
	}

	for _, relPath := range extracted {
		src := filepath.Join(tmpDir, filepath.FromSlash(relPath))
		dst := filepath.Join(sharedDir, filepath.FromSlash(relPath))

		if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
			return fmt.Errorf("revert: mkdir %s: %w", filepath.Dir(dst), err)
		}

		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("revert: read %s: %w", relPath, err)
		}

		if err := os.WriteFile(dst, data, 0644); err != nil {
			return fmt.Errorf("revert: write %s: %w", relPath, err)
		}
	}

	conn, err := sql.Open("sqlite", s.dbPath+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return fmt.Errorf("revert: open db: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return fmt.Errorf("revert: ping db: %w", err)
	}

	goose.SetBaseFS(migrationsFS)
	_ = goose.SetDialect("sqlite")
	if err := goose.Up(conn, "migrations"); err != nil {
		return fmt.Errorf("revert: migrations: %w", err)
	}

	s.conn = conn
	s.queries = db.New(conn)
	s.mu = sync.Mutex{}
	s.currentSession = 0
	s.currentSessionAt = time.Time{}

	decksDir := filepath.Join(sharedDir, "decks")
	if err := s.SyncDecks(decksDir); err != nil {
		return fmt.Errorf("revert: sync: %w", err)
	}

	return nil
}

// validateReserveArchive checks that reservePath is a readable gzipped tar
// containing at least a crds.db entry.
func validateReserveArchive(path string) error {
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

// extractTarGz extracts a .tar.gz file into dir, returning the list of
// relative paths extracted.
func extractTarGz(path, dir string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)
	var extracted []string

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		target := filepath.Join(dir, filepath.FromSlash(hdr.Name))

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return nil, err
			}

		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return nil, err
			}
			out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fs.FileMode(hdr.Mode))
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(out, tr)
			closeErr := out.Close()
			if closeErr != nil {
				return nil, closeErr
			}
			if err != nil {
				return nil, err
			}
			extracted = append(extracted, hdr.Name)

		default:
			// skip symlinks, sockets, etc.
		}
	}

	return extracted, nil
}
