package quota_repository

import (
	"sync"
	"testing"
)

func TestReserveConcurrent(t *testing.T) {
	repo := NewQuotaRepository(map[string]int{
		"Bandung": 50,
	})

	var mu sync.Mutex
	success := 0
	rejected := 0
	var wg sync.WaitGroup

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := repo.Reserve("Bandung", "2026-01-15", ReservationData{
				Name:        "Test",
				NIK:         "1234567890123456",
				Kantor:      "Bandung",
				Tanggal:     "2026-01-15",
				MasaBerlaku: "5",
			})
			mu.Lock()
			switch {
			case err == nil:
				success++
			case err == ErrQuotaExhausted:
				rejected++
			}
			mu.Unlock()
		}()
	}

	wg.Wait()

	if success != 50 {
		t.Errorf("expected 50 successful reservations, got %d", success)
	}
	if rejected != 150 {
		t.Errorf("expected 150 rejected reservations, got %d", rejected)
	}

	sisa, err := repo.GetQuota("Bandung", "2026-01-15")
	if err != nil {
		t.Fatalf("GetQuota: %v", err)
	}
	if sisa != 0 {
		t.Errorf("expected sisa kuota 0, got %d", sisa)
	}
}

func TestQuotaPerTanggal(t *testing.T) {
	repo := NewQuotaRepository(map[string]int{
		"Bandung": 1,
	})

	if err := repo.Reserve("Bandung", "2026-01-15", ReservationData{}); err != nil {
		t.Fatalf("reserve pertama seharusnya sukses: %v", err)
	}
	if err := repo.Reserve("Bandung", "2026-01-15", ReservationData{}); err != ErrQuotaExhausted {
		t.Errorf("reserve kedua seharusnya ErrQuotaExhausted, got %v", err)
	}

	sisa, err := repo.GetQuota("Bandung", "2026-01-16")
	if err != nil {
		t.Fatalf("GetQuota tanggal lain: %v", err)
	}
	if sisa != 1 {
		t.Errorf("tanggal lain seharusnya masih 1, got %d", sisa)
	}
}

func TestKantorInvalid(t *testing.T) {
	repo := NewQuotaRepository(map[string]int{
		"Bandung": 10,
	})

	if _, err := repo.GetQuota("Atlantis", "2026-01-15"); err != ErrKantorInvalid {
		t.Errorf("expected ErrKantorInvalid, got %v", err)
	}
	if err := repo.Reserve("Atlantis", "2026-01-15", ReservationData{}); err != ErrKantorInvalid {
		t.Errorf("expected ErrKantorInvalid, got %v", err)
	}
}
