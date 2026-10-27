# Status Implementasi

Tanggal audit terakhir: 26 Oktober 2026 (Asia/Makassar).

## Sudah berfungsi dan diuji

- Aplikasi Go server-rendered terhubung ke MariaDB database `alam-segar`.
- Folder template Travl dan Seaside tetap menjadi sumber aset read-only; file PHP handler template tidak dieksekusi atau dipublikasikan.
- Landing page memakai struktur asli Seaside `02-index-full-page.html` (video lokal, info bar, kartu kamar, blok lifestyle, alur booking, footer), sedangkan katalog/detail/booking tetap memakai aset Seaside tanpa mengubah folder sumber.
- Login owner/tenant berbasis database, bcrypt, session server-side, rotasi session setelah login, CSRF, secure cookie setting, login rate limit, dan role guard.
- CRUD tambah/ubah/daftar properti dan kamar untuk owner; perubahan tarif kamar mempunyai histori terpisah. Fasilitas dan foto kamar dikelola owner dan ditampilkan pada katalog Seaside.
- PLN/PDAM kamar disimpan sebagai metadata privat dan tidak dibuat sebagai invoice.
- Booking publik memiliki validasi, rate limit, CSRF, nomor referensi, dan idempotency key agar double-submit tidak membuat dua data.
- Owner dapat menyetujui booking dengan nominal dan keputusan penggunaan uang booking, atau menolak dengan alasan. Pemohon publik menerima email transaksional persetujuan/penolakan, bukti ditolak/terverifikasi, hold dilepas, dan refund tanpa harus menjadi role internal.
- Setelah dana booking diverifikasi, owner dapat membuat atau menghubungkan akun tenant dari data pemohon. Akun baru mendapat tautan aktivasi 24 jam, verifikasi email terjadi saat calon penghuni memilih kata sandi, dan formulir kontrak memuat tenant/kamar/periode/tarif booking secara otomatis.
- Persetujuan booking menggunakan transaksi serta lock kamar; booking/kontrak yang overlap memblokir persetujuan berikutnya.
- Dashboard memakai hitungan database dan empty state, bukan angka template. Dashboard owner menampilkan grafik ApexCharts enam bulan serta donut status kamar dari ledger/data hunian; tabel owner memakai DataTables dengan pencarian, sort, pagination, dan bahasa Indonesia.
- Uji otomatis kebijakan kata sandi, anchor tanggal 29/31, dan pembagian angsuran deterministik.
- Registrasi tenant, verifikasi email bertoken, approval/rejection owner, login tenant, lupa/reset kata sandi, dan pencabutan sesi lama.
- Uang booking dengan unggah bukti, verifikasi manual owner, hold/release, penggunaan per booking, konversi kontrak, serta refund manual berbukti.
- Kontrak bulanan/tahunan, angsuran tahunan, invoice, due date ber-anchor, extension/version history, check-in/out dan foto inspeksi.
- Verifikasi pembayaran manual, partial payment sesuai kontrak, alokasi, kredit tenant, pemakaian maupun refund kredit tenant, kuitansi, void, dan refund berbukti.
- Ledger double-entry idempoten, COA, multi-bank, kas kecil owner, saldo awal ke ekuitas, transfer internal, expense wajib bukti, rekonsiliasi, pengakuan pendapatan, dan tutup buku.
- Terminasi dini menyimpan alasan/versi/inspeksi, membatalkan jadwal serta piutang masa depan, dan memproses keputusan refund atau retensi dana secara manual dengan jurnal.
- Laporan laba rugi, neraca, aging piutang, arus kas langsung, kas/bank, dan jurnal berbasis ledger dengan filter tanggal/properti serta ekspor CSV.
- Inventaris/aset, tagihan kerusakan, keluhan dua arah, status/vendor, pengeluaran yang dapat ditautkan ke tiket/aset, pembelian aset tetap, pengumuman tenant, notifikasi, email queue/worker, dan scheduler idempoten.
- Pengaturan owner untuk nama usaha, tanggal go-live, H-minus/H-plus reminder, dan ambang kontrak akan berakhir; scheduler membaca setting tersebut.
- Tanggal bisnis disimpan sebagai UTC-midnight untuk kolom `DATE`; migrasi korektif memperbaiki data lokal lama yang sempat bergeser satu hari akibat konversi WITA→UTC.
- Otorisasi file privat owner/tenant. PLN/PDAM pribadi tetap metadata dan tidak masuk invoice maupun ledger kos.
- Form tambah aset, pengeluaran, pengumuman, rekening, kas, saldo awal, transfer internal, dan rekonsiliasi menggunakan modal; tabel data tetap memenuhi lebar halaman.
- Portal tenant menjadi PWA installable: manifest, service worker shell-only, fallback offline, shortcut aplikasi, meta Android/iOS, indikator offline, bottom navigation ponsel, serta halaman kamar rinci (foto, fasilitas, kontrak, utilitas pribadi, inventaris, dan peraturan). Halaman/data privat dan finansial tidak dimasukkan ke cache PWA.
- Modul Smart Kos end-to-end: izin tamu dengan persetujuan dan kode akses, guard waktu check-in, reservasi fasilitas dengan kapasitas serta pencegahan bentrok, penerimaan/konfirmasi paket, permintaan perpanjangan-pindah-keluar, bantuan darurat, audit trail, notifikasi website/email, dashboard tindakan, dan scheduler idempoten.
- Pusat Dokumen Digital untuk invoice, kuitansi, kontrak, dan berita acara inspeksi: snapshot immutable, versi koreksi, reautentikasi owner/tenant, HMAC integritas, pembatalan beralasan, cetak/simpan PDF dari browser, dan tautan verifikasi publik tanpa PII. Tanda tangan ini adalah persetujuan digital internal dan tidak diklaim sebagai TTE tersertifikasi eksternal.
- Hardening web menyeluruh: CSP nonce tanpa `unsafe-inline` untuk script, anti-framing/MIME/referrer/cross-origin headers, HSTS saat HTTPS, request/body/header timeout dan limit, panic recovery, gzip selektif, trusted-proxy CIDR yang fail-closed, serta rate-limit persisten IP untuk seluruh endpoint publik sensitif.
- Optimasi session mengeluarkan static asset/PWA/upload/health dari query database, menggabungkan session+user menjadi satu query, dan menahan write `last_seen_at` menjadi maksimal sekali per lima menit. Scheduler membersihkan session, token, attempts, dan bucket limiter kedaluwarsa.
- Smoke test autentikasi seluruh halaman owner dan tenant, serta E2E registrasi→verifikasi→approval→login, reset password, overpayment→kredit→refund, terminasi dini→refund→pengakuan pendapatan, dan biaya pemeliharaan terkait tiket/aset.
- Migrasi aktif berjumlah 19. Uji booking tambahan membuktikan approval dan release memasukkan email calon penghuni ke queue; uji booking berbayar membuktikan verifikasi dana, posting ledger, provisioning akun, serta antrean aktivasi.
- Uji E2E Smart Kos membuktikan alur tenant→owner→tenant, penolakan slot fasilitas bentrok (`409`), larangan check-in tamu/penyelesaian reservasi sebelum waktunya, paket diambil, bantuan diselesaikan, serta scheduler dua kali tanpa duplikasi reminder.
- Uji E2E dokumen membuktikan kontrak dua tanda tangan, invoice cukup tanda tangan owner, kata sandi salah ditolak `401`, akses dokumen tenant lain ditolak `404`, dokumen void dapat diterbitkan ulang sebagai versi baru, dan halaman verifikasi tidak memuat nama/nominal/kamar.
- Regresi hardening membuktikan seluruh route publik/owner/tenant tetap aktif, CSP nonce header cocok dengan HTML, halaman privat `no-store`, limiter mengembalikan `429` dan `Retry-After` pada percobaan keenam, gzip mengurangi landing 9.383 menjadi 2.478 byte (73,6%), serta burst lokal 100 request selesai tanpa error. Runtime/dependency telah diperbarui ke Go 1.26, driver MySQL `v1.10.1`, `x/crypto v0.56.0`, dan `edwards25519 v1.2.0`; `govulncheck` melaporkan nol kerentanan symbol/package yang dipakai aplikasi.

## Integrasi eksternal yang belum dites

- SMTP nyata belum diuji karena kredensial SMTP belum disediakan; driver lokal `log` sudah diuji tanpa mencetak token rahasia.
- Belum dilakukan penetration test oleh pihak eksternal atau pengujian visual manual pada seluruh kombinasi HP kecil/foldable/browser. Smoke load lokal sudah dijalankan, tetapi capacity test pada spesifikasi server final tetap perlu dilakukan. Landing, dashboard/chart, DataTables/modal, dashboard tenant, dan detail kamar sudah diperiksa visual di browser lokal desktop.
- Backup dan restore sudah didokumentasikan, tetapi restore berkala tetap harus dijalankan oleh operator pada lingkungan staging.
- Data berkode `DEMO`/`E2E` dan akun development berpassword lemah wajib diganti/dihapus sebelum penggunaan nyata. Kredensial lokal tidak didokumentasikan di repository publik.

Core workflow aplikasi sudah tersambung end-to-end dan lulus test/lint/smoke test lokal. Checklist go-live di `docs/OPERATIONS.md` tetap wajib.

