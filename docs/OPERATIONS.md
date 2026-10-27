# Operasional Lokal

## Proses aplikasi

Dari folder `app`:

```powershell
go run ./cmd/migrate
go run ./cmd/seed
go run ./cmd/server
```

Server lokal default: `http://127.0.0.1:8080`.

## Email, queue, dan scheduler

Email disimpan terlebih dahulu di tabel `notification_deliveries`. Jalankan worker secara berkala dari folder `app`:

```powershell
go run ./cmd/worker
```

`MAIL_DRIVER=log` hanya untuk lokal: worker menandai antrean terkirim tanpa mengirim isi email. Untuk server nyata, atur `MAIL_DRIVER=smtp` beserta `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, dan `SMTP_FROM`. Worker mengklaim satu pekerjaan secara transaksional, mencoba ulang kegagalan maksimal lima kali, dan mengambil kembali klaim yang macet lebih dari 15 menit.

Email booking publik dapat memiliki `user_id` kosong karena calon penghuni belum menjadi akun internal. Kolom ini sengaja nullable mulai migrasi 015; penerima tetap berasal dari email pengajuan booking dan tidak memperoleh akses portal. Akun tenant baru dibuat/dihubungkan owner setelah dana booking terverifikasi.

Jalankan scheduler sedikitnya sekali sehari:

```powershell
go run ./cmd/scheduler
```

Scheduler menandai invoice overdue, menandai kontrak yang mendekati akhir, memposting pengakuan pendapatan sewa yang jatuh tempo, serta membuat pengingat sebelum/saat/setelah jatuh tempo. Scheduler juga menyelesaikan reservasi fasilitas yang sudah lewat, mengubah izin tamu lama menjadi kedaluwarsa, mengingatkan kunjungan/reservasi hari ini, dan mengingatkan paket yang belum diambil pada hari ke-1 dan ke-3. Nilai default H-3, H+3, dan 30 hari dapat diubah owner di `/pemilik/pengaturan`; perubahan ini tidak membuat denda otomatis. Event email memakai kunci unik sehingga pengulangan scheduler tidak menggandakan kiriman. Scheduler turut membersihkan session kedaluwarsa, token reset/verifikasi lama, histori percobaan login/booking lebih dari 30 hari, dan bucket rate-limit lama. Di Windows Task Scheduler, arahkan `Start in` ke `D:\A. ALAM-SEGAR\app` dan gunakan executable Go hasil build untuk server live, bukan `go run`.

## Guard keamanan dan performa

- `APP_BASE_URL` harus berupa URL HTTP/HTTPS yang valid. Di luar mode `development`, aplikasi menolak start bila URL belum HTTPS, `APP_SECURE_COOKIE` belum `true`, atau `APP_HASH_KEY` kurang dari 32 karakter/masih memakai nilai lokal.
- Reverse proxy tetap wajib membatasi body/header, memasang sertifikat TLS, dan hanya meneruskan traffic ke bind address privat aplikasi. Server Go membatasi header 1 MiB, body 32 MiB, read/write/idle timeout, serta deadline handler 25 detik sebagai lapisan kedua.
- CSP memakai nonce acak per respons dan tidak mengizinkan inline JavaScript tanpa nonce. Jangan menambahkan atribut `onclick` atau script inline tanpa meneruskan `{{.CSPNonce}}`.
- Login memiliki limiter kredensial dan limiter IP. Reset password/registrasi maksimal 5 percobaan per IP per jam; booking dan bukti booking maksimal 20 per jam. `429` mengirim `Retry-After`. Aplikasi hanya membaca `X-Forwarded-For` bila peer langsung cocok dengan `APP_TRUSTED_PROXY_CIDRS`, lalu memvalidasi rantainya dari kanan ke kiri. Isi hanya CIDR reverse proxy yang benar; jangan pernah memakai CIDR publik luas.
- Static asset dan file publik tidak memuat session. Halaman ber-session memakai satu query join; write `last_seen_at` ditahan lima menit. Gzip hanya untuk payload teks, bukan gambar/PDF/upload yang sudah terkompresi.
- Jalankan `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` bersama test berkala. Per 26 Oktober 2026 hasil symbol/package aplikasi adalah nol kerentanan terpanggil. Baseline telah dinaikkan ke Go 1.26, driver MySQL `v1.10.1`, `x/crypto v0.56.0`, dan `edwards25519 v1.2.0`. Satu advisory level modul tetap dilaporkan untuk paket `x/crypto/openpgp` yang tidak diimpor aplikasi dan memang tidak memiliki versi perbaikan; jangan menambahkan pemakaian paket tersebut.

## Penyimpanan privat

Dokumen sensitif disimpan di `app/storage/private`. Direktori ini tidak dipetakan ke static file server; bukti pembayaran, refund, pengeluaran, inspeksi, dan keluhan hanya dilayani endpoint yang memeriksa role serta kepemilikan data. Foto katalog kamar disimpan terpisah di `app/storage/public/rooms` dan memang dapat diakses publik.

## PWA portal penghuni

Manifest tersedia di `/manifest.webmanifest` dan service worker di `/service-worker.js`. Android/Chrome menampilkan aksi pemasangan ketika browser mengizinkan; pada iOS/iPadOS penghuni memakai menu Bagikan → **Tambahkan ke Layar Utama**. Service worker hanya menyimpan shell statis `/assets/admin/` dan `/pwa/`. Navigasi selalu network-first, sedangkan data portal, kuitansi, dokumen privat, upload, serta respons POST tidak pernah disimpan oleh service worker. Halaman owner/tenant/auth juga mengirim `Cache-Control: no-store`.

Naikkan konstanta `CACHE_VERSION` di `app/static/pwa/service-worker.js` setiap kali aset shell yang sudah terpasang harus dipaksa diperbarui.

## Smart Kos dan dokumen digital

- Owner memproses tamu, fasilitas, paket, permintaan hunian, serta bantuan darurat melalui `/pemilik/smart-kos`; tenant memakai `/penghuni/layanan` atau shortcut PWA.
- Persetujuan permintaan perpanjangan/pindah/keluar tidak mengubah kontrak otomatis. Owner tetap menyelesaikan perubahan melalui modul Kontrak agar invoice, okupansi, dan ledger tidak terputus.
- Bantuan darurat hanya memberi notifikasi owner; UI secara eksplisit tidak mengaku sebagai pengganti layanan darurat pemerintah.
- Owner menerbitkan dan menandatangani snapshot di `/pemilik/dokumen`. Kontrak dan berita acara menunggu tanda tangan tenant; invoice/kuitansi selesai setelah tanda tangan owner.
- Tanda tangan memerlukan kata sandi aktif, terikat pada hash dokumen, actor, dan timestamp. Koreksi memakai pembatalan beralasan lalu versi baru; dokumen lama tidak ditimpa.
- URL `/verifikasi-dokumen/{kode}` hanya mengungkap nomor, jenis, versi, status, waktu, serta jejak hash—tanpa identitas tenant, kamar, isi, atau nominal.
- Mekanisme ini adalah persetujuan digital internal dengan audit trail. Integrasi penyelenggara tanda tangan elektronik tersertifikasi memerlukan provider dan kredensial live terpisah.

## Siklus booking sampai portal penghuni

1. Calon penghuni memilih kamar di `/kamar`, lalu mengirim `/booking`. Pengajuan belum menahan kamar.
2. Owner membuka `/pemilik/booking`, menyetujui nominal dan penggunaan uang booking. Sistem mengunci overlap, menahan kamar, dan mengantrekan email tautan pembayaran.
3. Pemohon membuka tautan `/booking/pembayaran`, memilih rekening, dan mengunggah bukti. Saldo dan ledger belum berubah.
4. Owner memeriksa mutasi bank lalu memverifikasi bukti. Baru saat itu bank didebit, kewajiban uang booking dikredit, dan booking menjadi terkonfirmasi.
5. Owner menekan **Buat & Kirim Akun Penghuni**. Email aktivasi 24 jam dikirim; pemohon memilih kata sandi dan sekaligus memverifikasi email.
6. Setelah akun aktif, owner menekan **Buat Kontrak dari Booking**. Form terisi otomatis; owner memeriksa syarat, membuat kontrak/tagihan, lalu melakukan check-in beserta inspeksi.
7. Penghuni masuk ke `/penghuni` dan dapat memasang PWA, melihat kamar/kontrak/tagihan, mengunggah pembayaran sewa, membaca pengumuman, dan membuat keluhan.

Jangan memberikan kata sandi sementara melalui chat. Jika email aktivasi kedaluwarsa, owner dapat menekan **Kirim Ulang Aktivasi Akun** pada booking terkonfirmasi.

## Backup database lokal

Contoh backup manual MariaDB/XAMPP:

```powershell
C:\database\mysql\bin\mysqldump.exe -h 127.0.0.1 -P 3306 -u root --single-transaction --routines --events --databases alam-segar > alam-segar-backup.sql
```

File privat harus dibackup terpisah bersama database. Restore wajib diuji pada database staging terpisah sebelum dipakai pada data nyata. Jangan menimpa database utama untuk sekadar menguji restore.

Contoh backup file aplikasi (PowerShell):

```powershell
Compress-Archive -Path 'D:\A. ALAM-SEGAR\app\storage\private','D:\A. ALAM-SEGAR\app\storage\public' -DestinationPath 'D:\backup\alam-segar-storage.zip'
```

Restore database ke nama staging, lalu verifikasi jumlah jurnal seimbang sebelum digunakan:

```powershell
C:\database\mysql\bin\mysql.exe -h 127.0.0.1 -P 3306 -u root -e "CREATE DATABASE alam_segar_restore CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci"
C:\database\mysql\bin\mysql.exe -h 127.0.0.1 -P 3306 -u root alam_segar_restore < alam-segar-backup.sql
```

Sesuaikan nama database pada dump bila memakai opsi `--databases`. Simpan backup terenkripsi dan batasi akses karena file privat berisi data tenant.

## Checklist deployment

- Ganti `APP_HASH_KEY`, password owner development, dan kredensial database; gunakan user database khusus dengan hak minimum.
- Set `APP_ENV=production`, `APP_BASE_URL` HTTPS, dan `APP_SECURE_COOKIE=true`.
- Konfigurasikan SMTP, scheduler harian, worker antrean, rotasi log, backup database, dan backup kedua folder storage.
- Pastikan reverse proxy membatasi ukuran request dan meneruskan HTTPS dengan benar.
- Jalankan `go test ./...`, `go vet ./...`, `go run golang.org/x/vuln/cmd/govulncheck@latest ./...`, `go run ./cmd/migrate`, lalu uji restore backup di staging.
- Jalankan `go run ./cmd/check-ledger`; perintah harus melaporkan total debit sama dengan kredit dan `jurnal_tidak_seimbang=0`.
- Jangan menaruh `storage/private`, `.env`, dump database, atau backup di direktori web publik.

## Catatan data demo

Perintah seed hanya membuat owner development secara default. Data properti/kamar contoh dibuat hanya bila `DEV_SEED_DEMO_DATA=true`; data berkode `DEMO` harus diganti atau dipisahkan sebelum penggunaan nyata.

