package model

type ErrorResponse struct {
	Error string `json:"error"`
}

type QuotaResponse struct {
	Sisa    int    `json:"sisa"`
	Kantor  string `json:"kantor"`
	Tanggal string `json:"tanggal"`
}

type ReservasiRequest struct {
	Name          string `json:"nama"`
	NIK           string `json:"nik"`
	Kantor        string `json:"kantor"`
	Tanggal       string `json:"tanggal"`
	JenisPassport string `json:"jenis_paspor"`
	MasaBerlaku   string `json:"masa_berlaku"`
	Percepatan    bool   `json:"percepatan"`
}

type ReservasiResponse struct {
	KodeBooking     string  `json:"kode_booking"`
	KodeBilling     string  `json:"kode_billing"`
	Nama            string  `json:"nama"`
	Kantor          string  `json:"kantor"`
	Tanggal         string  `json:"tanggal"`
	JenisPassport   string  `json:"jenis_paspor"`
	MasaBerlaku     string  `json:"masa_berlaku"`
	Percepatan      bool    `json:"percepatan"`
	BiayaDasar      float64 `json:"biaya_dasar"`
	BiayaPercepatan float64 `json:"biaya_percepatan"`
	Total           float64 `json:"total"`
	RincianBiaya    string  `json:"rincian_biaya"`
}

type UploadResponse struct {
	StoredName   string `json:"stored_name"`
	OriginalName string `json:"original_name"`
	ContentType  string `json:"content_type"`
	Size         int64  `json:"size"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Nama     string `json:"nama"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ResetPasswordRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	Username string `json:"username"`
	Nama     string `json:"nama"`
}
