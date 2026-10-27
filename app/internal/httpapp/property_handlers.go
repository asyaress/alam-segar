package httpapp

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"alam-segar/kos-management/internal/auth"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var codePattern = regexp.MustCompile(`^[A-Z0-9]+(?:-[A-Z0-9]+)*$`)

type OwnerProperty struct {
	ID          int64
	Code        string
	Name        string
	Address     string
	Status      string
	Published   bool
	RoomCount   int64
	ActiveRooms int64
}

type PropertyForm struct {
	ID          int64
	Code        string
	Slug        string
	Name        string
	Address     string
	Description string
	RulesText   string
	Phone       string
	Status      string
	Published   bool
}

type PropertyOption struct {
	ID   int64
	Name string
}

type OwnerRoom struct {
	ID             int64
	PropertyName   string
	Number         string
	TypeName       string
	MonthlyRate    int64
	Condition      string
	Published      bool
	OccupancyLabel string
}

type RoomForm struct {
	ID                        int64
	PropertyID                int64
	TypeName                  string
	Number                    string
	Slug                      string
	BlockName                 string
	FloorName                 string
	MonthlyRate               int64
	AnnualRate                int64
	ElectricityCustomerNumber string
	WaterCustomerNumber       string
	Condition                 string
	Published                 bool
	Description               string
	AmenitiesText             string
	PhotoAlt                  string
	Photos                    []RoomPhotoView
}

type RoomPhotoView struct {
	ID      int64
	Path    string
	AltText string
}

func (a *App) ownerPropertyList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	data := PageData{Title: "Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT p.id, p.code, p.name, p.address, p.status, p.is_published,
			COUNT(r.id), COALESCE(SUM(CASE WHEN r.condition_status <> 'inactive' THEN 1 ELSE 0 END), 0)
		FROM properties p
		LEFT JOIN rooms r ON r.property_id = p.id AND r.deleted_at IS NULL
		WHERE p.owner_id = ? AND p.deleted_at IS NULL
		GROUP BY p.id, p.code, p.name, p.address, p.status, p.is_published
		ORDER BY p.name`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item OwnerProperty
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Address, &item.Status, &item.Published, &item.RoomCount, &item.ActiveRooms); err != nil {
			a.serverError(w, r, err)
			return
		}
		data.Properties = append(data.Properties, item)
	}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Data properti berhasil disimpan."
	}
	a.render(w, http.StatusOK, "owner_properties", data)
}

func (a *App) ownerPropertyCreateForm(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	a.render(w, http.StatusOK, "owner_property_form", PageData{
		Title: "Tambah Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken,
		PropertyForm: PropertyForm{Status: "active"},
	})
}

func (a *App) ownerPropertyCreate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	form, ok := a.parsePropertyForm(w, r, session)
	if !ok {
		return
	}
	if message := validatePropertyForm(form); message != "" {
		a.render(w, http.StatusUnprocessableEntity, "owner_property_form", PageData{
			Title: "Tambah Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, PropertyForm: form, Error: message,
		})
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(r.Context(), `
		INSERT INTO properties
			(owner_id, code, slug, name, address, description, rules_text, contact_phone, is_published, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, session.User.ID, form.Code, form.Slug, form.Name,
		form.Address, nullableString(form.Description), nullableString(form.RulesText), nullableString(form.Phone), form.Published, form.Status)
	if err != nil {
		if isDuplicate(err) {
			a.render(w, http.StatusConflict, "owner_property_form", PageData{Title: "Tambah Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, PropertyForm: form, Error: "Kode atau slug properti sudah digunakan."})
			return
		}
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "property.create", "property", id, nil, form, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/properti?berhasil=ditambah", http.StatusSeeOther)
}

func (a *App) ownerPropertyEditForm(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	id, err := positivePathID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	form, err := a.propertyByID(r.Context(), session.User.ID, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.render(w, http.StatusOK, "owner_property_form", PageData{Title: "Ubah Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, PropertyForm: form})
}

func (a *App) ownerPropertyUpdate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	id, err := positivePathID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	before, err := a.propertyByID(r.Context(), session.User.ID, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	form, ok := a.parsePropertyForm(w, r, session)
	if !ok {
		return
	}
	form.ID = id
	if message := validatePropertyForm(form); message != "" {
		a.render(w, http.StatusUnprocessableEntity, "owner_property_form", PageData{Title: "Ubah Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, PropertyForm: form, Error: message})
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(r.Context(), `
		UPDATE properties SET code=?, slug=?, name=?, address=?, description=?, rules_text=?, contact_phone=?, is_published=?, status=?
		WHERE id=? AND owner_id=? AND deleted_at IS NULL`, form.Code, form.Slug, form.Name, form.Address,
		nullableString(form.Description), nullableString(form.RulesText), nullableString(form.Phone), form.Published, form.Status, id, session.User.ID)
	if err != nil {
		if isDuplicate(err) {
			a.render(w, http.StatusConflict, "owner_property_form", PageData{Title: "Ubah Properti — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, PropertyForm: form, Error: "Kode atau slug properti sudah digunakan."})
			return
		}
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "property.update", "property", id, before, form, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/properti?berhasil=diubah", http.StatusSeeOther)
}

func (a *App) ownerRoomList(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	data := PageData{Title: "Kamar — Alam Segar", User: session.User, CSRFToken: session.CSRFToken}
	rows, err := a.db.QueryContext(r.Context(), `
		SELECT r.id, p.name, r.number, COALESCE(rt.name, 'Tanpa tipe'), r.monthly_rate,
			r.condition_status, r.is_published,
			CASE
				WHEN l.id IS NOT NULL THEN 'Dihuni'
				WHEN b.id IS NOT NULL THEN 'Ditahan Booking'
				WHEN r.condition_status = 'maintenance' THEN 'Perawatan'
				WHEN r.condition_status = 'inactive' THEN 'Nonaktif'
				ELSE 'Tersedia'
			END AS occupancy_label
		FROM rooms r
		JOIN properties p ON p.id = r.property_id
		LEFT JOIN room_types rt ON rt.id = r.room_type_id
		LEFT JOIN leases l ON l.room_id = r.id AND l.status = 'active' AND CURRENT_DATE >= l.start_date AND CURRENT_DATE < l.end_date
		LEFT JOIN booking_requests b ON b.room_id = r.id
			AND b.status IN ('approved_waiting_payment','proof_submitted','confirmed')
			AND CURRENT_DATE >= b.planned_start_date AND CURRENT_DATE < b.planned_end_date
		WHERE p.owner_id = ? AND p.deleted_at IS NULL AND r.deleted_at IS NULL
		ORDER BY p.name, r.number`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var room OwnerRoom
		if err := rows.Scan(&room.ID, &room.PropertyName, &room.Number, &room.TypeName, &room.MonthlyRate, &room.Condition, &room.Published, &room.OccupancyLabel); err != nil {
			a.serverError(w, r, err)
			return
		}
		data.RoomsAdmin = append(data.RoomsAdmin, room)
	}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Data kamar berhasil disimpan."
	}
	a.render(w, http.StatusOK, "owner_rooms", data)
}

func (a *App) ownerRoomCreateForm(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	a.renderRoomForm(w, r, http.StatusOK, session, PageData{Title: "Tambah Kamar — Alam Segar", RoomForm: RoomForm{Condition: "ready"}})
}

func (a *App) ownerRoomCreate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	form, ok := a.parseRoomForm(w, r, session)
	if !ok {
		return
	}
	if message := validateRoomForm(form); message != "" {
		a.renderRoomForm(w, r, http.StatusUnprocessableEntity, session, PageData{Title: "Tambah Kamar — Alam Segar", RoomForm: form, Error: message})
		return
	}

	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var savedPhotos []savedUpload
	committed := false
	defer func() {
		if !committed {
			for _, photo := range savedPhotos {
				_ = os.Remove(filepath.Join(a.cfg.PublicFilesRoot, filepath.FromSlash(photo.RelativePath)))
			}
		}
	}()
	if err := assertPropertyOwner(r.Context(), tx, form.PropertyID, session.User.ID); err != nil {
		http.Error(w, "Properti tidak ditemukan.", http.StatusNotFound)
		return
	}
	typeID, err := findOrCreateRoomType(r.Context(), tx, form)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `
		INSERT INTO rooms
			(property_id, room_type_id, number, slug, block_name, floor_name, monthly_rate, annual_rate,
			 capacity, electricity_customer_number, water_customer_number, condition_status, is_published, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?)`, form.PropertyID, typeID, form.Number, form.Slug,
		nullableString(form.BlockName), nullableString(form.FloorName), form.MonthlyRate, nullablePositive(form.AnnualRate),
		nullableString(form.ElectricityCustomerNumber), nullableString(form.WaterCustomerNumber), form.Condition, form.Published, nullableString(form.Description))
	if err != nil {
		if isDuplicate(err) {
			a.renderRoomForm(w, r, http.StatusConflict, session, PageData{Title: "Tambah Kamar — Alam Segar", RoomForm: form, Error: "Nomor atau slug kamar sudah digunakan pada properti ini."})
			return
		}
		a.serverError(w, r, err)
		return
	}
	id, _ := result.LastInsertId()
	if err := syncRoomAmenitiesTx(r.Context(), tx, id, form.PropertyID, form.AmenitiesText); err != nil {
		a.serverError(w, r, err)
		return
	}
	savedPhotos, err = a.saveRoomPhotoUploads(r, tx, id, form.PhotoAlt, 0)
	if err != nil {
		a.renderRoomForm(w, r, http.StatusUnprocessableEntity, session, PageData{Title: "Tambah Kamar — Alam Segar", RoomForm: form, Error: err.Error()})
		return
	}
	_, err = tx.ExecContext(r.Context(), `
		INSERT INTO room_price_histories (room_id, monthly_rate, annual_rate, effective_from, reason, changed_by)
		VALUES (?, ?, ?, CURRENT_DATE, 'Tarif awal kamar', ?)`, id, form.MonthlyRate, nullablePositive(form.AnnualRate), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "room.create", "room", id, nil, form, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	committed = true
	http.Redirect(w, r, "/pemilik/kamar?berhasil=ditambah", http.StatusSeeOther)
}

func (a *App) ownerRoomEditForm(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	id, err := positivePathID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	form, err := a.roomByID(r.Context(), session.User.ID, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	a.renderRoomForm(w, r, http.StatusOK, session, PageData{Title: "Ubah Kamar — Alam Segar", RoomForm: form})
}

func (a *App) ownerRoomUpdate(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	id, err := positivePathID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	before, err := a.roomByID(r.Context(), session.User.ID, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	form, ok := a.parseRoomForm(w, r, session)
	if !ok {
		return
	}
	form.ID = id
	if message := validateRoomForm(form); message != "" {
		a.renderRoomForm(w, r, http.StatusUnprocessableEntity, session, PageData{Title: "Ubah Kamar — Alam Segar", RoomForm: form, Error: message})
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var savedPhotos []savedUpload
	committed := false
	defer func() {
		if !committed {
			for _, photo := range savedPhotos {
				_ = os.Remove(filepath.Join(a.cfg.PublicFilesRoot, filepath.FromSlash(photo.RelativePath)))
			}
		}
	}()
	if before.PropertyID != form.PropertyID {
		var historyCount int
		if err := tx.QueryRowContext(r.Context(), `
			SELECT
				(SELECT COUNT(*) FROM leases WHERE room_id=?)
				+
				(SELECT COUNT(*) FROM booking_requests WHERE room_id=?)`, id, id).Scan(&historyCount); err != nil {
			a.serverError(w, r, err)
			return
		}
		if historyCount > 0 {
			a.renderRoomForm(w, r, http.StatusConflict, session, PageData{Title: "Ubah Kamar — Alam Segar", RoomForm: form, Error: "Properti kamar tidak dapat dipindahkan karena sudah memiliki riwayat booking atau kontrak."})
			return
		}
	}
	if err := assertPropertyOwner(r.Context(), tx, form.PropertyID, session.User.ID); err != nil {
		http.Error(w, "Properti tidak ditemukan.", http.StatusNotFound)
		return
	}
	if form.Condition == "inactive" {
		var allocations int
		if err := tx.QueryRowContext(r.Context(), `
			SELECT
				(SELECT COUNT(*) FROM leases WHERE room_id=? AND status IN ('draft','active','expiring') AND end_date >= CURRENT_DATE)
				+
				(SELECT COUNT(*) FROM booking_requests WHERE room_id=? AND status IN ('approved_waiting_payment','proof_submitted','confirmed') AND planned_end_date > CURRENT_DATE)`, id, id).Scan(&allocations); err != nil {
			a.serverError(w, r, err)
			return
		}
		if allocations > 0 {
			a.renderRoomForm(w, r, http.StatusConflict, session, PageData{Title: "Ubah Kamar — Alam Segar", RoomForm: form, Error: "Kamar tidak dapat dinonaktifkan karena masih memiliki booking atau kontrak aktif."})
			return
		}
	}
	typeID, err := findOrCreateRoomType(r.Context(), tx, form)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	result, err := tx.ExecContext(r.Context(), `
		UPDATE rooms r JOIN properties p ON p.id=r.property_id
		SET r.property_id=?, r.room_type_id=?, r.number=?, r.slug=?, r.block_name=?, r.floor_name=?,
			r.monthly_rate=?, r.annual_rate=?, r.electricity_customer_number=?, r.water_customer_number=?,
			r.condition_status=?, r.is_published=?, r.description=?
		WHERE r.id=? AND p.owner_id=? AND r.deleted_at IS NULL`, form.PropertyID, typeID, form.Number, form.Slug,
		nullableString(form.BlockName), nullableString(form.FloorName), form.MonthlyRate, nullablePositive(form.AnnualRate),
		nullableString(form.ElectricityCustomerNumber), nullableString(form.WaterCustomerNumber), form.Condition, form.Published,
		nullableString(form.Description), id, session.User.ID)
	if err != nil {
		if isDuplicate(err) {
			a.renderRoomForm(w, r, http.StatusConflict, session, PageData{Title: "Ubah Kamar — Alam Segar", RoomForm: form, Error: "Nomor atau slug kamar sudah digunakan pada properti ini."})
			return
		}
		a.serverError(w, r, err)
		return
	}
	_, _ = result.RowsAffected()
	if before.MonthlyRate != form.MonthlyRate || before.AnnualRate != form.AnnualRate {
		_, err = tx.ExecContext(r.Context(), `
			INSERT INTO room_price_histories (room_id, monthly_rate, annual_rate, effective_from, reason, changed_by)
			VALUES (?, ?, ?, CURRENT_DATE, 'Perubahan tarif katalog kamar', ?)`, id, form.MonthlyRate, nullablePositive(form.AnnualRate), session.User.ID)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}
	if err := syncRoomAmenitiesTx(r.Context(), tx, id, form.PropertyID, form.AmenitiesText); err != nil {
		a.serverError(w, r, err)
		return
	}
	var existingPhotos int
	if err := tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM room_photos WHERE room_id=?`, id).Scan(&existingPhotos); err != nil {
		a.serverError(w, r, err)
		return
	}
	savedPhotos, err = a.saveRoomPhotoUploads(r, tx, id, form.PhotoAlt, existingPhotos)
	if err != nil {
		a.renderRoomForm(w, r, http.StatusUnprocessableEntity, session, PageData{Title: "Ubah Kamar — Alam Segar", RoomForm: form, Error: err.Error()})
		return
	}
	if err := insertAuditTx(r.Context(), tx, session.User.ID, "room.update", "room", id, before, form, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	committed = true
	http.Redirect(w, r, "/pemilik/kamar?berhasil=diubah", http.StatusSeeOther)
}

func (a *App) parsePropertyForm(w http.ResponseWriter, r *http.Request, session *auth.Session) (PropertyForm, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 128<<10)
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return PropertyForm{}, false
	}
	return PropertyForm{
		Code: strings.ToUpper(strings.TrimSpace(r.FormValue("code"))), Slug: strings.ToLower(strings.TrimSpace(r.FormValue("slug"))),
		Name: strings.TrimSpace(r.FormValue("name")), Address: strings.TrimSpace(r.FormValue("address")),
		Description: strings.TrimSpace(r.FormValue("description")), RulesText: strings.TrimSpace(r.FormValue("rules_text")),
		Phone: strings.TrimSpace(r.FormValue("phone")), Status: r.FormValue("status"), Published: r.FormValue("published") == "1",
	}, true
}

func validatePropertyForm(form PropertyForm) string {
	switch {
	case len(form.Code) < 2 || len(form.Code) > 40 || !codePattern.MatchString(form.Code):
		return "Kode properti harus 2–40 karakter berupa huruf besar, angka, atau tanda hubung."
	case len(form.Slug) < 2 || len(form.Slug) > 120 || !slugPattern.MatchString(form.Slug):
		return "Slug harus berupa huruf kecil, angka, dan tanda hubung."
	case len([]rune(form.Name)) < 3 || len([]rune(form.Name)) > 150:
		return "Nama properti harus 3–150 karakter."
	case len([]rune(form.Address)) < 5:
		return "Alamat properti wajib diisi."
	case form.Status != "active" && form.Status != "inactive":
		return "Status properti tidak valid."
	case len([]rune(form.Description)) > 5000 || len([]rune(form.RulesText)) > 10000:
		return "Deskripsi atau peraturan terlalu panjang."
	}
	return ""
}

func (a *App) propertyByID(ctx context.Context, ownerID, id int64) (PropertyForm, error) {
	var form PropertyForm
	err := a.db.QueryRowContext(ctx, `
		SELECT id, code, slug, name, address, COALESCE(description,''), COALESCE(rules_text,''), COALESCE(contact_phone,''), status, is_published
		FROM properties WHERE id=? AND owner_id=? AND deleted_at IS NULL`, id, ownerID).
		Scan(&form.ID, &form.Code, &form.Slug, &form.Name, &form.Address, &form.Description, &form.RulesText, &form.Phone, &form.Status, &form.Published)
	return form, err
}

func (a *App) parseRoomForm(w http.ResponseWriter, r *http.Request, session *auth.Session) (RoomForm, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, 30<<20)
	if err := r.ParseMultipartForm(30 << 20); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return RoomForm{}, false
	}
	propertyID, _ := strconv.ParseInt(r.FormValue("property_id"), 10, 64)
	monthlyRate, _ := strconv.ParseInt(r.FormValue("monthly_rate"), 10, 64)
	annualRate, _ := strconv.ParseInt(r.FormValue("annual_rate"), 10, 64)
	return RoomForm{
		PropertyID: propertyID, TypeName: strings.TrimSpace(r.FormValue("type_name")), Number: strings.TrimSpace(r.FormValue("number")),
		Slug: strings.ToLower(strings.TrimSpace(r.FormValue("slug"))), BlockName: strings.TrimSpace(r.FormValue("block_name")),
		FloorName: strings.TrimSpace(r.FormValue("floor_name")), MonthlyRate: monthlyRate, AnnualRate: annualRate,
		ElectricityCustomerNumber: strings.TrimSpace(r.FormValue("electricity_customer_number")), WaterCustomerNumber: strings.TrimSpace(r.FormValue("water_customer_number")),
		Condition: r.FormValue("condition"), Published: r.FormValue("published") == "1", Description: strings.TrimSpace(r.FormValue("description")),
		AmenitiesText: strings.TrimSpace(r.FormValue("amenities")), PhotoAlt: strings.TrimSpace(r.FormValue("photo_alt")),
	}, true
}

func validateRoomForm(form RoomForm) string {
	switch {
	case form.PropertyID <= 0:
		return "Properti wajib dipilih."
	case len([]rune(form.TypeName)) < 2 || len([]rune(form.TypeName)) > 120:
		return "Nama tipe kamar harus 2–120 karakter."
	case len(form.Number) < 1 || len(form.Number) > 40:
		return "Nomor kamar wajib diisi dan maksimal 40 karakter."
	case len(form.Slug) < 2 || len(form.Slug) > 140 || !slugPattern.MatchString(form.Slug):
		return "Slug kamar harus berupa huruf kecil, angka, dan tanda hubung."
	case form.MonthlyRate <= 0:
		return "Tarif bulanan harus lebih dari nol."
	case form.AnnualRate < 0:
		return "Tarif tahunan tidak boleh negatif."
	case form.Condition != "ready" && form.Condition != "maintenance" && form.Condition != "inactive":
		return "Kondisi kamar tidak valid."
	case len([]rune(form.Description)) > 5000:
		return "Deskripsi kamar terlalu panjang."
	case len([]rune(form.AmenitiesText)) > 2000 || len([]rune(form.PhotoAlt)) > 255:
		return "Daftar fasilitas atau teks alternatif foto terlalu panjang."
	}
	return ""
}

func (a *App) roomByID(ctx context.Context, ownerID, id int64) (RoomForm, error) {
	var form RoomForm
	var annual sql.NullInt64
	err := a.db.QueryRowContext(ctx, `
		SELECT r.id, r.property_id, COALESCE(rt.name,''), r.number, r.slug, COALESCE(r.block_name,''), COALESCE(r.floor_name,''),
			r.monthly_rate, r.annual_rate, COALESCE(r.electricity_customer_number,''), COALESCE(r.water_customer_number,''),
			r.condition_status, r.is_published, COALESCE(r.description,'')
		FROM rooms r JOIN properties p ON p.id=r.property_id LEFT JOIN room_types rt ON rt.id=r.room_type_id
		WHERE r.id=? AND p.owner_id=? AND r.deleted_at IS NULL`, id, ownerID).
		Scan(&form.ID, &form.PropertyID, &form.TypeName, &form.Number, &form.Slug, &form.BlockName, &form.FloorName,
			&form.MonthlyRate, &annual, &form.ElectricityCustomerNumber, &form.WaterCustomerNumber, &form.Condition, &form.Published, &form.Description)
	if annual.Valid {
		form.AnnualRate = annual.Int64
	}
	if err == nil {
		var amenities sql.NullString
		_ = a.db.QueryRowContext(ctx, `SELECT GROUP_CONCAT(a.name ORDER BY a.name SEPARATOR ', ') FROM room_amenities ra JOIN amenities a ON a.id=ra.amenity_id WHERE ra.room_id=?`, id).Scan(&amenities)
		if amenities.Valid {
			form.AmenitiesText = amenities.String
		}
		rows, photoErr := a.db.QueryContext(ctx, `SELECT id,storage_path,alt_text FROM room_photos WHERE room_id=? ORDER BY sort_order,id`, id)
		if photoErr != nil {
			return form, photoErr
		}
		for rows.Next() {
			var photo RoomPhotoView
			if scanErr := rows.Scan(&photo.ID, &photo.Path, &photo.AltText); scanErr != nil {
				rows.Close()
				return form, scanErr
			}
			form.Photos = append(form.Photos, photo)
		}
		rows.Close()
	}
	return form, err
}

func syncRoomAmenitiesTx(ctx context.Context, tx *sql.Tx, roomID, propertyID int64, raw string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM room_amenities WHERE room_id=?`, roomID); err != nil {
		return err
	}
	seen := map[string]bool{}
	count := 0
	for _, item := range strings.Split(raw, ",") {
		name := strings.TrimSpace(item)
		key := strings.ToLower(name)
		if name == "" || seen[key] {
			continue
		}
		if len([]rune(name)) > 120 {
			return errors.New("nama fasilitas maksimal 120 karakter")
		}
		count++
		if count > 20 {
			return errors.New("maksimal 20 fasilitas per kamar")
		}
		seen[key] = true
		if _, err := tx.ExecContext(ctx, `INSERT INTO amenities (property_id,name) VALUES (?,?) ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`, propertyID, name); err != nil {
			return err
		}
		var amenityID int64
		if err := tx.QueryRowContext(ctx, `SELECT id FROM amenities WHERE property_id=? AND name=?`, propertyID, name).Scan(&amenityID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO room_amenities (room_id,amenity_id) VALUES (?,?)`, roomID, amenityID); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) saveRoomPhotoUploads(r *http.Request, tx *sql.Tx, roomID int64, altText string, existing int) ([]savedUpload, error) {
	if r.MultipartForm == nil {
		return nil, nil
	}
	headers := r.MultipartForm.File["photos"]
	if existing+len(headers) > 10 {
		return nil, errors.New("maksimal 10 foto per kamar")
	}
	if altText == "" {
		altText = "Foto kamar"
	}
	var saved []savedUpload
	for index, header := range headers {
		file, err := header.Open()
		if err != nil {
			return saved, err
		}
		upload, saveErr := a.persistPublicImage(file, header, "rooms", 5<<20)
		_ = file.Close()
		if saveErr != nil {
			return saved, saveErr
		}
		saved = append(saved, upload)
		_, err = tx.ExecContext(r.Context(), `INSERT INTO room_photos (room_id,storage_path,alt_text,sort_order,is_public) VALUES (?,?,?,?,1)`, roomID, upload.RelativePath, altText, existing+index)
		if err != nil {
			return saved, err
		}
	}
	return saved, nil
}

func (a *App) ownerRoomDeletePhoto(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(session, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	roomID, e1 := parseID(r.PathValue("id"))
	photoID, e2 := parseID(r.PathValue("photoID"))
	if e1 != nil || e2 != nil {
		http.NotFound(w, r)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var storagePath string
	err = tx.QueryRowContext(r.Context(), `SELECT rp.storage_path FROM room_photos rp JOIN rooms rm ON rm.id=rp.room_id JOIN properties p ON p.id=rm.property_id WHERE rp.id=? AND rp.room_id=? AND p.owner_id=? FOR UPDATE`, photoID, roomID, session.User.ID).Scan(&storagePath)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `DELETE FROM room_photos WHERE id=?`, photoID); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err := tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	_ = os.Remove(filepath.Join(a.cfg.PublicFilesRoot, filepath.FromSlash(storagePath)))
	http.Redirect(w, r, "/pemilik/kamar/"+strconv.FormatInt(roomID, 10)+"/ubah", http.StatusSeeOther)
}

func (a *App) renderRoomForm(w http.ResponseWriter, r *http.Request, status int, session *auth.Session, data PageData) {
	data.User = session.User
	data.CSRFToken = session.CSRFToken
	rows, err := a.db.QueryContext(r.Context(), `SELECT id, name FROM properties WHERE owner_id=? AND status='active' AND deleted_at IS NULL ORDER BY name`, session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var option PropertyOption
		if err := rows.Scan(&option.ID, &option.Name); err != nil {
			a.serverError(w, r, err)
			return
		}
		data.PropertyOptions = append(data.PropertyOptions, option)
	}
	a.render(w, status, "owner_room_form", data)
}

func assertPropertyOwner(ctx context.Context, tx *sql.Tx, propertyID, ownerID int64) error {
	var id int64
	return tx.QueryRowContext(ctx, `SELECT id FROM properties WHERE id=? AND owner_id=? AND deleted_at IS NULL FOR UPDATE`, propertyID, ownerID).Scan(&id)
}

func findOrCreateRoomType(ctx context.Context, tx *sql.Tx, form RoomForm) (int64, error) {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO room_types (property_id, name, base_monthly_rate, base_annual_rate)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id)`, form.PropertyID, form.TypeName, form.MonthlyRate, nullablePositive(form.AnnualRate))
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRowContext(ctx, `SELECT id FROM room_types WHERE property_id=? AND name=?`, form.PropertyID, form.TypeName).Scan(&id)
	return id, err
}

func positivePathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("id tidak valid")
	}
	return id, nil
}

func nullablePositive(value int64) any {
	if value <= 0 {
		return nil
	}
	return value
}

func isDuplicate(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "duplicate entry")
}

func insertAuditTx(ctx context.Context, tx *sql.Tx, actorID any, action, entityType string, entityID int64, before, after any, ipHash string) error {
	beforeJSON, _ := json.Marshal(before)
	afterJSON, _ := json.Marshal(after)
	_, err := tx.ExecContext(ctx, `
		INSERT INTO audit_logs (actor_user_id, action, entity_type, entity_id, before_json, after_json, ip_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, actorID, action, entityType, strconv.FormatInt(entityID, 10), nullableJSON(beforeJSON), nullableJSON(afterJSON), ipHash)
	return err
}

func nullableJSON(value []byte) any {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return string(value)
}
