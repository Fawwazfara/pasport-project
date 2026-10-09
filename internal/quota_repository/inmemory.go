package quota_repository

import "sync"

type quotaRepository struct {
	mu      sync.Mutex
	initial map[string]int
	quotas  map[string]int
}

func NewQuotaRepository(initialQuotas map[string]int) *quotaRepository {
	return &quotaRepository{
		initial: initialQuotas,
		quotas:  make(map[string]int),
	}
}

func quotaKey(kantor, tanggal string) string {
	return kantor + "|" + tanggal
}

func (r *quotaRepository) GetQuota(kantor string, tanggal string) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	awal, ok := r.initial[kantor]
	if !ok {
		return 0, ErrKantorInvalid
	}
	if sisa, ok := r.quotas[quotaKey(kantor, tanggal)]; ok {
		return sisa, nil
	}
	return awal, nil
}

func (r *quotaRepository) Reserve(kantor string, tanggal string, data ReservationData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	awal, ok := r.initial[kantor]
	if !ok {
		return ErrKantorInvalid
	}

	key := quotaKey(kantor, tanggal)
	sisa, ada := r.quotas[key]
	if !ada {
		sisa = awal
	}
	if sisa <= 0 {
		return ErrQuotaExhausted
	}
	r.quotas[key] = sisa - 1
	return nil
}
