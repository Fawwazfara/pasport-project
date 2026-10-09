# Reservasi Paspor — Prototype Penelitian

Prototype sistem reservasi paspor (terinspirasi M-Paspor) yang dibuat sebagai bahan
penelitian skripsi perbandingan performa cloud (AWS / GCP / Azure). **Yang diuji adalah
REST API-nya**; UI web hanya pelengkap untuk memperagakan alur.

> Prototype penelitian, bukan layanan resmi. Semua data bersifat dummy.

- Backend: Go (standard library `net/http` + `ServeMux`), tanpa framework besar.
- UI: HTML + CSS + JavaScript biasa (tanpa build step), di-embed ke binary dengan `go:embed`.
- Satu `Dockerfile` multi-stage + `docker-compose.yml`; `docker compose up` menjalankan semuanya.

## Alur layanan

1. Pilih kantor dan tanggal kunjungan, lihat sisa kuota (tanpa reload).
2. Isi data diri (nama, NIK, jenis paspor, masa berlaku, percepatan).
3. Unggah dokumen: KTP, Kartu Keluarga, dan satu dokumen pendukung.
4. Tampilkan ringkasan + kode booking + kode billing (pembayaran hanya simulasi).

## Endpoint

| Method | Path             | Keterangan                                                        |
| ------ | ---------------- | ----------------------------------------------------------------- |
| GET    | `/api/kuota`     | Sisa kuota. Query: `kantor`, `tanggal` (YYYY-MM-DD). Read-heavy.  |
| POST   | `/api/reservasi` | Buat reservasi (JSON), kurangi kuota atomik, balikan kode & biaya.|
| POST   | `/api/upload`    | Unggah dokumen (multipart). Max 2 MB, jpg/png/pdf.                |
| GET    | `/health`        | Health check.                                                     |

Semua error berformat konsisten `{"error": "pesan"}` dengan status code yang tepat
(400 validasi, 404 tidak ditemukan, 405 method salah, 409 kuota habis, 500 internal).

### Aturan domain (biaya)

| Jenis paspor | Masa berlaku | Biaya  |
| ------------ | ------------ | ------ |
| Biasa        | 5 tahun      | Rp350.000 |
| Biasa        | 10 tahun     | Rp650.000 |
| Elektronik   | 5 tahun      | Rp650.000 |
| Elektronik   | 10 tahun     | Rp950.000 |

Percepatan (selesai hari yang sama) menambah **Rp1.000.000** di luar biaya penerbitan.

## Arsitektur

Logika endpoint tidak bergantung pada implementasi penyimpanan. Dua interface dipisah per package:

- `internal/quota_repository` — `QuotaRepository` (`GetQuota`, `Reserve`). Pengurangan kuota
  dilakukan secara **atomik** di `Reserve` (hasil: sukses atau `ErrQuotaExhausted`), sehingga
  tidak ada double booking. Implementasi saat ini: **in-memory** (dilindungi `sync.Mutex`).
- `internal/file_storage` — `FileStorage` (`Save`). `Save` mengembalikan metadata hasil simpan
  (`SavedFile`: nama, ukuran, content type). Implementasi saat ini: **disk lokal**, nama file
  diacak (hex 16 byte), bukan dari input pengguna.

Pemilihan implementasi lewat environment variable, jadi penggantian ke DynamoDB/Firestore/
Cosmos DB dan S3/GCS/Blob cukup menambah implementasi baru tanpa mengubah handler.

```
cmd/server/main.go        # routing + handler HTTP
internal/model/           # struct request/response JSON
internal/quota_repository # interface + implementasi in-memory + test konkurensi
internal/file_storage     # interface + implementasi disk lokal
web/                      # index.html, app.js, style.css (di-embed)
```

## Konfigurasi (environment variable)

| Variabel          | Default                                                         | Keterangan                       |
| ----------------- | --------------------------------------------------------------- | -------------------------------- |
| `PORT`            | `8080`                                                          | Port server.                     |
| `STORAGE_DIR`     | `./storage`                                                     | Direktori penyimpanan file.      |
| `QUOTA_BACKEND`   | `inmemory`                                                      | Backend kuota (saat ini hanya ini). |
| `STORAGE_BACKEND` | `local`                                                         | Backend file (saat ini hanya ini).  |
| `INITIAL_QUOTAS`  | `Bandung:50,Jakarta Selatan:50,Surabaya:50,Yogyakarta:50`       | Kuota awal per kantor per tanggal, format `Kantor:Jumlah,...`. |

## Menjalankan

### Docker (satu perintah)

```bash
docker compose up --build
```

UI tersedia di `http://localhost:8081` (host 8081 dipetakan ke container 8080 agar tidak
bentrok dengan layanan lain di 8080). Storage dipersistensikan lewat named volume.

### Lokal

```bash
go run ./cmd/server
# atau
PORT=8080 STORAGE_DIR=./storage go run ./cmd/server
```

Buka `http://localhost:8080`.

## UI

Satu halaman wizard 4 langkah berbahasa Indonesia (pilih kantor & jadwal → data pemohon →
unggah dokumen → ringkasan). Kuota ditampilkan tanpa reload, error validasi muncul di bawah
field, tombol menampilkan status loading, dan ada tombol cetak ringkasan. Tampilan netral
ala layanan pemerintah dengan satu warna aksen, tetap ada label "Prototype penelitian,
bukan layanan resmi". Aset (`index.html`, `style.css`, `app.js`) di-embed ke binary.

## Contoh curl

```bash
# Health
curl http://localhost:8080/health

# 1) Sisa kuota
curl "http://localhost:8080/api/kuota?kantor=Bandung&tanggal=2026-12-01"

# 2) Reservasi
curl -X POST http://localhost:8080/api/reservasi \
  -H 'Content-Type: application/json' \
  -d '{"nama":"Budi","nik":"1234567890123456","kantor":"Bandung","tanggal":"2026-12-01","jenis_paspor":"elektronik","masa_berlaku":"10","percepatan":true}'

# 3) Unggah dokumen (kode_booking dari respons reservasi)
curl -X POST http://localhost:8080/api/upload \
  -F kode_booking=5BD252CBE3D7 \
  -F jenis_dokumen=ktp \
  -F file=@ktp.jpg
```

Contoh respons reservasi:

```json
{
  "kode_booking": "5BD252CBE3D7",
  "kode_billing": "766C623E",
  "nama": "Budi",
  "kantor": "Bandung",
  "tanggal": "2026-12-01",
  "jenis_paspor": "elektronik",
  "masa_berlaku": "10",
  "percepatan": true,
  "biaya_dasar": 950000,
  "biaya_percepatan": 1000000,
  "total": 1950000,
  "rincian_biaya": "Biaya penerbitan biaya dasar Rp 950000 + percepatan Rp 1000000"
}
```

## Pengujian

```bash
go vet ./...
go test ./...
```

`internal/quota_repository/inmemory_test.go` memuat test konkurensi: kuota 50, **200
request reservasi bersamaan -> tepat 50 sukses dan 150 ditolak** (diuji dengan race-safe
counter), serta test kuota per-tanggal dan kantor tidak valid.

## Build & CI (Jenkins)

Untuk kemudahan build dan latihan Jenkins tersedia:

- `Makefile` — `make vet`, `make test` (dengan `-race`), `make build`, `make run`, `make docker`.
- `Jenkinsfile` — pipeline deklaratif: `Checkout` → `Verify` (`go vet` + `go test -race`)
  → `Build Binary` → `Build Image` (`docker build`). Agent memakai image `golang:1.22`;
  tahap `Build Image` butuh agent dengan Docker. Registry/push belum disertakan.

## Asumsi yang diambil

- Kuota bersifat per **kantor + tanggal** dan tidak "reset" harian otomatis; nilai awal
  diambil dari `INITIAL_QUOTAS`. Karena in-memory, kuota kembali ke nilai awal ketika
  proses/server di-restart (belum ada persistensi kuota).
- Validasi "tanggal tidak boleh lewat" memakai zona waktu lokal server (`time.Local`).
- `kode_booking` dan `kode_billing` adalah string acak dummy (`math/rand`) untuk keperluan
  simulasi, bukan identitas transaksi yang aman. Pembayaran murni simulasi.
- `/api/upload` sengaja tidak memverifikasi keberadaan `kode_booking` (belum ada store
  reservasi pada fase ini).
- Deteksi tipe file memakai `http.DetectContentType` terhadap byte awal file, sehingga
  isi file diperiksa dan bukan hanya ekstensi.
- Atomicity kuota dijamin dalam satu proses (mutex). Jika nanti dijalankan multi-instance,
  jaminan harus disediakan oleh backend terdistribusi (mis. conditional write DynamoDB).

## Yang sengaja belum dikerjakan

- Integrasi pembayaran nyata (kode billing hanya dummy).
- Persistensi kuota & metadata reservasi ke database (DynamoDB/Firestore/Cosmos DB).
- Implementasi penyimpanan file ke object storage (S3/GCS/Blob) — interface sudah siap.
- Autentikasi/login dan halaman admin.
- Verifikasi `kode_booking` saat unggah, pemindaian antivirus, dan deduplikasi dokumen.
- Validasi NIK terhadap Dukcapil dan aturan penggantian paspor lama.
- TLS/HTTPS (diasumsikan diterminasi di reverse proxy).
