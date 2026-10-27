# Instruksi permanen untuk coding agent — Kos Management System

Spesifikasi utama proyek ada di `KONTEKS_LENGKAP_CODEX_KOS_MANAGEMENT.md`. **Wajib baca seluruh dokumen sebelum implementasi.** Bila instruksi singkat pengguna tidak secara eksplisit mengubah kebijakan bisnis, keputusan dari spesifikasi tetap berlaku.

- Prioritas integritas uang/akuntansi, privasi tenant, alur booking/kontrak yang benar, lalu UI. Semua antarmuka Bahasa Indonesia.
- Dua role login: owner dan tenant. Public applicant bukan role internal ketiga.
- Jangan ubah/hapus kode yang sudah berjalan secara tidak perlu; audit repositori dulu, hindari migrasi destruktif dan paparan rahasia.
- Setiap fitur wajib lengkap secara end-to-end: model/data, validation, authorization, business service, route/UI, success/error/empty states, test.
- Pembayaran transfer harus diverifikasi owner; tidak mengaku bank auto-terintegrasi.
- PLN token dan PDAM individual dibayar penghuni langsung; bukan invoice/pendapatan/beban kos. Utilitas bersama dibayar pemilik.
- Booking disetujui pemilik, booking money pemakaian dipilih per booking; refund dan timeout booking masih keputusan manual.
- Bulanan/tahunan; annual preset installments; due fleksibel; 29/30/31 fallback akhir bulan dan kembali anchor asal; no deposit, no automatic late fees.
- Expense wajib foto bukti. Kas kecil dipegang owner; multi-bank; opening balances tidak dianggap pendapatan. Ledger double-entry dan posting idempotent; close book bulanan oleh owner.
- Owner-only financial reports; tenant dapat lihat data dirinya saja; file privat dengan authorization.
- UI mobile-first, teks cukup besar, responsif HP kecil/layar lipat, bahasa Indonesia, tidak ada tombol dummy.
- Jalankan test dan lint; laporkan yang dikerjakan/yang belum secara akurat. Tidak mengklaim complete jika belum dites. Dokumentasikan setup, queue, email, schedule, backup.

Jika kode repo lama memakai stack lain, beradaptasi dengan stack itu selama bisa memenuhi spesifikasi; jangan rewrite total hanya karena dokumen menyebut preferensi Laravel.
