package app

import "crds/internal/storage"

type App struct {
	Store     *storage.Store
	State     *storage.StateStore
	SharedDir string
	DataDir   string
}
