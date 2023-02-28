package inmemoryhashrepo

import (
	"errors"
	"sync"

	"github.com/ScienceSoft-Inc/integrity-sum/internal/model"
)

var ErrNotFound = errors.New("not found")

type InMemoryHashRepo struct {
	mtx    sync.RWMutex
	hashes map[string]string
}

func New() *InMemoryHashRepo {
	return &InMemoryHashRepo{
		mtx:    sync.RWMutex{},
		hashes: make(map[string]string),
	}
}

func (m *InMemoryHashRepo) Put(fh model.FileHash) {
	m.mtx.Lock()
	m.hashes[fh.Path] = fh.Hash
	m.mtx.Unlock()
}

func (m *InMemoryHashRepo) Get(path string) (model.FileHash, error) {
	m.mtx.RLock()
	h, ok := m.hashes[path]
	m.mtx.RUnlock()
	if ok {
		return model.FileHash{}, ErrNotFound
	}
	return model.FileHash{Path: path, Hash: h}, nil
}
