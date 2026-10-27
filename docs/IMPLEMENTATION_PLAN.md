# Rencana Implementasi Sistem Manajemen Kos

## Keputusan arsitektur

- Backend: satu aplikasi Go server-rendered, bukan microservice dan bukan frontend PHP terpisah.
- Database: MariaDB/MySQL database `alam-segar` yang sudah dibuat pengguna.
- UI publik: markup dan aset Seaside.
- UI pemilik/penghuni: markup, partial, kelas CSS, dan aset Travl.
- Folder template sumber diperlakukan read-only. Adaptasi dilakukan hanya pada view aplikasi Go.
- Timestamps disimpan UTC; tanggal bisnis menggunakan `DATE` dan ditampilkan dalam `Asia/Makassar`.
- Nilai uang disimpan sebagai integer rupiah (`BIGINT`), bukan floating point.

## Pemetaan reuse template

| Kebutuhan | Sumber template |
|---|---|
| Landing page | `Seaside HTML/index.html` |
| Katalog kamar | `room.html` dan `room-2-cols.html` |
| Detail kamar | `room-single.html` |
| Pengajuan booking | `booking.html` |
| Login | `package/page-login.php` |
| Dashboard pemilik | `package/index.php` |
| Daftar kamar | `package/room-list.php` |
| Penghuni | `package/guest-list.php` dan `guest-details.php` |
| Tagihan/kuitansi | `package/ecom-invoice.php` dan komponen tabel/form |

## Backlog bertahap

- [x] Audit spesifikasi, template, runtime, dan database.
- [x] Validasi visual template desktop 1440px dan mobile 390px.
- [ ] Fondasi aplikasi: konfigurasi, migrasi, auth dua role, session, CSRF, audit log. *(konfigurasi/auth/session/CSRF/audit dasar sudah ada; registrasi dan email verification belum)*
- [ ] Properti, tipe kamar, kamar, fasilitas, foto, dan katalog publik. *(CRUD dasar, katalog, dan detail sudah ada; upload foto/fasilitas belum)*
- [ ] Registrasi tenant dan approval owner.
- [ ] Booking, manual hold, bukti transfer, verifikasi, refund manual, konversi kontrak. *(pengajuan, approval/reject, lock bentrok, dan idempotency sudah ada; pembayaran/refund/konversi belum)*
- [ ] Kontrak bulanan/tahunan, installment preset, check-in/out, amendment.
- [ ] Invoice, pembayaran, alokasi, kuitansi, credit balance, reminder.
- [ ] Ledger double-entry, kas/bank, saldo awal, expense wajib bukti, close book.
- [ ] Inventaris, inspeksi, kerusakan, keluhan, dan pemeliharaan.
- [ ] Laporan, email queue, scheduler, export, backup, aksesibilitas, dan QA.

Setiap kotak hanya boleh ditandai selesai setelah route, validation, authorization, UI, dan test terkait benar-benar berjalan.

