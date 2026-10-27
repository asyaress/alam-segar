# Sistem Manajemen Kos Alam Segar

Aplikasi Go server-rendered yang menggunakan aset visual asli berikut tanpa mengubah folder sumbernya:

- `../package` — Travl untuk panel pemilik dan portal penghuni.
- `../hotel-booking-boostrap-5-template-seaside-2025-08-25-10-30-06-utc` — Seaside untuk halaman publik; landing mengikuti struktur `Seaside HTML/02-index-full-page.html`.

Backend, validasi, otorisasi, dan akses MySQL dijalankan oleh Go. PHP handler bawaan template tidak digunakan.

Runtime minimum proyek adalah Go 1.26.

## Menjalankan lokal

1. Pastikan MariaDB berjalan dan database `alam-segar` tersedia.
2. Salin `.env.example` menjadi `.env`, lalu sesuaikan kredensial lokal bila perlu.
3. Dari folder `app`, jalankan:

```powershell
go mod download
go run ./cmd/migrate
go run ./cmd/seed
go run ./cmd/server
```

Buka `http://127.0.0.1:8080`. Akun seed hanya boleh dipakai untuk pengembangan lokal dan harus diganti sebelum deployment.

Portal owner tersedia di `/pemilik` dan portal penghuni di `/penghuni`. Portal penghuni memiliki manifest/service worker PWA; pemasangan hanya ditawarkan browser pada konteks aman (HTTPS atau localhost). Layanan smart-living tersedia di `/pemilik/smart-kos` dan `/penghuni/layanan`, sedangkan arsip serta tanda tangan dokumen digital tersedia di `/pemilik/dokumen` dan `/penghuni/dokumen`.

## Membuat atau mereset akun owner

Gunakan environment variable agar kata sandi tidak disimpan di source code:

```powershell
$env:OWNER_NAME='Nama Pemilik'
$env:OWNER_EMAIL='pemilik@example.com'
$env:OWNER_PASSWORD='Kata-Sandi-Kuat-123!'
go run ./cmd/create-owner
```

Perintah ini membuat akun owner aktif atau mereset akun owner yang memakai email yang sama, mencabut session lama, dan mencatat perubahan pada audit log. Kata sandi lemah ditolak. Pengecualian `ALLOW_WEAK_DEV_PASSWORD=true` hanya tersedia pada environment development untuk pengujian lokal dan tidak boleh dipakai saat deployment.

## Batas keamanan penting

- File privat berada di `storage/private` dan tidak pernah dilayani sebagai static directory.
- Hanya subdirektori aset CSS/JS/gambar/font template yang dilayani; file PHP demo/template tidak diekspos.
- Saldo, pembayaran, dan jurnal berasal dari ledger double-entry dan alur finansial yang telah diuji lokal. Sebelum go-live tetap wajib menjalankan audit, uji restore backup, mengganti seluruh rahasia lokal, dan mengikuti `../docs/OPERATIONS.md`.
- Service worker hanya meng-cache aset shell. Halaman owner/tenant/auth memakai `Cache-Control: no-store`; dokumen privat, kuitansi, dan data finansial tidak dicache untuk penggunaan offline.
- Dokumen digital menyimpan snapshot dan hash, mewajibkan reautentikasi kata sandi saat tanda tangan, serta menyediakan verifikasi publik tanpa PII. Ini adalah persetujuan digital internal, bukan klaim tanda tangan elektronik tersertifikasi pihak ketiga.
- Semua respons memakai CSP nonce per request, proteksi framing/MIME/referrer/cross-origin, batas body 32 MiB, deadline request, panic recovery, dan rate limit persisten untuk login, reset password, registrasi, booking, serta unggah booking.
- Aset tidak membaca session dari database. Session login dimuat dengan satu query dan `last_seen` maksimal ditulis sekali per lima menit. Respons HTML/CSS/JS/SVG memakai gzip bila didukung browser.

## Proses pendamping

```powershell
go run ./cmd/scheduler  # harian: workflow bisnis serta cleanup session/token/rate-limit
go run ./cmd/worker     # proses antrean email
go run ./cmd/check-ledger # audit keseimbangan setiap jurnal posted
```

Lihat `../docs/OPERATIONS.md` untuk SMTP, scheduler Windows, backup/restore, dan checklist deployment.

Pengaturan jadwal reminder dan tanggal mulai penggunaan tersedia di `http://127.0.0.1:8080/pemilik/pengaturan`. Pengembalian dana, terminasi dini, pembelian aset, dan pengeluaran pemeliharaan selalu memakai alur manual berbukti dan jurnal; tidak ada integrasi bank otomatis.

