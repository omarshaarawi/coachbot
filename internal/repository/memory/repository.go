package memory

import (
	"sync"

	"github.com/omarshaarawi/coachbot/internal/models"
)

type Repository struct {
	metadata *models.LeagueMetadata
	mu       sync.RWMutex
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) SaveMetadata(metadata *models.LeagueMetadata) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metadata = metadata
}

func (r *Repository) GetMetadata() *models.LeagueMetadata {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.metadata
}
