package quota_repository

import "errors"

var (
	ErrQuotaExhausted = errors.New("kuota habis")
	ErrKantorInvalid  = errors.New("kantor tidak valid")
)

type ReservationData struct {
	Name          string
	NIK           string
	Kantor        string
	Tanggal       string
	JenisPassport string
	MasaBerlaku   string
	Percepatan    bool
}

type QuotaRepository interface {
	GetQuota(kantor string, tanggal string) (int, error)
	Reserve(kantor string, tanggal string, data ReservationData) error
}
