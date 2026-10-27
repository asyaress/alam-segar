# KONTEKS INDUK CODEX — SISTEM MANAJEMEN KOS TERINTEGRASI

**Nama kerja:** Kos Management System / KosKu (nama branding dapat diubah)
**Versi spesifikasi bisnis:** 1.0 | **Bahasa antarmuka:** Bahasa Indonesia | **Mata uang:** IDR | **Zona waktu bisnis:** Asia/Makassar
**Status:** Dokumen konteks implementasi. Seluruh ketentuan di bawah berasal dari empat putaran wawancara pemilik. Jika ada hal belum ditentukan, ikuti kebijakan sementara yang ditulis secara eksplisit; jangan membuat kebijakan baru diam-diam.

---

## 0. PERINTAH UTAMA UNTUK CODEX

Bangun aplikasi web **end-to-end yang benar-benar berfungsi** untuk pemilik kos dan penghuni, bukan dashboard dekoratif. Semua menu harus terhubung dengan data, workflow, otorisasi, validasi, pencatatan transaksi, dan pengujian. Kerjakan tugas secara bertahap di repositori sampai fitur sesuai acceptance criteria berjalan; setelah setiap tahap lanjutkan ke tahap berikutnya tanpa menunggu instruksi baru, sepanjang alat dan konteks memungkinkan. Jangan menyatakan selesai jika masih berupa mockup, tombol semu, angka hardcode, atau stub. Bila keterbatasan sesi membuat pekerjaan tidak dapat diselesaikan, tinggalkan catatan progres konkret, langkah melanjutkan, dan kegagalan tes; **jangan klaim telah menjalankan hal yang belum dilakukan**.

**Sebelum mengubah kode:** inspeksi struktur repositori, stack, migrasi, fitur aktif, route, git status, konfigurasi, tes, dan README. **Jika sudah ada aplikasi, pertahankan fitur yang sudah bekerja; lakukan perubahan sekecil mungkin yang perlu**. Jangan menimpa file pengguna, mereset database produksi, menghapus data, menyimpan rahasia di source, atau menjalankan migrasi destruktif tanpa izin eksplisit. Jika repositori masih kosong, bangun dari awal dengan arsitektur yang dijelaskan pada §17.

**Prioritas hasil:** integritas data dan uang > kontrol akses > alur operasional lengkap > UX responsif > kosmetik. Semua label, alert, validasi, email, kuitansi, laporan, menu, dan teks UI berbahasa Indonesia. Nama kelas/kode boleh bahasa Inggris sesuai konvensi.

## 1. VISI PRODUK DAN AKTOR

Aplikasi untuk mengelola satu bangunan kos saat ini, tetapi siap untuk **multi-bangunan/multi-properti** di masa depan. Menghubungkan pendataan bangunan, kamar, aset, kontrak, booking, tagihan, transfer bank, kas kecil, akuntansi, operasional, keluhan, notifikasi, laporan, serta portal penghuni tanpa input ganda.

**Hanya dua role login internal:**
1. **Pemilik (owner):** dapat mengelola properti, kamar, harga, akun, booking, kontrak, inventaris, kas/bank, akuntansi, pengeluaran, pembukuan, laporan, notifikasi, keluhan, dan pengaturan. Tidak ada role admin staf/manajer/teknisi sebagai akun login.
2. **Penghuni (tenant):** dapat melihat dan mengelola hanya data miliknya sendiri: profil, kamar aktif, kontrak & riwayat, tagihan, upload bukti transfer, riwayat verifikasi dan kuitansi, pengumuman, inventaris kamar yang relevan, serta keluhan.

**Calon penghuni:** pengunjung publik yang dapat melihat kamar yang ditawarkan, registrasi atau mengajukan booking. Ini status/alur pendaftaran, bukan role internal ketiga. Akun pendaftar yang belum disetujui **tidak** dapat membuka portal penghuni aktif. Pemilik dapat membuat akun penghuni atau menyetujui registrasi mandiri.

**Hak akses:** semua query, download, API, perubahan status, dan notifikasi tenant wajib dibatasi berdasarkan tenant_id dan kontrak terkait; property_id/room_id dari permintaan pengguna tidak boleh dipercaya begitu saja. Pemilik berwenang atas properti yang dimilikinya; sistem multi-properti harus siap multi-owner secara aman jika nanti diperlukan, tanpa membangun akses silang tidak sengaja.

## 2. KEPUTUSAN BISNIS YANG MENGIKAT (HASIL WAWANCARA TAHAP 1–4)

- **Properti:** sekarang satu bangunan; arsitektur dan UI mendukung banyak bangunan, blok, lantai dan kamar. Harga kamar sekarang sama, namun dapat ditetapkan per tipe/kamar dan berubah ke depan.
- **Hunian:** maksimum **satu penghuni per kamar** dalam kontrak aktif. Satu akun dapat punya riwayat banyak kontrak dari waktu ke waktu, tetapi tidak boleh menempati dua kamar aktif secara bersamaan kecuali pemilik secara eksplisit mengubah kebijakan di masa depan (tidak perlu fitur ini sekarang).
- **Sewa:** bulanan atau tahunan; jatuh tempo **fleksibel per kontrak**, tidak selalu tanggal masuk. Pembayaran tahunan bisa sekaligus atau menggunakan **skema cicilan preset**. Izin pembayaran sebagian terhadap suatu tagihan ditentukan **per kontrak**. Keterlambatan hanya pengingat, **tanpa denda otomatis**.
- **Aritmetika tanggal:** jika tanggal acuan jatuh pada 29/30/31 dan bulan tujuan lebih pendek, gunakan **tanggal terakhir bulan tersebut**. Tetap gunakan tanggal acuan asli pada bulan berikutnya, jangan menggeser anchor permanen. Periode tanggal harus konsisten, teruji untuk tahun kabisat dan perubahan tahun.
- **Deposit/jaminan:** **tidak digunakan**. Jangan memaksakan modul deposit sebagai bagian alur utama.
- **Booking:** calon penghuni **mengajukan**, pemilik **menyetujui/menolak**; booking memakai **uang booking**. Pemilik dapat memilih **per booking** apakah dana akan mengurangi sewa pertama atau menjadi biaya booking tambahan. **Aturan pembatalan/refund booking belum ditentukan:** ditangani keputusan manual pemilik yang terdokumentasi. **Batas waktu bayar booking belum ditentukan:** jangan membuat pembatalan otomatis; pemilik dapat melepas atau membatalkan manual. Hindari bentrokan tanggal booking/kontrak dalam transaksi atomik.
- **Bayar sewa:** **transfer bank**, dapat ke **beberapa rekening milik usaha/pemilik** yang terdaftar. Penghuni mengunggah bukti; **pemilik memverifikasi manual setelah mengecek mutasi**. Bukti upload saja tidak menandakan lunas. Jangan mengklaim adanya verifikasi otomatis bank.
- **Perpanjangan kontrak:** kontrak lama diperpanjang **dengan riwayat perubahan** (versi/amandemen), bukan menghapus nilai sebelumnya. Pemilik menentukan harga baru tiap kamar/kontrak.
- **Pindah kamar:** **akhiri kontrak lama dan buat kontrak baru**, kedua kontrak terhubung ke penghuni yang sama, semua histori tetap ada. Jangan hanya mengganti `room_id` pada kontrak lama.
- **Keluar sebelum waktunya:** keputusan potongan/pengembalian **kasus per kasus**, dengan alasan, nominal, bukti, dan konsekuensi akuntansi.
- **Inventaris:** aset dan kondisi **per kamar**; checklist dan foto kondisi pada **check-in dan check-out**. Kerusakan oleh penghuni **ditanggung penghuni**, nominal berdasarkan **biaya aktual atau kesepakatan**.
- **Listrik kamar:** setiap kamar menggunakan **token PLN sendiri** dan penghuni membeli langsung. Kos **tidak** menagih atau mencatat pembelian token sebagai pemasukan/beban usaha. Boleh simpan nomor meter/ID pelanggan untuk informasi privat kamar.
- **Air kamar:** masing-masing punya **nomor pelanggan/meter PDAM dan tagihan sendiri**; pembacaan meter dan penagihan dilakukan **langsung oleh PDAM**, pembayaran langsung oleh penghuni ke PDAM. **Jangan** buat tagihan air kamar, pencatatan angka meter manual, atau perhitungan tarif per m³ dalam invoice kos. Cukup simpan identitas PDAM dan petunjuk pembayaran bila diperlukan.
- **Utilitas bersama:** **listrik dan air fasilitas umum punya meter/rekening terpisah**, ditanggung **pemilik** dan masuk biaya operasional properti melalui pembayaran yang benar-benar terjadi.
- **Kas kecil:** dipegang dan dicatat oleh **pemilik**. Jumlah dana kas dapat ditentukan lewat aplikasi. Mendukung saldo awal, pemasukan/pengeluaran, transfer dari/ke bank, pengisian ulang, audit saldo.
- **Biaya operasional:** listrik bersama, PDAM bersama, WiFi, kebersihan, pemeliharaan, gaji jika ada, perlengkapan, dll. **Setiap transaksi pengeluaran yang dicatat wajib memiliki foto/scan bukti**. Aturan persetujuan staf tidak diperlukan karena hanya ada pemilik sebagai pengelola.
- **Catatan awal:** mulai operasional aplikasi **tanpa mengimpor histori transaksi lama**; tetap input **saldo aktual awal kas dan setiap rekening** pada tanggal go-live. Jika ada kontrak, piutang, kewajiban, atau booking aktif saat awal implementasi, sediakan pencatatan saldo awal terpisah secara opsional; jangan pernah menganggap saldo awal sebagai pendapatan.
- **Notifikasi:** di dalam **website dan email**. WhatsApp, push native, atau SMS bukan scope awal.
- **Tutup buku:** pemilik melakukan **penutupan bulanan secara manual** setelah rekonsiliasi dan pemeriksaan. Posting ke periode tertutup dilarang; gunakan mekanisme koreksi yang terlacak di periode terbuka.
- **Keluhan:** penghuni mengajukan keluhan/perbaikan dari website, dapat unggah foto, pemilik memberi respons/status dan biaya jika timbul.
- **Laporan:** dashboard okupansi dan bisnis, piutang/tunggakan, pendapatan, beban, rekening/kas, laba rugi, neraca, arus kas, buku besar/jurnal, inventaris, pemeliharaan, dan laporan per properti serta per periode.

## 3. PRINSIP PRODUCT DESIGN

- **Tidak membuat transaksi keuangan ganda.** Pembayaran yang diverifikasi menggerakkan subledger penyewaan sekaligus jurnal/akun kas-bank; bukan dibuat ulang oleh pemilik.
- **Pisahkan status kerja dari status uang:** formulir bukti transfer `pending` tidak berarti uang masuk, reservasi `approved` tidak sama dengan booking `paid`, invoice `sent` tidak sama dengan `paid`.
- **Tidak membebankan utilitas pribadi ke bisnis.** Nomor PDAM/PLN hanya metadata kamar, bukan piutang kepada tenant.
- **Histori immutable:** simpan perubahan kontrak, tarif, status, jurnal posted, upload bukti dan penghapusan logis yang aman.
- **Satu sumber kebenaran:** saldo dihitung dari jurnal/transaksi posted, okupansi dari kontrak/check-in dan reservasi, piutang dari tagihan/pembayaran valid.
- **Tidak boleh delete transaksi akuntansi yang sudah posted;** pembatalan lewat reversal/adjustment dengan alasan. Jangan mengubah nominal posted diam-diam.
- **Untuk fitur yang belum pasti** (refund booking, booking timeout, pengembalian kontrak dini), keputusan manual pemilik per kasus dengan field alasan dan audit. Jangan membuat aturan otomatis yang belum diminta.

## 4. MODUL PEMILIK: PROPERTI DAN KAMAR

**Data properti:** nama kos, kode, alamat, foto cover/galeri, deskripsi, peraturan, kontak, jumlah lantai/blok, fasilitas bersama, informasi utilitas bersama, status aktif/nonaktif, identitas rekening default, metadata.

**Data kamar:** properti, blok/lantai, nomor unik per properti, tipe, luas opsional, harga dasar per bulan/tahun, foto, fasilitas, kapasitas default satu, ID token PLN, nomor pelanggan PDAM, status kondisi (layak, renovasi, nonaktif), histori tarif, catatan, aset terkait. Harga kesepakatan kontrak disalin sebagai snapshot saat kontrak dibuat; perubahan tarif katalog tidak mengubah kontrak berjalan secara diam-diam.

**Status kamar secara bisnis:** tersedia; diajukan; ditahan/menunggu booking; ter-booking; dihuni; checkout-pending; perawatan; nonaktif. **Status keterisian/booking harus diturunkan dari kontrak, reservasi, jadwal dan kondisi fisik**; jangan hanya mengandalkan kolom status yang bisa bertentangan. Validasi interval waktu untuk booking masa depan (ketersediaan kamar per tanggal).

**Aksi pemilik:** tambah/edit kamar, unggah foto, nonaktifkan kamar jika kosong, melihat histori penghuni dan pembayaran (hanya pemilik), melihat aset serta inspeksi, blokir penyewaan untuk perbaikan, ekspor daftar kamar.

**Halaman publik:** landing page kos, katalog kamar yang dipublikasikan, foto/fasilitas/harga mulai dari, indikator ketersediaan berdasar tanggal, detail kamar, formulir pengajuan booking dan pendaftaran. Jangan paparkan identitas tenant, rincian bank internal, nomor pelanggan utilitas, atau data pemilik yang sensitif.

## 5. MODUL PEMILIK: REGISTRASI, PENGHUNI, DAN KONTRAK

**Registrasi:** akun penghuni dapat dibuat pemilik atau calon penghuni mendaftar sendiri. Verifikasi email, status pending/approved/rejected, aktivasi setelah disetujui. Cegah duplikasi akun dan eskalasi role. Identitas minimum: nama, kontak, email, kontak darurat (opsional), dokumen identitas jika benar-benar diperlukan dan privat; data sensitif diberi akses terbatas dan retensi yang wajar.

**Kontrak:** nomor, tenant, kamar, properti, jenis bulanan/tahunan, tanggal mulai/akhir, anchor tanggal, due date spesifik/skema, harga disepakati, skema cicilan preset bila tahunan, partial payment allowed, ketentuan khusus, status draft/active/expiring/ended/terminated, dokumen perjanjian, histori versi/perpanjangan, pemilik pembuat, tautan booking awal.

**Pengecekan:** satu kamar tidak boleh punya dua kontrak yang overlap dan aktif, satu tenant tidak boleh dua kamar aktif, kontrak hanya dibuat pada kamar layak/tersedia, booking yang digunakan tidak bisa ditukar ke kontrak lain tanpa reversal.

**Perpanjangan:** buat catatan perubahan/amendment yang menyimpan snapshot sebelum dan sesudah (periode, harga, metode bayar, tanggal efektif, alasan), tagihan periode baru otomatis direncanakan. Hindari double invoice pada tanggal batas.

**Check-in:** verifikasi kontrak dan pembayaran yang disyaratkan, checklist fasilitas + foto kondisi, tanggal actual check-in, serah terima, aktifkan akses portal tenant dan status keterisian.

**Check-out:** catat tanggal keluar, checklist/foto akhir, pemeriksaan inventaris, kerusakan serta tagihan ganti rugi, pembayaran/refund sesuai keputusan manual, penyelesaian kontrak; kamar **tidak dianggap langsung siap dijual** sebelum inspeksi/status kondisi dituntaskan. Pindah kamar: selesai kontrak lama lalu kontrak baru.

## 6. MODUL BOOKING DAN UANG BOOKING

**Workflow:** `submitted` -> `under_review` -> `approved_waiting_payment` -> `proof_submitted` -> `confirmed` -> `converted_to_contract`; cabang `rejected`, `cancelled`, `released`, dan `refund_pending/refunded` sesuai realita. Pemilik menilai permohonan, tidak otomatis menyetujui. Pada `approved_waiting_payment` kamar dapat ditahan sementara; hold dilepaskan **manual** karena batas waktu otomatis belum ditetapkan. Konfirmasi pembayaran hanya setelah owner memverifikasi mutasi.

**Data:** pemohon, kamar, tanggal check-in rencana dan periode rencana, nominal booking, rekening tujuan, due-by nullable, status, waktu submit/approve, alasan tindakan, file bukti, payment verification, pilihan penggunaan booking (`credit_first_rent` atau `additional_booking_fee`), nilai diakui/dikompensasikan, histori dan catatan refund manual.

**Aturan:** pengajuan baru tidak otomatis mencadangkan kamar secara permanen; ketika owner menyetujui dan membuat hold, pemeriksaan bentrok tanggal harus transaksional. Refund booking ditentukan manual owner; sistem harus menghindari double refund dan saldo liability negatif. Dana booking yang digunakan sebagai pengurang sewa harus ditransfer secara internal dari kewajiban booking ke tagihan sewa, **bukan dihitung sebagai penerimaan bank kedua kali**. Booking fee tambahan baru diakui sesuai kebijakan/ketika hak atas dana telah diperoleh, bukan saat bukti transfer belum diverifikasi.

## 7. MODUL PENAGIHAN, CICILAN DAN PEMBAYARAN

**Tagihan/invoice:** nomor unik, property/room/tenant/contract, jenis (sewa, booking, ganti rugi, penyesuaian), periode layanan, tanggal terbit, tanggal jatuh tempo, komponen, total, diskon/penyesuaian yang diizinkan, status draft/issued/partially_paid/paid/overdue/void/refunded dan outstanding hasil kalkulasi. Tanpa baris air PDAM kamar atau token PLN kamar.

**Sewa bulanan:** jadwalkan tagihan berdasarkan kontrak, tanggal jatuh tempo fleksibel sesuai perjanjian; untuk anchor tanggal 31: Februari jatuh pada tanggal terakhir Februari, Maret kembali 31. Periode pembayaran dihitung konsisten dan tidak tumpang tindih.

**Sewa tahunan:** harga disimpan untuk periode kontrak. Bisa invoice sekaligus atau **preset installment scheme** (misal satu kali, dua kali, triwulanan/4 kali, enam kali, bulanan/12 kali; preset dapat dikelola pemilik). Total nilai angsuran tepat sama dengan total kontrak setelah pembulatan; sisa pembulatan masuk angsuran terakhir. Jadwal jatuh tempo bisa disesuaikan per kontrak sebelum diterbitkan.

**Partial payment:** izin adalah atribut kontrak. Jika tidak diizinkan, pembayaran terverifikasi kurang dari sisa tagihan harus ditolak atau diselesaikan melalui kebijakan pemilik yang tercatat; jika diizinkan, status `partially_paid` dan sisa tagihan tetap aktif.

**Alur transfer:** tenant pilih tagihan -> pilih rekening tujuan yang diizinkan -> upload bukti transfer + nominal/tanggal/reference -> status `pending_verification` -> owner cocokkan ke rekening -> `verified` atau `rejected` dengan alasan -> hanya saat verified jurnal dan alokasi pembayaran dibuat. Tampilkan resi/kuitansi sesudah verified. Validasi file, nominal dan duplikasi bukti/referensi; dukung transfer yang membayar lebih dari satu tagihan hanya jika dialokasikan jelas, tanpa double book. Pembayaran di muka dan kelebihan bayar harus jadi kredit/liabilitas untuk dipakai nanti, tidak otomatis laba.

**Keterlambatan:** tandai invoice overdue berdasar tanggal lokal dan saldo outstanding. Tidak ada denda. Pengingat website/email bisa misalnya sebelum jatuh tempo, saat jatuh tempo, dan setelah jatuh tempo; jadwal bisa dipilih pemilik, anti-spam/idempotent.

**Refund sewa:** pemilik membuat keputusan manual yang disertai nominal, alasan, sumber rekening, kaitan pembayaran/invoice, dan bukti. Status refund dan reversal jurnal konsisten; tidak boleh melebihi saldo yang berhak dikembalikan.

## 8. MODUL INVENTARIS, INSPEKSI DAN KERUSAKAN

**Master aset:** kode, nama barang, kategori, serial opsional, tanggal/perolehan/harga, lokasi properti/kamar/area umum, kondisi, foto, dokumen garansi, catatan perawatan, status pindah/rusak/hilang/dijual, histori mutasi. Aset tetap vs barang habis pakai dipisah menurut kebijakan akuntansi.

**Check-in / check-out:** checklist item per kamar (ada/tidak, kondisi, jumlah), foto bertanggal, catatan, petugas pencatat (owner), tanda tangan/persetujuan digital sederhana bila memungkinkan. Simpan snapshot supaya perubahan master aset di masa depan tidak mengubah hasil inspeksi lama.

**Kerusakan:** bila tanggung jawab tenant, owner memilih biaya aktual dari invoice perbaikan atau nominal hasil kesepakatan; buat tagihan ganti rugi terpisah, bisa dilihat tenant. Biaya vendor/perbaikan tetap pengeluaran usaha tersendiri, bukan otomatis di-net-off. Dilarang membuat piutang tanpa keputusan pemilik.

## 9. MODUL KELUHAN DAN PEMELIHARAAN

Tenant mengajukan tiket dari kamar/kontraknya, kategori (listrik, air, AC, furnitur, keamanan, kebersihan, lain), deskripsi, foto, tingkat urgensi. Owner menerima, memberi respons, mengubah status `open/in_progress/waiting_for_tenant/resolved/closed`, mencatat tindakan, pengeluaran perbaikan (dengan bukti), aset terdampak, dan penutupan. Tenant melihat tiketnya saja. Setiap status penting memicu notifikasi website/email. Tidak perlu akun petugas/teknisi; nama vendor atau teknisi boleh dicatat sebagai teks atau master vendor (tanpa login).

## 10. MODUL KEUANGAN: KAS, BANK, BIAYA, TRANSFER

**Akun keuangan:** beberapa rekening bank (bank, pemilik rekening, nomor rekening yang ditampilkan ke tenant hanya jika diizinkan untuk pembayaran, saldo pembukuan, status), satu atau lebih kas kecil/tunai dengan penanggung jawab pemilik, properti terkait bila perlu. Saldo awal aktual per rekening/kas pada tanggal go-live dengan jurnal pembuka; bukan pendapatan. Jangan menyimpan kredensial internet banking.

**Penerimaan:** booking terverifikasi, sewa terverifikasi, ganti rugi terverifikasi, pemasukan lain bila diinput owner. Semua dengan tanggal, rekening, referensi, dokumen, properti, penghubung entitas bisnis, jurnal.

**Pengeluaran:** kategori COA (listrik operasional, PDAM operasional, internet, kebersihan, perawatan, gaji, perlengkapan, pembelian aset, lain), vendor/payee, properti, tanggal, rekening/kas, nominal, keterangan, **foto/scan bukti wajib**. Pembayaran invoice vendor dapat dicatat bila diperlukan; status draft belum memengaruhi kas sampai diposting.

**Transfer internal:** antarbank, bank-ke-kas, kas-ke-bank; debit-kredit antar akun keuangan dalam satu transaksi. Tidak boleh dihitung sebagai pendapatan atau beban. Hindari overdraft saldo kas bila saldo tidak mencukupi (atau batasi berdasarkan setting eksplisit dengan audit).

**Rekonsiliasi:** owner cocokkan transaksi dengan rekening/uang fisik, tandai `reconciled`, telusuri transaksi belum cocok, koreksi melalui reversal. Tidak ada klaim sinkronisasi otomatis bank.

**Kas kecil:** isi saldo, bayar biaya, tampilkan laporan per periode, lampiran nota, saldo akhir, histori; owner memegang kas secara fisik.

## 11. MODUL AKUNTANSI: DOUBLE ENTRY DAN TUTUP BUKU

Implementasi **ledger debit-kredit seimbang**, mencakup:
- Chart of Accounts (COA) berkode, kategori aset, kewajiban, ekuitas, pendapatan, beban. Contoh: Kas, Bank per rekening, Piutang Sewa, Piutang Ganti Rugi, Pembayaran Sewa di Muka/Pendapatan Diterima di Muka, Kewajiban Uang Booking, Utang Refund, Perlengkapan/Aset Tetap, Akumulasi Penyusutan jika digunakan, Modal/Saldo Awal, Prive, Pendapatan Sewa, Pendapatan Booking yang Sah, Pemulihan Kerusakan, Beban Listrik Bersama, Beban Air Bersama, Beban Internet, Kebersihan, Perawatan, dan lain-lain.
- Journal entries header + lines minimal dua, seimbang tepat sampai satuan rupiah; posted entries immutable, transaksi bisnis memicu satu posting idempotent dengan `source_type/source_id/event_type` unik.
- Subledger piutang, booking liability, kontrak, bank/kas disinkronkan dengan general ledger. Tidak boleh membuat posting manual duplikat dari transaksi otomatis.
- **Prinsip akrual:** uang diterima berbeda dengan pendapatan yang dihasilkan. Sewa dibayar di muka: Dr Bank / Cr Piutang atau Kewajiban Sewa Di Muka sesuai invoice/posting, lalu akui pendapatan sesuai periode layanan. Jika jasa sewa sudah berjalan tetapi invoice belum diterbitkan, rancangan harus mencakup pendapatan masih harus ditagih/penyesuaian yang tepat agar laporan per bulan akurat. Hindari pengakuan ganda saat invoice diterbitkan dan pembayaran masuk. Setiap skenario harus diuji dengan ledger.
- **Uang booking:** Dr Bank / Cr Kewajiban Booking saat pembayaran verified. Saat dialokasikan ke sewa: Dr Kewajiban Booking / Cr Piutang Sewa atau akun yang sesuai dengan siklus invoice. Saat menjadi biaya booking yang sah: Dr Kewajiban Booking / Cr Pendapatan Booking pada saat haknya diperoleh. Refund: Dr Kewajiban Booking/Utang Refund / Cr Bank.
- **Biaya operasional:** Dr Beban bersangkutan / Cr Kas/Bank saat diakui/dibayar (atau melalui utang vendor bila jatuh tempo berbeda).
- **Transfer internal:** Dr Kas / Cr Bank, bukan beban usaha.
- **Ganti rugi:** pengeluaran perbaikan dan tagihan pemulihan kepada tenant dicatat terpisah, klasifikasi pendapatan penggantian mengikuti kebijakan yang konsisten.
- **Pengambilan pemilik:** Dr Prive / Cr Kas/Bank; jangan masuk beban kos.
- **Pembelian aset:** Dr Aset Tetap / Cr Bank/Kas; biaya penyusutan opsional bila diaktifkan dan kebijakannya ditetapkan; jangan langsung memasukkan seluruh nilai aset sebagai beban jika dikapitalisasi.
- **Saldo awal:** jurnal yang seimbang antara aset/kewajiban/ekuitas awal, termasuk saldo kas & bank; jika tanpa histori, mulai dari tanggal go-live dengan pernyataan keterbatasan komparatif.
- **Tutup buku bulanan:** owner memeriksa rekening, kas, invoice, pembayaran pending, jurnal tak seimbang, revenue recognition, saldo booking, kemudian manual `close_period`. Blok perubahan/posting backdated ke periode closed; koreksi via reversal/new adjustment di periode terbuka dengan tautan transaksi lama. Periode boleh dibuka ulang hanya bila disediakan prosedur khusus dengan audit kuat, secara default tidak.
- Laporan: trial balance, general ledger, jurnal, piutang umur, neraca, laba rugi, arus kas (langsung dari klasifikasi kas), pendapatan sewa, laporan per properti, kas/bank, serta ekspor CSV/XLSX/PDF jika dukungan tersedia. **Jangan menyajikan angka kas sebagai laba.**

## 12. PORTAL PENGHUNI (MOBILE FIRST)

**Halaman tenant:**
1. Dashboard: nama kos, nomor kamar, status kontrak, nominal & jatuh tempo terdekat, pemberitahuan terbaru.
2. Kamar Saya: foto dan fasilitas kamar yang bersangkutan, informasi identitas meter/nomor pelanggan PLN/PDAM milik kamarnya jika diperbolehkan, aturan kos, daftar inventaris yang diserahkan.
3. Kontrak Saya: periode, harga, skema cicilan, histori perpanjangan, dokumen kontrak, unduhan.
4. Tagihan Saya: daftar semua invoice miliknya, status dan rincian cicilan/sisa, rekening tujuan transfer.
5. Konfirmasi Pembayaran: unggah bukti, nominal, tanggal, referensi; status menunggu/dikonfirmasi/ditolak dan alasan. Riwayat dan kuitansi verified.
6. Keluhan Saya: buat tiket, foto, timeline tanggapan, status penyelesaian, tagihan ganti rugi bila valid.
7. Pengumuman & Notifikasi: dibaca/belum dibaca, email terkait.
8. Profil & keamanan: ubah data yang diizinkan, ubah password, verifikasi email.

**Tenant tidak boleh:** melihat tagihan, ID pribadi, foto bukti, kontrak, atau laporan penghuni lain; mem-posting jurnal; mengubah status pembayaran; menyetujui booking sendiri; mengganti kamar sendiri tanpa kontrak baru.

## 13. HALAMAN PEMILIK DAN NAVIGASI

**Navigasi owner yang berfungsi dan punya empty states:**
- Dashboard
- Properti: daftar/tambah/edit/detail, fasilitas umum, laporan lokasi
- Kamar: grid okupansi, tipe/harga, detail, galeri, pelanggan PLN/PDAM, kondisi
- Booking: pengajuan, review, pembayaran, refund manual, konversi kontrak
- Penghuni: akun, registrasi pending, identitas, riwayat, kontak
- Kontrak: buat/aktif/perpanjang/amandemen/akhiri/check-in/check-out
- Tagihan: daftar, jadwal, cicilan, pembayaran, verifikasi, kuitansi, refund
- Kas & Bank: saldo awal, rekening, kas kecil, mutasi, transfer internal, rekonsiliasi
- Pengeluaran: kategori, entri, foto bukti, pembelian aset/biaya, laporan
- Akuntansi: COA, jurnal, buku besar, piutang, neraca, laba rugi, arus kas, tutup buku
- Inventaris: master aset, penempatan, checklist inspeksi, foto, ganti rugi
- Pemeliharaan & Keluhan: tiket, status, histori, biaya, vendor
- Pengumuman & Notifikasi: buat pengumuman per properti, pesan terkirim, template email
- Laporan: keuangan/okupansi/umur piutang/performance per properti
- Pengaturan: profil usaha, tanggal go-live, master kategori, skema cicilan, akun penerimaan, template kontrak, jadwal reminder, branding, backup/pengelolaan data

**Dashboard owner** mengutamakan kartu: total kamar, terisi/kosong/maintenance/booking, tingkat hunian, sewa jatuh tempo, piutang overdue, booking menunggu persetujuan, transfer pending, saldo bank per akun, kas kecil, pendapatan terakui periode, beban dan hasil operasi, tiket belum selesai, kontrak akan berakhir; grafik tren dengan data asli dan filter periode/properti.

## 14. NOTIFIKASI WEB DAN EMAIL

Notifikasi dipicu oleh: registrasi disetujui/ditolak, booking diajukan/disetujui/ditolak, bukti booking terverifikasi/ditolak, booking dibatalkan/refund diproses, kontrak aktif/diperpanjang/berakhir, invoice diterbitkan, pengingat menjelang/saat/setelah jatuh tempo, bukti pembayaran masuk, pembayaran diverifikasi/ditolak, tiket baru/status berubah, pengumuman kos.

Owner dapat mengubah jadwal reminder; default rasional seperti H-7/H-3/H/H+3 boleh ditawarkan sebagai setting **bukan kebijakan denda**. Scheduled job harus idempotent, tidak men-spam, dan menghormati email verification/unsubscribe untuk komunikasi non-transaksional. Catat `notification_deliveries` dan kegagalan email; retry aman. Pesan email berbahasa Indonesia dan tidak membocorkan data privat. Tidak membangun WhatsApp/SMS.

## 15. DESAIN TAMPILAN DAN PENGALAMAN PENGGUNA

- Desain modern, rapi, elegan, simpel, mudah digunakan orang awam, konsisten di owner dan tenant. Warna netral dengan aksen satu warna, tipografi nyaman.
- Bahasa **Indonesia 100% di UI**; gunakan istilah 'Kamar Saya', 'Tagihan', 'Konfirmasi Transfer', 'Dana Kas', 'Laporan Keuangan', 'Tutup Buku', bukan labels Inggris acak.
- **Mobile first**, responsif dari lebar 320px, 360px, 375px, 390px, tablet, desktop, layar lipat dan split-screen. Tidak ada horizontal overflow; tabel di HP menjadi kartu/scroll lokal yang terbaca; sidebar menjadi drawer; tombol minimal nyaman untuk sentuh; font teks utama tidak terlalu kecil (sekitar ≥16px pada layar normal), hierarki informasi jelas.
- Semua fungsi wajib punya loading, sukses/error, empty state, konfirmasi untuk tindakan irreversibel, validasi input yang menjelaskan kesalahan, upload preview bukti, pagination/search/filter/sort.
- Aksesibilitas: navigasi keyboard, label formulir, fokus terlihat, kontras baik, semantic HTML, alt gambar, responsif saat zoom 200%.
- Dashboard menampilkan data riil, bukan grafik dummy setelah produksi. Seed demo jelas terpisah dari data asli.

## 16. KEAMANAN, INTEGRITAS, PERFORMA

- Autentikasi kuat; verifikasi email, reset password, rate limit login & booking, proteksi CSRF/XSS/SQL injection, validasi server-side untuk setiap action, policy per entitas.
- Owner/Tenant tidak boleh dapat akses lintas akun melalui IDOR, URL file, API, laporan, export, atau attachment. Foto identitas, bukti transfer, kontrak, bukti pengeluaran disimpan **private** dan diakses melalui authorization/signed download, bukan public URL.
- File upload cek MIME sesungguhnya, ukuran, ekstensi, simpan nama acak, batasi akses, hindari executable uploads. Jangan menampilkan nomor rekening atau data penghuni lebih dari diperlukan.
- Ledger **database transaction + lock + unique idempotency key** agar double click, retry, cron, verifikasi ganda, atau race condition booking tidak membuat double posting/double booking/double refund.
- Password hash aman; tidak hardcode kredensial/API key, `.env` rahasia tidak commit; pemisahan config dev/staging/production.
- Audit trail owner actions penting: create/edit/delete operasional, perubahan harga, persetujuan booking, approval registrasi, upload/verifikasi pembayaran, reversal, refund, saldo awal, perpanjangan kontrak, tutup buku, riwayat before/after, waktu, actor.
- Backup database dan file dengan prosedur restore terdokumentasi. Log tanpa membocorkan PII/nomor rekening penuh.
- Queue untuk email, penagihan terjadwal, reminder, laporan berat. Cron/scheduler serta worker benar-benar dikonfigurasi/didokumentasi. Cegah N+1, indeks foreign key + tanggal/status.

## 17. PREFERENSI TEKNIS (FLEKSIBEL KEPADA CODEX)

**Jika repositori sudah ada:** jangan ganti stack atau rewrite total. Gunakan komponen dan konvensi yang sudah ada, serta buat migrasi yang aman dan bertahap.

**Jika mulai dari nol:** prefer **Laravel versi stabil yang kompatibel lingkungan proyek + PHP + MySQL/MariaDB**, UI Blade/Livewire/Tailwind, panel owner Filament bila cocok dengan versinya; portal tenant dan katalog publik dioptimalkan mobile. Gunakan Laravel policy/gate, database migrations, seeders, Form Requests/validation, Eloquent scopes, service/action classes untuk proses finansial, events/notifications/queues/scheduler, dan automated tests. Jangan mengunci nomor versi tanpa cek runtime/kompatibilitas. Frontend framework kedua (Next.js/React) tidak wajib; hindari kompleksitas tanpa alasan. PWA opsional setelah web utama selesai.

**Domain modules/isolasi service** disarankan: Properties, Rooms, People, Bookings, Leases, Billing, Payments, Finance, Accounting, Assets, Maintenance, Notifications, Reports. Akuntansi sebaiknya di service transaksional, bukan side effect controller yang mudah double-post.

**Tanggal**: simpan timestamps secara konsisten (umumnya UTC) dan tampilkan zona Asia/Makassar; tanggal jatuh tempo adalah tanggal bisnis lokal (DATE) bukan konversi timestamp UTC yang bisa bergeser. **Uang**: gunakan integer rupiah / decimal fixed scale, tidak floating point JS/PHP untuk operasi uang; semua alokasi & pembulatan deterministik.

## 18. MODEL DATA KONSEPTUAL — ENTITAS DAN RELASI

Ini blueprint, bukan keharusan nama tabel literal; Codex sesuaikan jika repo sudah memiliki struktur.

**Pengguna dan properti:**
- `users` (id, role owner/tenant, verified_at, status, profile...)
- `properties` (id, owner_id, name, address, status, ...)
- `blocks`, `floors` atau data lantai langsung pada rooms sesuai skala
- `room_types`, `rooms` (property_id, number, type_id, capacity=1, base_rate, maintenance_flag, utility_customer_numbers private...)
- `room_photos`, `property_photos`, `room_amenities`, `pricing_histories`
- `tenant_profiles` (user_id, emergency_contact, ...)
- `registration_requests` (applicant, state, approval...)

**Booking dan kontrak:**
- `booking_requests` (property_id, room_id, applicant_id nullable, interval, amount, approval/status, policy decision, hold...)
- `booking_payment_submissions` / `payment_submissions` dengan referensi pembayaran yang diverifikasi
- `booking_liability_allocations`, `booking_refunds`
- `leases` (property_id, room_id, tenant_id, interval, monthly/annual, due_rule, rent_snapshot, partial_allowed, status...)
- `lease_amendments`, `lease_extensions` atau `lease_versions` immutable
- `installment_schemes`, `installment_scheme_items`, `lease_installments`
- `check_in_outs`, `handover_reports`

**Billing dan finance:**
- `invoices`, `invoice_lines`, `invoice_allocations`
- `payment_submissions` (file, payee bank, amount/date/reference, status, reviewer...) dan `payments` verified
- `payment_allocations`, `refunds`, `credit_balances` / liabilities
- `bank_accounts`, `cash_accounts`, `transfers`
- `expenses` (mandatory proof attachment, account, category, property...), `expense_attachments`, `vendors`, `opening_balances`

**Akuntansi:**
- `chart_of_accounts`, `journal_entries`, `journal_lines`, `accounting_periods`, `accounting_closures`, `posting_idempotency`, `reconciliations`, `reconciliation_items`
- revenue-recognition schedule/adjustments as needed

**Inventaris & operasional:**
- `assets`, `asset_movements`, `room_inventory_assignments`, `inspections`, `inspection_items`, `inspection_photos`, `damage_charges`
- `maintenance_tickets`, `maintenance_updates`, `maintenance_attachments`, `maintenance_expenses`
- `announcements`, `notifications`, `notification_deliveries`, `audit_logs`, `settings`

**Constraint:** foreign keys, composite unique room number per property, unique reference numbers, indexes date/status/owner scope, no overlapping active room allocation enforced at service transaction level with locks (SQL constraint if supported), unique ledger source/event, positive transaction amount where appropriate, debits=credits per posted journal, no negative outstanding allocations, enforce tenant ownership.

## 19. WORKFLOWS KRITIS DAN KASUS TEPI

**P0-A Booking:** public request -> owner review -> hold when approved -> transfer proof -> owner verifies -> confirmed -> linked lease. If owner rejects or cancels, release hold; refund/forfeit decision manual. Race condition two approvals same room/date must result in exactly one hold.

**P0-B Tenant lifecycle:** registration -> verified -> approved -> assigned via contract -> check-in inspection -> active -> extension/history -> checkout & settlement -> archived history; tenant access after checkout only historical self-data as configured.

**P0-C Monthly rent:** contract starts 31 Jan -> February due end-of-month -> March due 31 -> no duplicates/skips. Dates and invoicing anchored to original contract rule.

**P0-D Annual installment:** contract Rp12jt, 4 installments @Rp3jt on scheduled dates -> payments (including partial if enabled) update installment balances, but **pendapatan akrual sepanjang 12 bulan**. No duplicate recognition.

**P0-E Verification:** tenant uploads two proofs or clicks twice -> no ledger until owner verify -> verify twice produces **one** receipt and journal -> reject requires reason -> re-upload allowed.

**P0-F Booking to invoice:** Rp200rb verified booking credit towards Rp1,2jt rent -> tenant owes Rp1jt, bank booking cash exists once, booking liability cleared by allocation, bank not credited twice.

**P0-G Operational utilities:** owner pays shared PLN Rp300rb/PDAM Rp150rb -> expense posted from selected bank; tenant PLN token and tenant PDAM not in kos general ledger.

**P0-H Kas transfer:** bank->kas Rp1jt; expenses Rp150rb (proof compulsory) -> kas ending Rp850rb, business operating expense only Rp150rb; transfer has zero effect to profit.

**P0-I Room move:** end old lease A1, checkout inventory, create new lease B2; old invoice and room history remain; never edit prior lease room.

**P0-J Early exit:** decide prorated/manual refund; record owner justification and bank payment; occupancy and posted revenues corrected consistently per contract.

**P0-K Asset damage:** tenant breaks AC; expense vendor Rp450rb + tenant damage charge Rp450rb, track separately; booking/refund/other journal unaffected.

**P0-L Month close:** owner checks, closes September; backdate payment/expense to September forbidden, correction in October with reference to source; trial balance remains balanced.

**P0-M Multi-property:** second property isolated in filters/accounting dimension, owner sees both; a tenant in A cannot view tenant B; separate room no. '01' allowed per property.

**P0-N No deposit:** check-in should not block because deposit not received; no deposit invoice generated.

**P0-O No bank automation:** screenshot/transfer proof does not mark receipt until owner verified, and no claim of actual bank balance auto-sync; shown bank balance is **saldo pembukuan** unless reconciled.

## 20. ACCEPTANCE CRITERIA — YANG HARUS BENAR-BENAR DIBUKTIKAN

1. Pemilik bisa login; tenant hanya bisa login sebagai tenant setelah disetujui.
2. Pengunjung dapat melihat kamar yang dipublikasikan dan mengirim booking tanpa melihat informasi penghuni.
3. Dua booking disetujui untuk kamar dan tanggal bentrok tidak bisa sama-sama confirmed, termasuk saat request paralel.
4. Alur booking lengkap dengan bukti, review, status, credit ke sewa, fee tambahan, dan cancellation/refund manual berfungsi.
5. Dua properti dapat ditambahkan; laporan bisa difilter per properti.
6. Harga katalog yang berubah tidak diam-diam mengubah kontrak aktif.
7. Kontrak bulanan/tahunan dan skema angsuran preset menghasilkan jadwal/tagihan yang tepat.
8. Uji tanggal 31 Januari -> akhir Februari -> 31 Maret; uji 29 Februari tahun kabisat; tidak drift.
9. Tagihan boleh partial hanya bila kontrak mengizinkan; outstanding akurat dan no negative balance.
10. Penghuni unggah bukti; saldo bank dan invoice tidak berubah sebelum verifikasi; setelah verifikasi tepat satu posting.
11. Tenant hanya bisa membuka file/kamar/invoice/payment/kontrak/keluhan miliknya; akses ID orang lain ditolak.
12. PLN pribadi/PDAM pribadi tidak menghasilkan invoice atau pengeluaran kos.
13. Listrik/PDAM umum masuk akun pengeluaran terpisah dan properti yang benar.
14. Rekening bank jamak; transfer antarbank atau ke kas tidak mengubah laba rugi.
15. Saldo awal kas/bank membentuk jurnal seimbang dan bukan pendapatan.
16. Pengeluaran tanpa file bukti valid ditolak.
17. Pindah kamar menggunakan kontrak baru; kontrak dan histori kamar lama tetap dapat diakses yang berwenang.
18. Perpanjangan menyimpan versi harga, tanggal, dan syarat sebelumnya.
19. Checklist & foto check-in/out tersimpan dan dapat ditinjau oleh owner.
20. Kerusakan ditagih secara terpisah; pengeluaran servis dan pembayaran tenant memiliki jurnal masing-masing.
21. Pendapatan sewa dibedakan dari penerimaan kas; laporan akrual periode tahunan benar.
22. Setiap posting jurnal seimbang (debit = kredit), idempotent; reversal terhubung dengan sumbernya.
23. Close bulan mengunci transaksi backdate; koreksi tetap bisa dengan audit pada periode terbuka.
24. Semua reminder terkirim satu kali per event/penerima sesuai aturan, email dan notifikasi website; kegagalan dicatat.
25. Kuitansi hanya bisa diunduh untuk pembayaran verified dan tersedia buat tenant yang berhak.
26. Dashboard/laporan menggunakan data database, bukan fake numbers.
27. UI tetap layak pada layar mobile sempit, tablet, desktop, dan viewport foldable; tanpa tabel pecah.
28. Test otomatis berjalan di lingkungan yang tersedia, report failures transparan dan tersisa apa.
29. Proyek memiliki `.env.example`, README instalasi/run, migrasi/seed dev, akun demo aman (hanya local), queue/scheduler/email setup, prosedur backup.
30. Setelah deployment/staging, lint/smoke route, cek permissions dan tidak ada halaman error 500 pada alur utama.

## 21. URUTAN PENGEMBANGAN END-TO-END

**Fase 0 — Audit repositori & perencanaan.** Tentukan apa yang sudah ada, risiko regresi, dependencies, command test; buat daftar backlog checklist konkret. Jangan langsung mengganti stack.

**Fase 1 — Fondasi produk.** Auth dua role, email verification, registrasi/approval, property/rooms/type/photos/facilities, publik room listing, dashboard layout yang responsive, database constraints, policies.

**Fase 2 — Booking, kontrak, inspeksi.** Public application owner approval + manual hold, verified booking funds, tenant creation, lease, monthly/yearly pricing/installments, check-in/out, lease versioning, room availability.

**Fase 3 — Billing dan pembayaran.** Invoice generator, configurable due date, e-proof upload, owner manual verification, partial and credits, receipts, reminders. Cek idempotency dan auth.

**Fase 4 — Akuntansi dan kas/bank.** COA, source posting service, per-bank ledger, petty cash, opening balances, expenses mandatory proof, bank transfers, booking liabilities, accrual recognized income, refunds, profit/loss, balance sheet, cashflow, monthly close.

**Fase 5 — Inventaris, kerusakan, tiket.** Assets, photo checklists, maintenance ticket process, separate damages invoices and repair expenses.

**Fase 6 — Laporan, email, QA dan deployment.** Full dashboard, filters/export, idempotent scheduled reminders, accessibility/mobile fixes, regression tests, setup docs, backup, smoke tests.

**Pada tiap fase**: implement feature + migration + authorization + unit/feature tests + actual interactive UI + seed dev + update progress. Jangan hanya menyelesaikan CRUD tanpa workflow. Jalankan tes yang memungkinkan. Bila Codex perlu mengambil keputusan yang belum disepakati, pilih opsi konservatif dan terdokumentasi; jangan menciptakan kebijakan hukum/penalty/tarif baru.

## 22. DEFINITION OF DONE / HASIL AKHIR YANG WAJIB DILAPORKAN

Codex hanya boleh mengatakan fitur selesai jika: route dapat diakses sesuai izin, data tersimpan benar, validasi berjalan, transaksi finansial seimbang, user dapat menuntaskan workflow dari awal sampai akhir, dan tes terkait lulus. Untuk setiap fase jelaskan perubahan file, fitur aktif, perintah yang dijalankan, hasil test, dan blocker riil. Akhir proyek: tampilkan daftar URL halaman (lokal/staging sesuai yang benar-benar tersedia), cara login developer, cara setup email/queue, cara memasukkan saldo awal dan rekening, dan cara melakukan 1 siklus booking->kontrak->tagihan->transfer->verifikasi->laporan->tutup buku. Jika belum benar-benar dipasang di server, jangan klaim live.

---

## LAMPIRAN: DI LUAR SCOPE V1 DAN JANGAN DIBANGUN TANPA PERMINTAAN

Aplikasi native Android/iOS, WhatsApp bot, integrasi bank otomatis/open banking, integrasi PDAM/PLN secara API, multi-staf/role pengelola/teknisi, deposit sewa, denda keterlambatan otomatis, OCR pembayaran wajib, AI prediksi pendapatan, marketplace kos multi-pemilik publik, pelacakan meter kamar manual, dan pembayaran biaya air/token kamar kepada pemilik. Fokus pada ketentuan kos sebenarnya.

**Jika butuh pilihan produk yang tak ditentukan:** tulis default yang dipilih dan kenapa; pertahankan opsi yang bisa dikonfigurasi (bukan hardcode). Jangan bertanya ulang untuk informasi yang sudah ada di dokumen ini.
