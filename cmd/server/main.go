package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"passport/internal/file_storage"
	"passport/internal/model"
	"passport/internal/quota_repository"
	"passport/web"
)

const (
	envQuotaBackend   = "QUOTA_BACKEND"
	envStorageBackend = "STORAGE_BACKEND"
	envInitialQuotas  = "INITIAL_QUOTAS"
	envStorageDir     = "STORAGE_DIR"
	defaultQuotas     = "Bandung:50,Jakarta Selatan:50,Surabaya:50,Yogyakarta:50"
	maxFileSize       = 2 * 1024 * 1024
)

var jenisPassportBiaya = map[string]float64{
	"biasa5":       350000,
	"biasa10":      650000,
	"elektronik5":  650000,
	"elektronik10": 950000,
}

var allowedFileTypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"application/pdf": true,
}

func randomCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = "0123456789ABCDEF"[rand.Intn(16)]
	}
	return string(b)
}

func allDigits(s string) bool {
	for i := range s {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return len(s) > 0
}

func formatRupiah(n float64) string {
	return "Rp " + strconv.FormatFloat(n, 'f', 0, 64)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(model.ErrorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func parseQuotas(s string) map[string]int {
	result := make(map[string]int)
	for _, part := range strings.Split(s, ",") {
		kantor, jumlah, ok := strings.Cut(part, ":")
		kantor = strings.TrimSpace(kantor)
		if !ok || kantor == "" {
			log.Fatalf("INITIAL_QUOTAS tidak valid: %q (format: Kantor:Jumlah,...)", part)
		}
		n, err := strconv.Atoi(strings.TrimSpace(jumlah))
		if err != nil || n < 0 {
			log.Fatalf("INITIAL_QUOTAS tidak valid: %q (jumlah harus angka >= 0)", part)
		}
		result[kantor] = n
	}
	return result
}

type server struct {
	repo  quota_repository.QuotaRepository
	store file_storage.FileStorage
}

func (s *server) kuota(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	kantor := r.URL.Query().Get("kantor")
	tanggal := r.URL.Query().Get("tanggal")

	if kantor == "" || tanggal == "" {
		writeError(w, http.StatusBadRequest, "parameter kantor dan tanggal wajib")
		return
	}
	if _, err := time.Parse("2006-01-02", tanggal); err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid (format: YYYY-MM-DD)")
		return
	}

	sisa, err := s.repo.GetQuota(kantor, tanggal)
	if err != nil {
		if errors.Is(err, quota_repository.ErrKantorInvalid) {
			writeError(w, http.StatusBadRequest, "kantor tidak valid")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, model.QuotaResponse{Sisa: sisa, Kantor: kantor, Tanggal: tanggal})
}

func (s *server) reservasi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	var req model.ReservasiRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if len(req.NIK) != 16 || !allDigits(req.NIK) {
		writeError(w, http.StatusBadRequest, "NIK harus 16 digit angka")
		return
	}

	tgl, err := time.ParseInLocation("2006-01-02", req.Tanggal, time.Local)
	if err != nil {
		writeError(w, http.StatusBadRequest, "tanggal tidak valid (format: YYYY-MM-DD)")
		return
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	if tgl.Before(today) {
		writeError(w, http.StatusBadRequest, "tanggal tidak boleh lewat")
		return
	}

	if req.JenisPassport != "biasa" && req.JenisPassport != "elektronik" {
		writeError(w, http.StatusBadRequest, "jenis paspor tidak valid")
		return
	}
	if req.MasaBerlaku != "5" && req.MasaBerlaku != "10" {
		writeError(w, http.StatusBadRequest, "masa berlaku harus 5 atau 10")
		return
	}

	biayaDasar := jenisPassportBiaya[req.JenisPassport+req.MasaBerlaku]
	biayaPercepatan := 0.0
	rincian := "Biaya penerbitan "
	if req.Percepatan {
		biayaPercepatan = 1000000
		rincian += "biaya dasar " + formatRupiah(biayaDasar) + " + percepatan " + formatRupiah(biayaPercepatan)
	} else {
		rincian += "biaya dasar " + formatRupiah(biayaDasar)
	}

	err = s.repo.Reserve(req.Kantor, req.Tanggal, quota_repository.ReservationData{
		Name:          req.Name,
		NIK:           req.NIK,
		Kantor:        req.Kantor,
		Tanggal:       req.Tanggal,
		JenisPassport: req.JenisPassport,
		MasaBerlaku:   req.MasaBerlaku,
		Percepatan:    req.Percepatan,
	})
	if err != nil {
		switch {
		case errors.Is(err, quota_repository.ErrKantorInvalid):
			writeError(w, http.StatusBadRequest, "kantor tidak valid")
		case errors.Is(err, quota_repository.ErrQuotaExhausted):
			writeError(w, http.StatusConflict, "kuota habis untuk tanggal "+req.Tanggal+" di kantor "+req.Kantor)
		default:
			writeError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	writeJSON(w, model.ReservasiResponse{
		KodeBooking:     randomCode(12),
		KodeBilling:     randomCode(8),
		Nama:            req.Name,
		Kantor:          req.Kantor,
		Tanggal:         req.Tanggal,
		JenisPassport:   req.JenisPassport,
		MasaBerlaku:     req.MasaBerlaku,
		Percepatan:      req.Percepatan,
		BiayaDasar:      biayaDasar,
		BiayaPercepatan: biayaPercepatan,
		Total:           biayaDasar + biayaPercepatan,
		RincianBiaya:    rincian,
	})
}

func (s *server) upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		writeError(w, http.StatusBadRequest, "gagal parse form: "+err.Error())
		return
	}

	kodeBooking := r.FormValue("kode_booking")
	jenisDokumen := r.FormValue("jenis_dokumen")

	if kodeBooking == "" {
		writeError(w, http.StatusBadRequest, "kode_booking wajib")
		return
	}
	if jenisDokumen != "ktp" && jenisDokumen != "kk" && jenisDokumen != "dokumen_pendukung" {
		writeError(w, http.StatusBadRequest, "jenis dokumen tidak valid")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file wajib")
		return
	}
	defer file.Close()

	isi, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal baca file")
		return
	}

	if int64(len(isi)) > maxFileSize {
		writeError(w, http.StatusBadRequest, "file terlalu besar (max 2MB)")
		return
	}

	detectedType := http.DetectContentType(isi)
	if !allowedFileTypes[detectedType] {
		writeError(w, http.StatusBadRequest, "jenis file tidak diizinkan (hanya jpg, png, pdf)")
		return
	}

	saved, err := s.store.Save(header.Filename, isi, detectedType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal simpan file: "+err.Error())
		return
	}

	writeJSON(w, model.UploadResponse{
		StoredName:   saved.Name,
		OriginalName: header.Filename,
		ContentType:  saved.ContentType,
		Size:         saved.Size,
	})
}

func (s *server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method tidak diizinkan")
		return
	}
	writeJSON(w, map[string]string{"status": "ok"})
}

func newQuotaRepo(initialQuotas map[string]int) quota_repository.QuotaRepository {
	switch os.Getenv(envQuotaBackend) {
	case "", "inmemory":
		return quota_repository.NewQuotaRepository(initialQuotas)
	default:
		log.Fatalf("%s tidak dikenal: %s (tersedia: inmemory)", envQuotaBackend, os.Getenv(envQuotaBackend))
		return nil
	}
}

func newFileStore(dir string) file_storage.FileStorage {
	switch os.Getenv(envStorageBackend) {
	case "", "local":
		return file_storage.NewLocalStorage(dir)
	default:
		log.Fatalf("%s tidak dikenal: %s (tersedia: local)", envStorageBackend, os.Getenv(envStorageBackend))
		return nil
	}
}

func main() {
	quotasStr := os.Getenv(envInitialQuotas)
	if quotasStr == "" {
		quotasStr = defaultQuotas
	}
	initialQuotas := parseQuotas(quotasStr)

	storageDir := os.Getenv(envStorageDir)
	if storageDir == "" {
		storageDir = "./storage"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	s := &server{
		repo:  newQuotaRepo(initialQuotas),
		store: newFileStore(storageDir),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/kuota", s.kuota)
	mux.HandleFunc("/api/reservasi", s.reservasi)
	mux.HandleFunc("/api/upload", s.upload)
	mux.HandleFunc("/health", s.health)

	static := http.FileServer(http.FS(web.Files))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "endpoint tidak ditemukan")
			return
		}
		static.ServeHTTP(w, r)
	})

	log.Printf("passport reserve berjalan di port %s", port)
	log.Printf("storage: %s", storageDir)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
