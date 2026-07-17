package navigation

import ui "crds/internal/ui"

type registryEntry struct {
	screen  ui.Screen
	factory func() ui.Screen
}

type Registry struct {
	entries map[ui.ScreenIndex]registryEntry
}

func NewRegistry() *Registry {
	return &Registry{entries: make(map[ui.ScreenIndex]registryEntry)}
}

func (r *Registry) Register(index ui.ScreenIndex, screen ui.Screen) {
	r.entries[index] = registryEntry{screen: screen}
}

func (r *Registry) RegisterFactory(index ui.ScreenIndex, factory func() ui.Screen) {
	r.entries[index] = registryEntry{factory: factory}
}

func (r *Registry) Get(index ui.ScreenIndex) (ui.Screen, bool) {
	entry, ok := r.entries[index]
	if !ok {
		return nil, false
	}
	if entry.screen != nil {
		return entry.screen, true
	}
	if entry.factory != nil {
		entry.screen = entry.factory()
		r.entries[index] = entry
		return entry.screen, true
	}
	return nil, false
}

func (r *Registry) Has(index ui.ScreenIndex) bool {
	_, ok := r.entries[index]
	return ok
}

func (r *Registry) Remove(index ui.ScreenIndex) {
	delete(r.entries, index)
}

func (r *Registry) Len() int {
	return len(r.entries)
}
