package httpapp

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"mime"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"alam-segar/kos-management/internal/auth"
	"alam-segar/kos-management/internal/config"
)

//go:embed templates/*.html
var templateFiles embed.FS

type contextKey string

const sessionContextKey contextKey = "session"

type App struct {
	cfg           config.Config
	db            *sql.DB
	sessions      *auth.SessionStore
	templates     *template.Template
	dummyPassword string
	location      *time.Location
}

type PageData struct {
	Title           string
	CurrentPath     string
	CSRFToken       string
	CSPNonce        string
	User            *auth.User
	Error           string
	Success         string
	Property        PublicProperty
	Rooms           []PublicRoom
	Metrics         OwnerMetrics
	DashboardCharts []OwnerChartPoint
	Bookings        []RecentBooking
	BookingForm     BookingForm
	OwnerBookings   []OwnerBooking
	Properties      []OwnerProperty
	PropertyForm    PropertyForm
	RoomsAdmin      []OwnerRoom
	RoomForm        RoomForm
	PropertyOptions []PropertyOption
	RoomDetail      PublicRoomDetail
	Payload         any
	DevLink         string
}

type PublicProperty struct {
	Name        string
	Description string
	Phone       string
}

type PublicRoom struct {
	ID           int64
	PropertySlug string
	Number       string
	Slug         string
	TypeName     string
	MonthlyRate  int64
	Description  string
	CoverPhoto   string
}

type PublicRoomDetail struct {
	ID           int64
	PropertyName string
	PropertySlug string
	RoomNumber   string
	RoomSlug     string
	TypeName     string
	MonthlyRate  int64
	AnnualRate   int64
	Description  string
	Amenities    []string
	Photos       []string
}

type OwnerMetrics struct {
	TotalRooms       int64
	OccupiedRooms    int64
	AvailableRooms   int64
	Maintenance      int64
	PendingBookings  int64
	PendingPayments  int64
	OverdueInvoices  int64
	OpenTickets      int64
	Receivables      int64
	Cash             int64
	SmartActions     int64
	OpenEmergencies  int64
	WaitingParcels   int64
	PendingDocuments int64
}

type OwnerChartPoint struct {
	Label             string
	Revenue, Expenses int64
}

type RecentBooking struct {
	Reference string
	Name      string
	Room      string
	StartDate string
	Status    string
}

type BookingForm struct {
	Name          string
	Email         string
	Phone         string
	RoomID        string
	StartDate     string
	EndDate       string
	Message       string
	Reference     string
	SubmissionKey string
}

type OwnerBooking struct {
	ID            int64
	Reference     string
	ApplicantName string
	Email         string
	Phone         string
	Room          string
	StartDate     string
	EndDate       string
	Status        string
	BookingAmount int64
	PaymentID     int64
	PaymentStatus string
	TenantID      int64
	AccountReady  bool
}

func New(cfg config.Config, db *sql.DB) (*App, error) {
	location, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		return nil, fmt.Errorf("load Asia/Makassar: %w", err)
	}

	funcs := template.FuncMap{
		"rupiah":        formatRupiah,
		"itoa":          func(value int64) string { return strconv.FormatInt(value, 10) },
		"sub":           func(a, b int64) int64 { return a - b },
		"requestLabel":  smartRequestLabel,
		"smartCategory": smartCategoryLabel,
		"documentType":  documentTypeLabel,
		"tanggal": func(value time.Time) string {
			if value.IsZero() {
				return "-"
			}
			return value.In(location).Format("02 Jan 2006")
		},
		"statusLabel": func(status string) string {
			labels := map[string]string{
				"submitted":                "Diajukan",
				"under_review":             "Ditinjau",
				"approved_waiting_payment": "Menunggu Pembayaran",
				"proof_submitted":          "Bukti Dikirim",
				"confirmed":                "Terkonfirmasi",
				"converted_to_contract":    "Menjadi Kontrak",
				"rejected":                 "Ditolak",
				"cancelled":                "Dibatalkan",
				"released":                 "Dilepas",
				"refund_pending":           "Menunggu Refund",
				"refunded":                 "Direfund",
				"draft":                    "Draf",
				"active":                   "Aktif",
				"expiring":                 "Akan Berakhir",
				"ended":                    "Berakhir",
				"terminated":               "Dihentikan",
				"issued":                   "Terbit",
				"partially_paid":           "Dibayar Sebagian",
				"paid":                     "Lunas",
				"overdue":                  "Terlambat",
				"void":                     "Dibatalkan",
				"pending_verification":     "Menunggu Verifikasi",
				"verified":                 "Terverifikasi",
				"pending":                  "Menunggu Persetujuan",
				"approved":                 "Disetujui",
				"checked_in":               "Sudah Masuk",
				"checked_out":              "Sudah Keluar",
				"expired":                  "Kedaluwarsa",
				"completed":                "Selesai",
				"reviewing":                "Sedang Ditinjau",
				"open":                     "Terbuka",
				"acknowledged":             "Sudah Ditanggapi",
				"resolved":                 "Selesai",
				"collected":                "Sudah Diambil",
				"returned":                 "Dikembalikan",
				"arrived":                  "Sudah Tiba",
				"notified":                 "Penghuni Diberi Tahu",
				"awaiting_tenant":          "Menunggu Tanda Tangan Penghuni",
				"owner_signed":             "Ditandatangani Pemilik",
			}
			if label, ok := labels[status]; ok {
				return label
			}
			return status
		},
	}
	templates, err := template.New("pages").Funcs(funcs).ParseFS(templateFiles, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}

	dummyPassword, err := auth.HashPassword("Bukan-Kata-Sandi-123!")
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:           cfg,
		db:            db,
		sessions:      auth.NewSessionStore(db, cfg.SecureCookie, cfg.HashKey, cfg.TrustedProxyCIDRs),
		templates:     templates,
		dummyPassword: dummyPassword,
		location:      location,
	}, nil
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /assets/admin/", a.assetHandler("/assets/admin/", a.cfg.AdminAssetsRoot,
		map[string]bool{"css": true, "js": true, "images": true, "icons": true, "vendor": true, "audio": true}))
	mux.Handle("GET /assets/public/", a.assetHandler("/assets/public/", a.cfg.PublicAssetsRoot,
		map[string]bool{"css": true, "js": true, "images": true, "fonts": true, "video": true}))
	mux.Handle("GET /uploads/", a.assetHandler("/uploads/", a.cfg.PublicFilesRoot,
		map[string]bool{"rooms": true}))
	mux.HandleFunc("GET /manifest.webmanifest", a.pwaFile("manifest.webmanifest", "application/manifest+json; charset=utf-8"))
	mux.HandleFunc("GET /service-worker.js", a.pwaFile("service-worker.js", "application/javascript; charset=utf-8"))
	mux.HandleFunc("GET /pwa/offline.html", a.pwaFile("offline.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /pwa/icon.svg", a.pwaFile("icon.svg", "image/svg+xml"))
	mux.HandleFunc("GET /pwa/icon-maskable.svg", a.pwaFile("icon-maskable.svg", "image/svg+xml"))

	mux.HandleFunc("GET /", a.home)
	mux.HandleFunc("GET /kamar", a.rooms)
	mux.HandleFunc("GET /properti/{propertySlug}/kamar/{roomSlug}", a.roomDetail)
	mux.HandleFunc("GET /booking", a.bookingForm)
	mux.HandleFunc("POST /booking", a.submitBooking)
	mux.HandleFunc("GET /booking/pembayaran", a.bookingPaymentForm)
	mux.HandleFunc("POST /booking/pembayaran", a.submitBookingPayment)
	mux.HandleFunc("GET /daftar", a.registrationForm)
	mux.HandleFunc("POST /daftar", a.registerTenant)
	mux.HandleFunc("GET /verifikasi-email", a.verifyEmail)
	mux.HandleFunc("GET /masuk", a.loginForm)
	mux.HandleFunc("POST /masuk", a.login)
	mux.HandleFunc("GET /lupa-password", a.forgotPasswordForm)
	mux.HandleFunc("POST /lupa-password", a.forgotPassword)
	mux.HandleFunc("GET /reset-password", a.resetPasswordForm)
	mux.HandleFunc("POST /reset-password", a.resetPassword)
	mux.HandleFunc("GET /verifikasi-dokumen/{code}", a.verifyDocument)
	mux.HandleFunc("POST /keluar", a.logout)
	mux.HandleFunc("GET /pemilik", a.ownerDashboard)
	mux.HandleFunc("GET /pemilik/properti", a.ownerPropertyList)
	mux.HandleFunc("GET /pemilik/properti/baru", a.ownerPropertyCreateForm)
	mux.HandleFunc("POST /pemilik/properti/baru", a.ownerPropertyCreate)
	mux.HandleFunc("GET /pemilik/properti/{id}/ubah", a.ownerPropertyEditForm)
	mux.HandleFunc("POST /pemilik/properti/{id}/ubah", a.ownerPropertyUpdate)
	mux.HandleFunc("GET /pemilik/kamar", a.ownerRoomList)
	mux.HandleFunc("GET /pemilik/kamar/baru", a.ownerRoomCreateForm)
	mux.HandleFunc("POST /pemilik/kamar/baru", a.ownerRoomCreate)
	mux.HandleFunc("GET /pemilik/kamar/{id}/ubah", a.ownerRoomEditForm)
	mux.HandleFunc("POST /pemilik/kamar/{id}/ubah", a.ownerRoomUpdate)
	mux.HandleFunc("POST /pemilik/kamar/{id}/foto/{photoID}/hapus", a.ownerRoomDeletePhoto)
	mux.HandleFunc("GET /pemilik/booking", a.ownerBookingList)
	mux.HandleFunc("POST /pemilik/booking/{id}/setujui", a.approveBooking)
	mux.HandleFunc("POST /pemilik/booking/{id}/tolak", a.rejectBooking)
	mux.HandleFunc("POST /pemilik/booking/{id}/lepaskan", a.releaseBooking)
	mux.HandleFunc("POST /pemilik/booking/{id}/refund", a.refundBooking)
	mux.HandleFunc("POST /pemilik/booking/{id}/buat-akun", a.provisionBookingTenant)
	mux.HandleFunc("POST /pemilik/booking-pembayaran/{id}/verifikasi", a.verifyBookingPayment)
	mux.HandleFunc("POST /pemilik/booking-pembayaran/{id}/tolak", a.rejectBookingPayment)
	mux.HandleFunc("GET /pemilik/penghuni", a.ownerTenantList)
	mux.HandleFunc("GET /pemilik/penghuni/baru", a.ownerTenantCreateForm)
	mux.HandleFunc("POST /pemilik/penghuni/baru", a.ownerTenantCreate)
	mux.HandleFunc("POST /pemilik/registrasi/{id}/setujui", a.ownerApproveRegistration)
	mux.HandleFunc("POST /pemilik/registrasi/{id}/tolak", a.ownerRejectRegistration)
	mux.HandleFunc("GET /pemilik/kontrak", a.ownerLeaseList)
	mux.HandleFunc("GET /pemilik/kontrak/baru", a.ownerLeaseCreateForm)
	mux.HandleFunc("POST /pemilik/kontrak/baru", a.ownerLeaseCreate)
	mux.HandleFunc("GET /pemilik/kontrak/{id}", a.ownerLeaseDetail)
	mux.HandleFunc("POST /pemilik/kontrak/{id}/check-in", a.ownerLeaseCheckIn)
	mux.HandleFunc("POST /pemilik/kontrak/{id}/check-out", a.ownerLeaseCheckOut)
	mux.HandleFunc("POST /pemilik/kontrak/{id}/terminasi", a.ownerLeaseTerminate)
	mux.HandleFunc("POST /pemilik/kontrak/{id}/perpanjang", a.ownerLeaseExtend)
	mux.HandleFunc("POST /pemilik/kontrak/{id}/kerusakan", a.ownerDamageCharge)
	mux.HandleFunc("GET /pemilik/tagihan", a.ownerInvoiceList)
	mux.HandleFunc("POST /pemilik/tagihan/{id}/batalkan", a.ownerVoidInvoice)
	mux.HandleFunc("POST /pemilik/tagihan/{id}/pakai-kredit", a.ownerApplyTenantCredit)
	mux.HandleFunc("GET /pemilik/pembayaran", a.ownerPaymentList)
	mux.HandleFunc("POST /pemilik/pembayaran/{id}/verifikasi", a.ownerVerifyPayment)
	mux.HandleFunc("POST /pemilik/pembayaran/{id}/tolak", a.ownerRejectPayment)
	mux.HandleFunc("POST /pemilik/pembayaran/{id}/refund", a.ownerRefundPayment)
	mux.HandleFunc("POST /pemilik/kredit/{id}/refund", a.ownerRefundTenantCredit)
	mux.HandleFunc("GET /pemilik/keuangan", a.ownerFinance)
	mux.HandleFunc("POST /pemilik/keuangan/rekening", a.ownerCreateBankAccount)
	mux.HandleFunc("POST /pemilik/keuangan/kas", a.ownerCreateCashAccount)
	mux.HandleFunc("POST /pemilik/keuangan/saldo-awal", a.ownerPostOpeningBalance)
	mux.HandleFunc("POST /pemilik/keuangan/transfer", a.ownerInternalTransfer)
	mux.HandleFunc("POST /pemilik/keuangan/rekonsiliasi", a.ownerReconcileAccount)
	mux.HandleFunc("GET /pemilik/pengeluaran", a.ownerExpenseList)
	mux.HandleFunc("POST /pemilik/pengeluaran", a.ownerCreateExpense)
	mux.HandleFunc("GET /pemilik/akuntansi", a.ownerAccounting)
	mux.HandleFunc("POST /pemilik/akuntansi/tutup-buku", a.ownerClosePeriod)
	mux.HandleFunc("GET /pemilik/laporan", a.ownerReports)
	mux.HandleFunc("GET /pemilik/laporan/jurnal.csv", a.ownerJournalCSV)
	mux.HandleFunc("GET /pemilik/inventaris", a.ownerAssetList)
	mux.HandleFunc("POST /pemilik/inventaris", a.ownerCreateAsset)
	mux.HandleFunc("GET /pemilik/keluhan", a.ownerTicketList)
	mux.HandleFunc("POST /pemilik/keluhan/{id}/status", a.ownerUpdateTicket)
	mux.HandleFunc("GET /pemilik/pengumuman", a.ownerAnnouncementList)
	mux.HandleFunc("POST /pemilik/pengumuman", a.ownerCreateAnnouncement)
	mux.HandleFunc("GET /pemilik/smart-kos", a.ownerSmartLiving)
	mux.HandleFunc("POST /pemilik/smart-kos/fasilitas", a.ownerCreateFacility)
	mux.HandleFunc("POST /pemilik/smart-kos/reservasi/{id}/status", a.ownerUpdateFacilityReservation)
	mux.HandleFunc("POST /pemilik/smart-kos/tamu/{id}/status", a.ownerUpdateVisitorPass)
	mux.HandleFunc("POST /pemilik/smart-kos/paket", a.ownerCreateParcel)
	mux.HandleFunc("POST /pemilik/smart-kos/paket/{id}/status", a.ownerUpdateParcel)
	mux.HandleFunc("POST /pemilik/smart-kos/permintaan/{id}/status", a.ownerUpdateResidentRequest)
	mux.HandleFunc("POST /pemilik/smart-kos/darurat/{id}/status", a.ownerUpdateEmergencyAlert)
	mux.HandleFunc("GET /pemilik/dokumen", a.ownerDocuments)
	mux.HandleFunc("POST /pemilik/dokumen/terbitkan", a.ownerIssueDocument)
	mux.HandleFunc("GET /pemilik/dokumen/{id}", a.ownerDocumentDetail)
	mux.HandleFunc("POST /pemilik/dokumen/{id}/tanda-tangan", a.ownerSignDocument)
	mux.HandleFunc("POST /pemilik/dokumen/{id}/batalkan", a.ownerVoidDocument)
	mux.HandleFunc("GET /pemilik/pengaturan", a.ownerSettings)
	mux.HandleFunc("POST /pemilik/pengaturan", a.ownerSettingsUpdate)
	mux.HandleFunc("GET /pemilik/file/{kind}/{id}", a.ownerPrivateFile)
	mux.HandleFunc("GET /penghuni", a.tenantDashboard)
	mux.HandleFunc("GET /penghuni/kamar", a.tenantRoom)
	mux.HandleFunc("GET /penghuni/kontrak", a.tenantLeases)
	mux.HandleFunc("GET /penghuni/tagihan", a.tenantInvoices)
	mux.HandleFunc("POST /penghuni/tagihan/{id}/konfirmasi-transfer", a.tenantSubmitPayment)
	mux.HandleFunc("GET /penghuni/kuitansi/{id}", a.tenantReceipt)
	mux.HandleFunc("GET /penghuni/keluhan", a.tenantTickets)
	mux.HandleFunc("POST /penghuni/keluhan", a.tenantCreateTicket)
	mux.HandleFunc("POST /penghuni/keluhan/{id}/balas", a.tenantReplyTicket)
	mux.HandleFunc("GET /penghuni/notifikasi", a.tenantNotifications)
	mux.HandleFunc("GET /penghuni/pengumuman", a.tenantAnnouncements)
	mux.HandleFunc("GET /penghuni/layanan", a.tenantSmartLiving)
	mux.HandleFunc("POST /penghuni/layanan/fasilitas", a.tenantCreateFacilityReservation)
	mux.HandleFunc("POST /penghuni/layanan/reservasi/{id}/batal", a.tenantCancelFacilityReservation)
	mux.HandleFunc("POST /penghuni/layanan/tamu", a.tenantCreateVisitorPass)
	mux.HandleFunc("POST /penghuni/layanan/tamu/{id}/batal", a.tenantCancelVisitorPass)
	mux.HandleFunc("POST /penghuni/layanan/paket/{id}/ambil", a.tenantCollectParcel)
	mux.HandleFunc("POST /penghuni/layanan/permintaan", a.tenantCreateResidentRequest)
	mux.HandleFunc("POST /penghuni/layanan/permintaan/{id}/batal", a.tenantCancelResidentRequest)
	mux.HandleFunc("POST /penghuni/layanan/darurat", a.tenantCreateEmergencyAlert)
	mux.HandleFunc("GET /penghuni/dokumen", a.tenantDocuments)
	mux.HandleFunc("GET /penghuni/dokumen/{id}", a.tenantDocumentDetail)
	mux.HandleFunc("POST /penghuni/dokumen/{id}/tanda-tangan", a.tenantSignDocument)
	mux.HandleFunc("POST /penghuni/notifikasi/{id}/baca", a.tenantReadNotification)
	mux.HandleFunc("GET /penghuni/profil", a.tenantProfile)
	mux.HandleFunc("POST /penghuni/profil", a.tenantUpdateProfile)
	mux.HandleFunc("POST /penghuni/profil/password", a.tenantChangePassword)
	mux.HandleFunc("GET /penghuni/file/{kind}/{id}", a.tenantPrivateFile)
	mux.HandleFunc("GET /kesehatan", a.health)

	return a.recoverer(
		a.requestDeadline(
			a.requestLimits(
				a.gzipResponses(
					a.securityHeaders(
						a.publicRateLimit(
							a.loadSession(mux),
						),
					),
				),
			),
		),
	)
}

func (a *App) bookingForm(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if session == nil {
		var err error
		session, err = a.sessions.Start(r.Context(), w, nil)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	submissionKey, err := randomKey(32)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.renderBooking(w, r, http.StatusOK, PageData{Title: "Booking Kamar — Alam Segar", CSRFToken: session.CSRFToken, BookingForm: BookingForm{SubmissionKey: submissionKey}})
}

func (a *App) submitBooking(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid. Muat ulang halaman dan coba lagi.", http.StatusBadRequest)
		return
	}

	form := BookingForm{
		Name:          strings.TrimSpace(r.FormValue("name")),
		Email:         strings.ToLower(strings.TrimSpace(r.FormValue("email"))),
		Phone:         strings.TrimSpace(r.FormValue("phone")),
		RoomID:        strings.TrimSpace(r.FormValue("room_id")),
		StartDate:     strings.TrimSpace(r.FormValue("start_date")),
		EndDate:       strings.TrimSpace(r.FormValue("end_date")),
		Message:       strings.TrimSpace(r.FormValue("message")),
		SubmissionKey: strings.TrimSpace(r.FormValue("submission_key")),
	}
	data := PageData{Title: "Booking Kamar — Alam Segar", CSRFToken: session.CSRFToken, BookingForm: form}

	roomID, parseErr := strconv.ParseInt(form.RoomID, 10, 64)
	parsedEmail, emailErr := mail.ParseAddress(form.Email)
	startDate, startErr := parseBusinessDate(form.StartDate)
	endDate, endErr := parseBusinessDate(form.EndDate)
	today := businessToday(a.location)
	switch {
	case len(form.SubmissionKey) != 64:
		data.Error = "Form sudah kedaluwarsa. Muat ulang halaman booking."
	case len([]rune(form.Name)) < 3 || len([]rune(form.Name)) > 150:
		data.Error = "Nama harus berisi 3 sampai 150 karakter."
	case emailErr != nil || parsedEmail.Address != form.Email || len(form.Email) > 190:
		data.Error = "Alamat email tidak valid."
	case len(form.Phone) < 8 || len(form.Phone) > 40:
		data.Error = "Nomor telepon harus berisi 8 sampai 40 karakter."
	case parseErr != nil || roomID <= 0:
		data.Error = "Pilih kamar yang ingin diajukan."
	case startErr != nil || endErr != nil:
		data.Error = "Tanggal rencana masuk dan selesai tidak valid."
	case startDate.Before(today):
		data.Error = "Tanggal rencana masuk tidak boleh berada di masa lalu."
	case !endDate.After(startDate):
		data.Error = "Tanggal selesai harus setelah tanggal rencana masuk."
	case len([]rune(form.Message)) > 2000:
		data.Error = "Catatan maksimal 2.000 karakter."
	}
	if data.Error != "" {
		a.renderBooking(w, r, http.StatusUnprocessableEntity, data)
		return
	}

	ipHash := a.sessions.HashIP(r)
	var recentAttempts int
	if err := a.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM booking_request_attempts
		WHERE email_normalized = ? AND ip_hash = ?
		  AND attempted_at >= UTC_TIMESTAMP(6) - INTERVAL 1 HOUR`, form.Email, ipHash).
		Scan(&recentAttempts); err != nil {
		a.serverError(w, r, err)
		return
	}
	if recentAttempts >= 3 {
		data.Error = "Terlalu banyak pengajuan. Tunggu satu jam sebelum mencoba lagi."
		a.renderBooking(w, r, http.StatusTooManyRequests, data)
		return
	}

	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	_, _ = tx.ExecContext(r.Context(), `
		INSERT INTO booking_request_attempts (email_normalized, ip_hash, was_accepted)
		VALUES (?, ?, 0)`, form.Email, ipHash)

	var propertyID int64
	err = tx.QueryRowContext(r.Context(), `
		SELECT r.property_id
		FROM rooms r JOIN properties p ON p.id = r.property_id
		WHERE r.id = ? AND r.is_published = 1 AND r.condition_status = 'ready' AND r.deleted_at IS NULL
		  AND p.is_published = 1 AND p.status = 'active' AND p.deleted_at IS NULL
		FOR UPDATE`, roomID).Scan(&propertyID)
	if errors.Is(err, sql.ErrNoRows) {
		data.Error = "Kamar tidak tersedia untuk pengajuan publik."
		a.renderBooking(w, r, http.StatusConflict, data)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	reference, err := bookingReference()
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var applicantUserID any
	if session.User != nil && session.User.Role == "tenant" {
		applicantUserID = session.User.ID
	}
	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO booking_requests
			(reference_number, public_submission_key, property_id, room_id, applicant_user_id, applicant_name, applicant_email,
			 applicant_phone, planned_start_date, planned_end_date, status, owner_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'submitted', ?)`,
		reference, form.SubmissionKey, propertyID, roomID, applicantUserID, form.Name, form.Email, form.Phone,
		startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), nullableString(form.Message))
	if err != nil {
		if isDuplicate(err) {
			_ = tx.Rollback()
			var existingReference string
			if queryErr := a.db.QueryRowContext(r.Context(), `SELECT reference_number FROM booking_requests WHERE public_submission_key=?`, form.SubmissionKey).Scan(&existingReference); queryErr == nil {
				data.BookingForm.Reference = existingReference
				data.Success = "Pengajuan booking sudah diterima sebelumnya. Tidak dibuat pengajuan ganda."
				a.renderBooking(w, r, http.StatusOK, data)
				return
			}
		}
		a.serverError(w, r, err)
		return
	}
	_, _ = tx.ExecContext(r.Context(), `
		UPDATE booking_request_attempts SET was_accepted = 1
		WHERE email_normalized = ? AND ip_hash = ? ORDER BY id DESC LIMIT 1`, form.Email, ipHash)
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}

	data.BookingForm.Reference = reference
	data.Success = "Pengajuan booking berhasil dikirim. Pemilik akan meninjaunya sebelum kamar ditahan."
	a.renderBooking(w, r, http.StatusCreated, data)
}

func (a *App) ownerBookingList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	a.renderOwnerBookingList(w, r, http.StatusOK, session, "")
}

func (a *App) approveBooking(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	bookingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	amount, amountErr := strconv.ParseInt(strings.TrimSpace(r.FormValue("booking_amount")), 10, 64)
	usage := r.FormValue("booking_usage")
	if err != nil || bookingID <= 0 || amountErr != nil || amount <= 0 || (usage != "credit_first_rent" && usage != "additional_booking_fee") {
		a.renderOwnerBookingList(w, r, http.StatusUnprocessableEntity, session, "Nominal dan penggunaan uang booking wajib diisi dengan benar.")
		return
	}

	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var roomID int64
	var startDate, endDate time.Time
	var currentStatus, applicantEmail, reference string
	err = tx.QueryRowContext(r.Context(), `
		SELECT b.room_id, b.planned_start_date, b.planned_end_date, b.status,b.applicant_email,b.reference_number
		FROM booking_requests b JOIN properties p ON p.id = b.property_id
		WHERE b.id = ? AND p.owner_id = ? FOR UPDATE`, bookingID, session.User.ID).
		Scan(&roomID, &startDate, &endDate, &currentStatus, &applicantEmail, &reference)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if currentStatus != "submitted" && currentStatus != "under_review" {
		a.renderOwnerBookingList(w, r, http.StatusConflict, session, "Booking ini sudah diproses dan tidak dapat disetujui ulang.")
		return
	}
	var lockedRoom int64
	if err := tx.QueryRowContext(r.Context(), `SELECT id FROM rooms WHERE id = ? AND condition_status = 'ready' FOR UPDATE`, roomID).Scan(&lockedRoom); err != nil {
		a.renderOwnerBookingList(w, r, http.StatusConflict, session, "Kamar tidak berada dalam kondisi siap disewakan.")
		return
	}
	var conflicts int
	if err := tx.QueryRowContext(r.Context(), `
		SELECT
			(SELECT COUNT(*) FROM leases
			 WHERE room_id = ? AND status IN ('draft', 'active', 'expiring')
			   AND start_date < ? AND end_date > ?)
			+
			(SELECT COUNT(*) FROM booking_requests
			 WHERE room_id = ? AND id <> ?
			   AND status IN ('approved_waiting_payment', 'proof_submitted', 'confirmed', 'converted_to_contract')
			   AND planned_start_date < ? AND planned_end_date > ?)`,
		roomID, endDate, startDate, roomID, bookingID, endDate, startDate).Scan(&conflicts); err != nil {
		a.serverError(w, r, err)
		return
	}
	if conflicts > 0 {
		a.renderOwnerBookingList(w, r, http.StatusConflict, session, "Kamar sudah terikat booking atau kontrak lain pada periode tersebut.")
		return
	}
	result, err := tx.ExecContext(r.Context(), `
		UPDATE booking_requests
		SET status = 'approved_waiting_payment', booking_amount = ?, booking_usage = ?,
			hold_started_at = UTC_TIMESTAMP(6), reviewed_by = ?, reviewed_at = UTC_TIMESTAMP(6), owner_reason = NULL
		WHERE id = ? AND status IN ('submitted', 'under_review')`, amount, usage, session.User.ID, bookingID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		a.renderOwnerBookingList(w, r, http.StatusConflict, session, "Booking berubah saat diproses. Muat ulang halaman.")
		return
	}
	paymentURL := strings.TrimRight(a.cfg.BaseURL, "/") + "/booking/pembayaran?referensi=" + url.QueryEscape(reference) + "&email=" + url.QueryEscape(applicantEmail)
	usageLabel := "mengurangi tagihan sewa pertama"
	if usage == "additional_booking_fee" {
		usageLabel = "biaya booking tambahan"
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (NULL,'email',?,?,?,?)`, fmt.Sprintf("booking.approved.%d", bookingID), applicantEmail, "Booking Alam Segar disetujui", fmt.Sprintf("Booking %s disetujui. Uang booking %s akan dipakai sebagai %s. Unggah bukti transfer melalui %s. Bukti tetap diverifikasi manual oleh pemilik.", reference, formatRupiah(amount), usageLabel, paymentURL)); err != nil {
		a.serverError(w, r, err)
		return
	}
	_, _ = tx.ExecContext(r.Context(), `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, after_json, ip_hash)
		VALUES (?, 'booking.approve', 'booking_request', ?, ?, ?)`, session.User.ID,
		strconv.FormatInt(bookingID, 10), fmt.Sprintf(`{"status":"approved_waiting_payment","amount":%d,"usage":%q}`, amount, usage), a.sessions.HashIP(r))
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=disetujui", http.StatusSeeOther)
}

func (a *App) rejectBooking(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	bookingID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	reason := strings.TrimSpace(r.FormValue("reason"))
	if err != nil || bookingID <= 0 || len([]rune(reason)) < 5 {
		a.renderOwnerBookingList(w, r, http.StatusUnprocessableEntity, session, "Alasan penolakan minimal 5 karakter.")
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var applicantEmail, reference, currentStatus string
	err = tx.QueryRowContext(r.Context(), `SELECT b.applicant_email,b.reference_number,b.status FROM booking_requests b JOIN properties p ON p.id=b.property_id WHERE b.id=? AND p.owner_id=? FOR UPDATE`, bookingID, session.User.ID).Scan(&applicantEmail, &reference, &currentStatus)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if currentStatus != "submitted" && currentStatus != "under_review" {
		a.renderOwnerBookingList(w, r, http.StatusConflict, session, "Booking tidak ditemukan atau sudah diproses.")
		return
	}
	result, err := tx.ExecContext(r.Context(), `UPDATE booking_requests SET status='rejected',owner_reason=?,reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6) WHERE id=? AND status IN ('submitted','under_review')`, reason, session.User.ID, bookingID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected != 1 {
		a.renderOwnerBookingList(w, r, http.StatusConflict, session, "Booking tidak ditemukan atau sudah diproses.")
		return
	}
	if _, err := tx.ExecContext(r.Context(), `INSERT INTO notification_deliveries (user_id,channel,event_key,recipient,subject,body) VALUES (NULL,'email',?,?,?,?)`, fmt.Sprintf("booking.rejected.%d", bookingID), applicantEmail, "Booking Alam Segar belum dapat disetujui", "Booking "+reference+" ditolak. Alasan pemilik: "+reason); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "booking.reject", "booking_request", bookingID, map[string]any{"status": currentStatus}, map[string]any{"status": "rejected", "reason": reason}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/booking?berhasil=ditolak", http.StatusSeeOther)
}

func (a *App) renderBooking(w http.ResponseWriter, r *http.Request, status int, data PageData) {
	data.Property = a.publicProperty(r.Context())
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT r.id, p.slug, r.number, r.slug, COALESCE(rt.name, 'Kamar'), r.monthly_rate, COALESCE(r.description, ''),
			COALESCE((SELECT rp.storage_path FROM room_photos rp WHERE rp.room_id=r.id AND rp.is_public=1 ORDER BY rp.sort_order,rp.id LIMIT 1),'')
		FROM rooms r
		JOIN properties p ON p.id = r.property_id
		LEFT JOIN room_types rt ON rt.id = r.room_type_id
		WHERE r.is_published = 1 AND r.condition_status = 'ready' AND r.deleted_at IS NULL
		  AND p.is_published = 1 AND p.status = 'active' AND p.deleted_at IS NULL
		ORDER BY p.name, r.number`)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var room PublicRoom
		if err := rows.Scan(&room.ID, &room.PropertySlug, &room.Number, &room.Slug, &room.TypeName, &room.MonthlyRate, &room.Description, &room.CoverPhoto); err != nil {
			a.serverError(w, r, err)
			return
		}
		data.Rooms = append(data.Rooms, room)
	}
	a.render(w, status, "public_booking", data)
}

func (a *App) renderOwnerBookingList(w http.ResponseWriter, r *http.Request, status int, session *auth.Session, errorMessage string) {
	finance, _ := a.loadFinancePage(r.Context(), session.User.ID)
	data := PageData{Title: "Booking — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Error: errorMessage, Payload: finance}
	if r.URL.Query().Get("berhasil") == "disetujui" {
		data.Success = "Booking disetujui dan kamar ditahan secara manual."
	} else if r.URL.Query().Get("berhasil") == "ditolak" {
		data.Success = "Booking ditolak dengan alasan yang tercatat."
	} else if r.URL.Query().Get("berhasil") == "dana" {
		data.Success = "Dana booking diverifikasi dan dicatat ke ledger. Hubungkan akun penghuni untuk melanjutkan ke kontrak."
	} else if r.URL.Query().Get("berhasil") == "akun" {
		data.Success = "Akun penghuni sudah dihubungkan. Jika email belum terverifikasi, tautan aktivasi 24 jam telah masuk antrean email."
	}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT b.id, b.reference_number, b.applicant_name, b.applicant_email, b.applicant_phone,
			r.number, b.planned_start_date, b.planned_end_date, b.status, COALESCE(b.booking_amount,0),
			COALESCE(bps.id,0), COALESCE(bps.status,''), COALESCE(b.applicant_user_id,0),
			CASE WHEN u.id IS NOT NULL AND u.role='tenant' AND u.status='active' AND u.email_verified_at IS NOT NULL THEN 1 ELSE 0 END
		FROM booking_requests b
		JOIN properties p ON p.id = b.property_id
		JOIN rooms r ON r.id = b.room_id
		LEFT JOIN booking_payment_submissions bps ON bps.id=(SELECT MAX(x.id) FROM booking_payment_submissions x WHERE x.booking_request_id=b.id)
		LEFT JOIN users u ON u.id=b.applicant_user_id
		WHERE p.owner_id = ?
		ORDER BY FIELD(b.status, 'submitted', 'under_review', 'proof_submitted', 'approved_waiting_payment', 'confirmed'), b.created_at DESC`,
		session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item OwnerBooking
		var startDate, endDate time.Time
		if err := rows.Scan(&item.ID, &item.Reference, &item.ApplicantName, &item.Email, &item.Phone,
			&item.Room, &startDate, &endDate, &item.Status, &item.BookingAmount, &item.PaymentID, &item.PaymentStatus,
			&item.TenantID, &item.AccountReady); err != nil {
			a.serverError(w, r, err)
			return
		}
		item.StartDate = startDate.Format("02 Jan 2006")
		item.EndDate = endDate.Format("02 Jan 2006")
		data.OwnerBookings = append(data.OwnerBookings, item)
	}
	a.render(w, status, "owner_booking", data)
}

func bookingReference() (string, error) {
	buffer := make([]byte, 5)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return "BKG-" + time.Now().UTC().Format("20060102") + "-" + strings.ToUpper(hex.EncodeToString(buffer)), nil
}

func randomKey(byteLength int) (string, error) {
	buffer := make([]byte, byteLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := PageData{Title: "Alam Segar", CurrentPath: r.URL.Path}
	data.Property = a.publicProperty(r.Context())
	rooms, err := a.publicRooms(r.Context(), 3)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data.Rooms = rooms
	a.render(w, http.StatusOK, "public_home", data)
}

func (a *App) rooms(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Kamar — Alam Segar", CurrentPath: r.URL.Path}
	data.Property = a.publicProperty(r.Context())
	rooms, err := a.publicRooms(r.Context(), 0)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data.Rooms = rooms
	a.render(w, http.StatusOK, "public_rooms", data)
}

func (a *App) publicRooms(ctx context.Context, limit int) ([]PublicRoom, error) {
	query := `
		SELECT r.id, p.slug, r.number, r.slug, COALESCE(rt.name, 'Kamar'), r.monthly_rate, COALESCE(r.description, ''),
			COALESCE((SELECT rp.storage_path FROM room_photos rp WHERE rp.room_id=r.id AND rp.is_public=1 ORDER BY rp.sort_order,rp.id LIMIT 1),'')
		FROM rooms r
		JOIN properties p ON p.id = r.property_id
		LEFT JOIN room_types rt ON rt.id = r.room_type_id
		WHERE r.is_published = 1 AND r.condition_status = 'ready' AND r.deleted_at IS NULL
		  AND p.is_published = 1 AND p.status = 'active' AND p.deleted_at IS NULL
		ORDER BY p.name, r.number`
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}
	rows, err := a.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []PublicRoom
	for rows.Next() {
		var room PublicRoom
		if err := rows.Scan(&room.ID, &room.PropertySlug, &room.Number, &room.Slug, &room.TypeName, &room.MonthlyRate, &room.Description, &room.CoverPhoto); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}
	return rooms, rows.Err()
}

func (a *App) roomDetail(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Detail Kamar — Alam Segar", CurrentPath: r.URL.Path}
	var annual sql.NullInt64
	err := a.db.QueryRowContext(r.Context(), `
		SELECT rm.id,p.name, p.slug, rm.number, rm.slug, COALESCE(rt.name, 'Kamar'), rm.monthly_rate,
			rm.annual_rate, COALESCE(rm.description, '')
		FROM rooms rm
		JOIN properties p ON p.id = rm.property_id
		LEFT JOIN room_types rt ON rt.id = rm.room_type_id
		WHERE p.slug = ? AND rm.slug = ?
		  AND p.is_published = 1 AND p.status = 'active' AND p.deleted_at IS NULL
		  AND rm.is_published = 1 AND rm.condition_status = 'ready' AND rm.deleted_at IS NULL`,
		r.PathValue("propertySlug"), r.PathValue("roomSlug")).
		Scan(&data.RoomDetail.ID, &data.RoomDetail.PropertyName, &data.RoomDetail.PropertySlug, &data.RoomDetail.RoomNumber,
			&data.RoomDetail.RoomSlug, &data.RoomDetail.TypeName, &data.RoomDetail.MonthlyRate,
			&annual, &data.RoomDetail.Description)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if annual.Valid {
		data.RoomDetail.AnnualRate = annual.Int64
	}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT a.name FROM amenities a
		JOIN room_amenities ra ON ra.amenity_id = a.id
		JOIN rooms rm ON rm.id = ra.room_id
		JOIN properties p ON p.id = rm.property_id
		WHERE p.slug = ? AND rm.slug = ? ORDER BY a.name`, r.PathValue("propertySlug"), r.PathValue("roomSlug"))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var amenity string
		if err := rows.Scan(&amenity); err != nil {
			a.serverError(w, r, err)
			return
		}
		data.RoomDetail.Amenities = append(data.RoomDetail.Amenities, amenity)
	}
	rows.Close()
	rows, err = a.db.QueryContext(r.Context(), `SELECT storage_path FROM room_photos WHERE room_id=? AND is_public=1 ORDER BY sort_order,id`, data.RoomDetail.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	for rows.Next() {
		var photo string
		if err := rows.Scan(&photo); err != nil {
			rows.Close()
			a.serverError(w, r, err)
			return
		}
		data.RoomDetail.Photos = append(data.RoomDetail.Photos, photo)
	}
	rows.Close()
	a.render(w, http.StatusOK, "public_room_detail", data)
}

func (a *App) loginForm(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if session != nil && session.User != nil {
		a.redirectForRole(w, r, session.User.Role)
		return
	}
	if session == nil {
		var err error
		session, err = a.sessions.Start(r.Context(), w, nil)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	a.render(w, http.StatusOK, "login", PageData{Title: "Masuk — Alam Segar", CSRFToken: session.CSRFToken})
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid. Muat ulang halaman dan coba lagi.", http.StatusBadRequest)
		return
	}

	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	password := r.FormValue("password")
	ipHash := a.sessions.HashIP(r)
	if email == "" || password == "" {
		a.render(w, http.StatusUnprocessableEntity, "login", PageData{
			Title: "Masuk — Alam Segar", CSRFToken: session.CSRFToken, Error: "Email dan kata sandi wajib diisi.",
		})
		return
	}

	var failed int
	err := a.db.QueryRowContext(r.Context(), `
		SELECT COUNT(*) FROM login_attempts
		WHERE email_normalized = ? AND ip_hash = ? AND was_successful = 0
		  AND attempted_at >= UTC_TIMESTAMP(6) - INTERVAL 15 MINUTE`, email, ipHash).Scan(&failed)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if failed >= 5 {
		a.render(w, http.StatusTooManyRequests, "login", PageData{
			Title: "Masuk — Alam Segar", CSRFToken: session.CSRFToken,
			Error: "Terlalu banyak percobaan. Tunggu 15 menit sebelum mencoba lagi.",
		})
		return
	}

	user := &auth.User{}
	err = a.db.QueryRowContext(r.Context(), `
		SELECT id, role, status, name, email, password_hash, email_verified_at
		FROM users WHERE email = ? AND deleted_at IS NULL`, email).
		Scan(&user.ID, &user.Role, &user.Status, &user.Name, &user.Email, &user.PasswordHash, &user.EmailVerifiedAt)
	valid := err == nil && auth.VerifyPassword(user.PasswordHash, password)
	if errors.Is(err, sql.ErrNoRows) {
		_ = auth.VerifyPassword(a.dummyPassword, password)
	} else if err != nil {
		a.serverError(w, r, err)
		return
	}

	if !valid {
		_, _ = a.db.ExecContext(r.Context(), `
			INSERT INTO login_attempts (email_normalized, ip_hash, was_successful) VALUES (?, ?, 0)`, email, ipHash)
		a.render(w, http.StatusUnauthorized, "login", PageData{
			Title: "Masuk — Alam Segar", CSRFToken: session.CSRFToken,
			Error: "Email atau kata sandi tidak cocok.",
		})
		return
	}

	if user.Status != "active" || (user.Role == "tenant" && !user.EmailVerifiedAt.Valid) {
		a.render(w, http.StatusForbidden, "login", PageData{
			Title: "Masuk — Alam Segar", CSRFToken: session.CSRFToken,
			Error: "Akun belum aktif atau belum disetujui pemilik.",
		})
		return
	}

	newSession, err := a.sessions.RotateToUser(r.Context(), w, session, user.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	_ = newSession
	_, _ = a.db.ExecContext(r.Context(), `
		UPDATE users SET last_login_at = UTC_TIMESTAMP(6) WHERE id = ?;
		DELETE FROM login_attempts WHERE email_normalized = ? AND ip_hash = ?`, user.ID, email, ipHash)
	_, _ = a.db.ExecContext(r.Context(), `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, ip_hash)
		VALUES (?, 'auth.login', 'user', ?, ?)`, user.ID, strconv.FormatInt(user.ID, 10), ipHash)
	a.redirectForRole(w, r, user.Role)
}

func (a *App) logout(w http.ResponseWriter, r *http.Request) {
	session := sessionFrom(r.Context())
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	if err := a.sessions.Destroy(r.Context(), w, session); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/masuk", http.StatusSeeOther)
}

func (a *App) ownerDashboard(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	data := PageData{Title: "Owner Dashboard — Alam Segar", User: session.User, CSRFToken: session.CSRFToken}
	metrics, err := a.ownerMetrics(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data.Metrics = metrics
	data.Bookings, err = a.recentBookings(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data.DashboardCharts, err = a.ownerDashboardCharts(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "owner_dashboard", data)
}

func (a *App) tenantDashboard(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page, err := a.loadTenantDashboard(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "tenant_dashboard", PageData{Title: "My Kos — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page})
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	status := http.StatusOK
	response := map[string]string{"status": "sehat", "database": "terhubung"}
	if err := a.db.PingContext(ctx); err != nil {
		status = http.StatusServiceUnavailable
		response = map[string]string{"status": "bermasalah", "database": "tidak terhubung"}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func (a *App) publicProperty(ctx context.Context) PublicProperty {
	property := PublicProperty{Name: "Alam Segar", Description: "Hunian nyaman dengan pengelolaan yang jelas."}
	_ = a.db.QueryRowContext(ctx, `
		SELECT name, COALESCE(description, ''), COALESCE(contact_phone, '')
		FROM properties
		WHERE is_published = 1 AND status = 'active' AND deleted_at IS NULL
		ORDER BY id LIMIT 1`).Scan(&property.Name, &property.Description, &property.Phone)
	return property
}

func (a *App) ownerMetrics(ctx context.Context, ownerID int64) (OwnerMetrics, error) {
	var result OwnerMetrics
	err := a.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT r.id) AS total_rooms,
			COUNT(DISTINCT CASE WHEN l.id IS NOT NULL THEN r.id END) AS occupied_rooms,
			COUNT(DISTINCT CASE WHEN r.condition_status = 'maintenance' THEN r.id END) AS maintenance_rooms,
			COUNT(DISTINCT CASE
				WHEN r.condition_status = 'ready' AND l.id IS NULL AND b.id IS NULL THEN r.id END) AS available_rooms
		FROM properties p
		LEFT JOIN rooms r ON r.property_id = p.id AND r.deleted_at IS NULL AND r.condition_status <> 'inactive'
		LEFT JOIN leases l ON l.room_id = r.id AND l.status = 'active'
			AND CURRENT_DATE >= l.start_date AND CURRENT_DATE < l.end_date
		LEFT JOIN booking_requests b ON b.room_id = r.id
			AND b.status IN ('approved_waiting_payment', 'proof_submitted', 'confirmed')
			AND CURRENT_DATE >= b.planned_start_date AND CURRENT_DATE < b.planned_end_date
		WHERE p.owner_id = ? AND p.deleted_at IS NULL`, ownerID).
		Scan(&result.TotalRooms, &result.OccupiedRooms, &result.Maintenance, &result.AvailableRooms)
	if err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM booking_requests b
		JOIN properties p ON p.id = b.property_id
		WHERE p.owner_id = ? AND b.status IN ('submitted', 'under_review')`, ownerID).
		Scan(&result.PendingBookings); err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM payment_submissions ps JOIN bank_accounts ba ON ba.id=ps.bank_account_id WHERE ba.owner_id=? AND ps.status='pending_verification')+
		(SELECT COUNT(*) FROM booking_payment_submissions bps JOIN booking_requests b ON b.id=bps.booking_request_id JOIN properties p ON p.id=b.property_id WHERE p.owner_id=? AND bps.status='pending_verification')`, ownerID, ownerID).
		Scan(&result.PendingPayments); err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM digital_documents WHERE owner_id=? AND status IN ('issued','awaiting_tenant')`, ownerID).Scan(&result.PendingDocuments); err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(i.total_amount-i.paid_amount),0) FROM invoices i JOIN properties p ON p.id=i.property_id WHERE p.owner_id=? AND i.status IN ('issued','partially_paid','overdue') AND i.due_date<CURRENT_DATE`, ownerID).Scan(&result.OverdueInvoices, &result.Receivables); err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM maintenance_tickets mt JOIN properties p ON p.id=mt.property_id WHERE p.owner_id=? AND mt.status IN ('open','in_progress','waiting_for_tenant')`, ownerID).Scan(&result.OpenTickets); err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM visitor_passes v JOIN properties p ON p.id=v.property_id WHERE p.owner_id=? AND v.status='pending')+
		(SELECT COUNT(*) FROM facility_reservations fr JOIN properties p ON p.id=fr.property_id WHERE p.owner_id=? AND fr.status='pending')+
		(SELECT COUNT(*) FROM resident_requests rr JOIN properties p ON p.id=rr.property_id WHERE p.owner_id=? AND rr.status IN ('submitted','reviewing')),
		(SELECT COUNT(*) FROM emergency_alerts ea JOIN properties p ON p.id=ea.property_id WHERE p.owner_id=? AND ea.status IN ('open','acknowledged')),
		(SELECT COUNT(*) FROM parcel_deliveries pd JOIN properties p ON p.id=pd.property_id WHERE p.owner_id=? AND pd.status IN ('arrived','notified'))`, ownerID, ownerID, ownerID, ownerID, ownerID).Scan(&result.SmartActions, &result.OpenEmergencies, &result.WaitingParcels); err != nil {
		return result, err
	}
	if err := a.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN je.status='posted' THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0) FROM chart_of_accounts coa JOIN journal_lines jl ON jl.account_id=coa.id JOIN journal_entries je ON je.id=jl.journal_entry_id WHERE coa.owner_id=? AND coa.id IN (SELECT ledger_account_id FROM bank_accounts WHERE owner_id=? UNION SELECT ledger_account_id FROM cash_accounts WHERE owner_id=?)`, ownerID, ownerID, ownerID).Scan(&result.Cash); err != nil {
		return result, err
	}
	return result, nil
}

func (a *App) recentBookings(ctx context.Context, ownerID int64) ([]RecentBooking, error) {
	rows, err := a.db.QueryContext(ctx, `
		SELECT b.reference_number, b.applicant_name, r.number, b.planned_start_date, b.status
		FROM booking_requests b
		JOIN properties p ON p.id = b.property_id
		JOIN rooms r ON r.id = b.room_id
		WHERE p.owner_id = ?
		ORDER BY b.created_at DESC LIMIT 5`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []RecentBooking
	for rows.Next() {
		var booking RecentBooking
		var start time.Time
		if err := rows.Scan(&booking.Reference, &booking.Name, &booking.Room, &start, &booking.Status); err != nil {
			return nil, err
		}
		booking.StartDate = start.In(a.location).Format("02 Jan 2006")
		bookings = append(bookings, booking)
	}
	return bookings, rows.Err()
}

func (a *App) ownerDashboardCharts(ctx context.Context, ownerID int64) ([]OwnerChartPoint, error) {
	now := time.Now().In(a.location)
	// Parameter untuk kolom DATE selalu UTC-midnight agar tidak bergeser ke
	// tanggal sebelumnya saat zona bisnis WITA dikonversi oleh driver.
	firstThisMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	start := firstThisMonth.AddDate(0, -5, 0)
	end := firstThisMonth.AddDate(0, 1, 0)
	rows, err := a.db.QueryContext(ctx, `
		SELECT DATE_FORMAT(je.entry_date,'%Y-%m'),
			COALESCE(SUM(CASE WHEN coa.category='revenue' THEN CAST(jl.credit AS SIGNED)-CAST(jl.debit AS SIGNED) ELSE 0 END),0),
			COALESCE(SUM(CASE WHEN coa.category='expense' THEN CAST(jl.debit AS SIGNED)-CAST(jl.credit AS SIGNED) ELSE 0 END),0)
		FROM journal_entries je
		JOIN journal_lines jl ON jl.journal_entry_id=je.id
		JOIN chart_of_accounts coa ON coa.id=jl.account_id
		WHERE je.owner_id=? AND je.status='posted' AND je.entry_date>=? AND je.entry_date<?
		  AND coa.category IN ('revenue','expense')
		GROUP BY DATE_FORMAT(je.entry_date,'%Y-%m')`, ownerID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type totals struct{ revenue, expenses int64 }
	byMonth := map[string]totals{}
	for rows.Next() {
		var key string
		var value totals
		if err := rows.Scan(&key, &value.revenue, &value.expenses); err != nil {
			return nil, err
		}
		byMonth[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	monthNames := [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}
	points := make([]OwnerChartPoint, 0, 6)
	for i := 0; i < 6; i++ {
		month := start.AddDate(0, i, 0)
		value := byMonth[month.Format("2006-01")]
		points = append(points, OwnerChartPoint{Label: monthNames[int(month.Month())-1] + " " + month.Format("06"), Revenue: value.revenue, Expenses: value.expenses})
	}
	return points, nil
}

func (a *App) render(w http.ResponseWriter, status int, name string, data PageData) {
	if secured, ok := w.(*securityResponseWriter); ok {
		data.CSPNonce = secured.nonce
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if err := a.templates.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("render %s: %v", name, err)
	}
}

func (a *App) requireRole(w http.ResponseWriter, r *http.Request, role string) *auth.Session {
	session := sessionFrom(r.Context())
	if session == nil || session.User == nil {
		http.Redirect(w, r, "/masuk", http.StatusSeeOther)
		return nil
	}
	if session.User.Role != role || session.User.Status != "active" {
		http.Error(w, "Anda tidak berhak membuka halaman ini.", http.StatusForbidden)
		return nil
	}
	return session
}

func (a *App) redirectForRole(w http.ResponseWriter, r *http.Request, role string) {
	if role == "owner" {
		http.Redirect(w, r, "/pemilik", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/penghuni", http.StatusSeeOther)
}

func (a *App) loadSession(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sessionFreePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		session, err := a.sessions.Load(r.Context(), r)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			a.serverError(w, r, err)
			return
		}
		ctx := context.WithValue(r.Context(), sessionContextKey, session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func sessionFrom(ctx context.Context) *auth.Session {
	session, _ := ctx.Value(sessionContextKey).(*auth.Session)
	return session
}

func (a *App) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nonce, err := randomKey(16)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Origin-Agent-Cluster", "?1")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; base-uri 'self'; object-src 'none'; frame-ancestors 'none'; form-action 'self'; img-src 'self' data: blob:; media-src 'self'; font-src 'self' data: https://fonts.gstatic.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; script-src 'self' 'nonce-"+nonce+"'; connect-src 'self'; worker-src 'self'; manifest-src 'self'")
		if a.cfg.SecureCookie && strings.HasPrefix(strings.ToLower(a.cfg.BaseURL), "https://") {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		if strings.HasPrefix(r.URL.Path, "/pemilik") || strings.HasPrefix(r.URL.Path, "/penghuni") || r.URL.Path == "/masuk" || r.URL.Path == "/lupa-password" || r.URL.Path == "/reset-password" {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(&securityResponseWriter{ResponseWriter: w, nonce: nonce}, r)
	})
}

func (a *App) assetHandler(prefix, root string, allowedFirst map[string]bool) http.Handler {
	server := http.StripPrefix(prefix, http.FileServer(http.Dir(root)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		relative := strings.TrimPrefix(r.URL.Path, prefix)
		clean := path.Clean("/" + relative)
		parts := strings.Split(strings.TrimPrefix(clean, "/"), "/")
		if len(parts) < 2 || !allowedFirst[parts[0]] || strings.HasPrefix(parts[len(parts)-1], ".") {
			http.NotFound(w, r)
			return
		}
		ext := strings.ToLower(filepath.Ext(parts[len(parts)-1]))
		if ext == ".php" || ext == ".html" || ext == ".md" || ext == ".json" || ext == ".lock" {
			http.NotFound(w, r)
			return
		}
		info, err := os.Stat(filepath.Join(root, filepath.FromSlash(strings.TrimPrefix(clean, "/"))))
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Cache-Control", "public, max-age=3600")
		server.ServeHTTP(w, r)
	})
}

func (a *App) pwaFile(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+name && r.URL.Path != "/pwa/"+name {
			http.NotFound(w, r)
			return
		}
		filePath := filepath.Join(a.cfg.PWAAssetsRoot, name)
		if info, err := os.Stat(filePath); err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
		if contentType == "" {
			contentType = mime.TypeByExtension(filepath.Ext(name))
		}
		w.Header().Set("Content-Type", contentType)
		if name == "service-worker.js" {
			w.Header().Set("Cache-Control", "no-cache")
			w.Header().Set("Service-Worker-Allowed", "/")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}
		http.ServeFile(w, r, filePath)
	}
}

func (a *App) serverError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	http.Error(w, "Terjadi kesalahan pada server.", http.StatusInternalServerError)
}

func formatRupiah(value int64) string {
	raw := strconv.FormatInt(value, 10)
	var groups []string
	for len(raw) > 3 {
		groups = append([]string{raw[len(raw)-3:]}, groups...)
		raw = raw[:len(raw)-3]
	}
	groups = append([]string{raw}, groups...)
	return "Rp" + strings.Join(groups, ".")
}
