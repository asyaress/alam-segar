package httpapp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type SmartMetrics struct {
	PendingVisitors, PendingReservations, WaitingParcels, PendingRequests, OpenEmergencies int64
}

type SmartFacilityView struct {
	ID                                    int64
	Property, Name, Location, Description string
	OpensAt, ClosesAt, Status             string
	SlotMinutes, Capacity                 int
	RequiresApproval                      bool
}

type SmartReservationView struct {
	ID                                                                  int64
	Number, Facility, Tenant, Property, Room, StartsAt, EndsAt, Purpose string
	Status, OwnerNote                                                   string
	Date, CreatedAt                                                     time.Time
}

type VisitorPassView struct {
	ID                                                                                   int64
	Number, Tenant, Property, Room, VisitorName, VisitorPhone, StartsAt, EndsAt, Purpose string
	VehiclePlate, AccessCode, Status, OwnerNote                                          string
	VisitDate, CreatedAt                                                                 time.Time
}

type ParcelView struct {
	ID                                                                    int64
	Number, Tenant, Property, Room, Courier, TrackingLast4, Notes, Status string
	ReceivedAt                                                            time.Time
}

type ResidentRequestView struct {
	ID                                                                 int64
	Number, Tenant, Property, Room, Type, Notes, Status, OwnerResponse string
	RequestedDate                                                      time.Time
	HasRequestedDate                                                   bool
	CreatedAt                                                          time.Time
}

type EmergencyAlertView struct {
	ID                                                                      int64
	Number, Tenant, Property, Room, Category, Description, Status, Response string
	CreatedAt                                                               time.Time
}

type SmartTenantOption struct {
	LeaseID, TenantID    int64
	Name, Property, Room string
}

type OwnerSmartPage struct {
	Metrics       SmartMetrics
	Facilities    []SmartFacilityView
	Reservations  []SmartReservationView
	Visitors      []VisitorPassView
	Parcels       []ParcelView
	Requests      []ResidentRequestView
	Emergencies   []EmergencyAlertView
	Properties    []LeaseOption
	ActiveTenants []SmartTenantOption
}

type TenantSmartPage struct {
	HasLease       bool
	LeaseID        int64
	Property, Room string
	Facilities     []SmartFacilityView
	Reservations   []SmartReservationView
	Visitors       []VisitorPassView
	Parcels        []ParcelView
	Requests       []ResidentRequestView
	Emergencies    []EmergencyAlertView
}

type activeLeaseContext struct {
	LeaseID, PropertyID, RoomID int64
	Property, Room              string
}

func (a *App) ownerSmartLiving(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "owner")
	if session == nil {
		return
	}
	page, err := a.loadOwnerSmartPage(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Smart Kos Hub — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Perubahan layanan smart kos berhasil disimpan."
	}
	a.render(w, http.StatusOK, "owner_smart", data)
}

func (a *App) loadOwnerSmartPage(ctx context.Context, ownerID int64) (OwnerSmartPage, error) {
	var page OwnerSmartPage
	if err := a.db.QueryRowContext(ctx, `SELECT
		(SELECT COUNT(*) FROM visitor_passes v JOIN properties p ON p.id=v.property_id WHERE p.owner_id=? AND v.status='pending'),
		(SELECT COUNT(*) FROM facility_reservations fr JOIN properties p ON p.id=fr.property_id WHERE p.owner_id=? AND fr.status='pending'),
		(SELECT COUNT(*) FROM parcel_deliveries pd JOIN properties p ON p.id=pd.property_id WHERE p.owner_id=? AND pd.status IN ('arrived','notified')),
		(SELECT COUNT(*) FROM resident_requests rr JOIN properties p ON p.id=rr.property_id WHERE p.owner_id=? AND rr.status IN ('submitted','reviewing')),
		(SELECT COUNT(*) FROM emergency_alerts ea JOIN properties p ON p.id=ea.property_id WHERE p.owner_id=? AND ea.status IN ('open','acknowledged'))`, ownerID, ownerID, ownerID, ownerID, ownerID).Scan(&page.Metrics.PendingVisitors, &page.Metrics.PendingReservations, &page.Metrics.WaitingParcels, &page.Metrics.PendingRequests, &page.Metrics.OpenEmergencies); err != nil {
		return page, err
	}
	rows, err := a.db.QueryContext(ctx, `SELECT sf.id,p.name,sf.name,sf.location,COALESCE(sf.description,''),DATE_FORMAT(sf.opens_at,'%H:%i'),DATE_FORMAT(sf.closes_at,'%H:%i'),sf.slot_minutes,sf.capacity,sf.requires_approval,sf.status FROM shared_facilities sf JOIN properties p ON p.id=sf.property_id WHERE sf.owner_id=? ORDER BY p.name,sf.name`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v SmartFacilityView
		if err := rows.Scan(&v.ID, &v.Property, &v.Name, &v.Location, &v.Description, &v.OpensAt, &v.ClosesAt, &v.SlotMinutes, &v.Capacity, &v.RequiresApproval, &v.Status); err != nil {
			rows.Close()
			return page, err
		}
		page.Facilities = append(page.Facilities, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT fr.id,fr.reservation_number,sf.name,u.name,p.name,rm.number,fr.reservation_date,DATE_FORMAT(fr.starts_at,'%H:%i'),DATE_FORMAT(fr.ends_at,'%H:%i'),COALESCE(fr.purpose,''),fr.status,COALESCE(fr.owner_note,''),fr.created_at FROM facility_reservations fr JOIN shared_facilities sf ON sf.id=fr.facility_id JOIN users u ON u.id=fr.tenant_id JOIN properties p ON p.id=fr.property_id JOIN rooms rm ON rm.id=fr.room_id WHERE p.owner_id=? ORDER BY FIELD(fr.status,'pending','approved','completed','rejected','cancelled'),fr.reservation_date,fr.starts_at LIMIT 200`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v SmartReservationView
		if err := rows.Scan(&v.ID, &v.Number, &v.Facility, &v.Tenant, &v.Property, &v.Room, &v.Date, &v.StartsAt, &v.EndsAt, &v.Purpose, &v.Status, &v.OwnerNote, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Reservations = append(page.Reservations, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT v.id,v.pass_number,u.name,p.name,rm.number,v.visitor_name,COALESCE(v.visitor_phone,''),v.visit_date,DATE_FORMAT(v.starts_at,'%H:%i'),DATE_FORMAT(v.ends_at,'%H:%i'),v.purpose,COALESCE(v.vehicle_plate,''),COALESCE(v.access_code,''),v.status,COALESCE(v.owner_note,''),v.created_at FROM visitor_passes v JOIN users u ON u.id=v.tenant_id JOIN properties p ON p.id=v.property_id JOIN rooms rm ON rm.id=v.room_id WHERE p.owner_id=? ORDER BY FIELD(v.status,'pending','approved','checked_in','checked_out','rejected','cancelled','expired'),v.visit_date,v.starts_at LIMIT 200`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v VisitorPassView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.VisitorName, &v.VisitorPhone, &v.VisitDate, &v.StartsAt, &v.EndsAt, &v.Purpose, &v.VehiclePlate, &v.AccessCode, &v.Status, &v.OwnerNote, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Visitors = append(page.Visitors, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT pd.id,pd.parcel_number,u.name,p.name,rm.number,pd.courier,COALESCE(pd.tracking_last4,''),COALESCE(pd.notes,''),pd.status,pd.received_at FROM parcel_deliveries pd JOIN users u ON u.id=pd.tenant_id JOIN properties p ON p.id=pd.property_id JOIN rooms rm ON rm.id=pd.room_id WHERE p.owner_id=? ORDER BY FIELD(pd.status,'arrived','notified','collected','returned'),pd.received_at DESC LIMIT 200`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v ParcelView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Courier, &v.TrackingLast4, &v.Notes, &v.Status, &v.ReceivedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Parcels = append(page.Parcels, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT rr.id,rr.request_number,u.name,p.name,rm.number,rr.request_type,rr.requested_date,rr.notes,rr.status,COALESCE(rr.owner_response,''),rr.created_at FROM resident_requests rr JOIN users u ON u.id=rr.tenant_id JOIN properties p ON p.id=rr.property_id JOIN rooms rm ON rm.id=rr.room_id WHERE p.owner_id=? ORDER BY FIELD(rr.status,'submitted','reviewing','approved','completed','rejected','cancelled'),rr.created_at DESC LIMIT 200`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v ResidentRequestView
		var d sql.NullTime
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Type, &d, &v.Notes, &v.Status, &v.OwnerResponse, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		if d.Valid {
			v.RequestedDate = d.Time
			v.HasRequestedDate = true
		}
		page.Requests = append(page.Requests, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT ea.id,ea.alert_number,u.name,p.name,rm.number,ea.category,ea.description,ea.status,COALESCE(ea.owner_response,''),ea.created_at FROM emergency_alerts ea JOIN users u ON u.id=ea.tenant_id JOIN properties p ON p.id=ea.property_id JOIN rooms rm ON rm.id=ea.room_id WHERE p.owner_id=? ORDER BY FIELD(ea.status,'open','acknowledged','resolved','cancelled'),ea.created_at DESC LIMIT 200`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v EmergencyAlertView
		if err := rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Category, &v.Description, &v.Status, &v.Response, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Emergencies = append(page.Emergencies, v)
	}
	rows.Close()
	page.Properties, _, err = a.ownerPropertyRoomOptions(ctx, ownerID)
	if err != nil {
		return page, err
	}
	rows, err = a.db.QueryContext(ctx, `SELECT l.id,u.id,u.name,p.name,rm.number FROM leases l JOIN users u ON u.id=l.tenant_id JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE p.owner_id=? AND l.status IN ('active','expiring') AND CURRENT_DATE>=l.start_date AND CURRENT_DATE<l.end_date ORDER BY p.name,rm.number,u.name`, ownerID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v SmartTenantOption
		if err := rows.Scan(&v.LeaseID, &v.TenantID, &v.Name, &v.Property, &v.Room); err != nil {
			rows.Close()
			return page, err
		}
		page.ActiveTenants = append(page.ActiveTenants, v)
	}
	rows.Close()
	return page, nil
}

func (a *App) tenantSmartLiving(w http.ResponseWriter, r *http.Request) {
	session := a.requireRole(w, r, "tenant")
	if session == nil {
		return
	}
	page, err := a.loadTenantSmartPage(r.Context(), session.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	data := PageData{Title: "Smart Kos — Alam Segar", User: session.User, CSRFToken: session.CSRFToken, Payload: page}
	if r.URL.Query().Get("berhasil") != "" {
		data.Success = "Permintaan berhasil dikirim dan tercatat."
	}
	a.render(w, http.StatusOK, "tenant_smart", data)
}

func (a *App) loadTenantSmartPage(ctx context.Context, tenantID int64) (TenantSmartPage, error) {
	var page TenantSmartPage
	lease, err := a.activeLeaseForTenant(ctx, tenantID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return page, err
	}
	if err == nil {
		page.HasLease = true
		page.LeaseID = lease.LeaseID
		page.Property = lease.Property
		page.Room = lease.Room
		rows, qerr := a.db.QueryContext(ctx, `SELECT sf.id,p.name,sf.name,sf.location,COALESCE(sf.description,''),DATE_FORMAT(sf.opens_at,'%H:%i'),DATE_FORMAT(sf.closes_at,'%H:%i'),sf.slot_minutes,sf.capacity,sf.requires_approval,sf.status FROM shared_facilities sf JOIN properties p ON p.id=sf.property_id WHERE sf.property_id=? AND sf.status='active' ORDER BY sf.name`, lease.PropertyID)
		if qerr != nil {
			return page, qerr
		}
		for rows.Next() {
			var v SmartFacilityView
			if qerr = rows.Scan(&v.ID, &v.Property, &v.Name, &v.Location, &v.Description, &v.OpensAt, &v.ClosesAt, &v.SlotMinutes, &v.Capacity, &v.RequiresApproval, &v.Status); qerr != nil {
				rows.Close()
				return page, qerr
			}
			page.Facilities = append(page.Facilities, v)
		}
		rows.Close()
	}
	rows, err := a.db.QueryContext(ctx, `SELECT fr.id,fr.reservation_number,sf.name,'',p.name,rm.number,fr.reservation_date,DATE_FORMAT(fr.starts_at,'%H:%i'),DATE_FORMAT(fr.ends_at,'%H:%i'),COALESCE(fr.purpose,''),fr.status,COALESCE(fr.owner_note,''),fr.created_at FROM facility_reservations fr JOIN shared_facilities sf ON sf.id=fr.facility_id JOIN properties p ON p.id=fr.property_id JOIN rooms rm ON rm.id=fr.room_id WHERE fr.tenant_id=? ORDER BY fr.created_at DESC LIMIT 100`, tenantID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v SmartReservationView
		if err = rows.Scan(&v.ID, &v.Number, &v.Facility, &v.Tenant, &v.Property, &v.Room, &v.Date, &v.StartsAt, &v.EndsAt, &v.Purpose, &v.Status, &v.OwnerNote, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Reservations = append(page.Reservations, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT v.id,v.pass_number,'',p.name,rm.number,v.visitor_name,COALESCE(v.visitor_phone,''),v.visit_date,DATE_FORMAT(v.starts_at,'%H:%i'),DATE_FORMAT(v.ends_at,'%H:%i'),v.purpose,COALESCE(v.vehicle_plate,''),COALESCE(v.access_code,''),v.status,COALESCE(v.owner_note,''),v.created_at FROM visitor_passes v JOIN properties p ON p.id=v.property_id JOIN rooms rm ON rm.id=v.room_id WHERE v.tenant_id=? ORDER BY v.created_at DESC LIMIT 100`, tenantID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v VisitorPassView
		if err = rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.VisitorName, &v.VisitorPhone, &v.VisitDate, &v.StartsAt, &v.EndsAt, &v.Purpose, &v.VehiclePlate, &v.AccessCode, &v.Status, &v.OwnerNote, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Visitors = append(page.Visitors, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT pd.id,pd.parcel_number,'',p.name,rm.number,pd.courier,COALESCE(pd.tracking_last4,''),COALESCE(pd.notes,''),pd.status,pd.received_at FROM parcel_deliveries pd JOIN properties p ON p.id=pd.property_id JOIN rooms rm ON rm.id=pd.room_id WHERE pd.tenant_id=? ORDER BY pd.received_at DESC LIMIT 100`, tenantID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v ParcelView
		if err = rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Courier, &v.TrackingLast4, &v.Notes, &v.Status, &v.ReceivedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Parcels = append(page.Parcels, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT rr.id,rr.request_number,'',p.name,rm.number,rr.request_type,rr.requested_date,rr.notes,rr.status,COALESCE(rr.owner_response,''),rr.created_at FROM resident_requests rr JOIN properties p ON p.id=rr.property_id JOIN rooms rm ON rm.id=rr.room_id WHERE rr.tenant_id=? ORDER BY rr.created_at DESC LIMIT 100`, tenantID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v ResidentRequestView
		var d sql.NullTime
		if err = rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Type, &d, &v.Notes, &v.Status, &v.OwnerResponse, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		if d.Valid {
			v.RequestedDate = d.Time
			v.HasRequestedDate = true
		}
		page.Requests = append(page.Requests, v)
	}
	rows.Close()
	rows, err = a.db.QueryContext(ctx, `SELECT ea.id,ea.alert_number,'',p.name,rm.number,ea.category,ea.description,ea.status,COALESCE(ea.owner_response,''),ea.created_at FROM emergency_alerts ea JOIN properties p ON p.id=ea.property_id JOIN rooms rm ON rm.id=ea.room_id WHERE ea.tenant_id=? ORDER BY ea.created_at DESC LIMIT 100`, tenantID)
	if err != nil {
		return page, err
	}
	for rows.Next() {
		var v EmergencyAlertView
		if err = rows.Scan(&v.ID, &v.Number, &v.Tenant, &v.Property, &v.Room, &v.Category, &v.Description, &v.Status, &v.Response, &v.CreatedAt); err != nil {
			rows.Close()
			return page, err
		}
		page.Emergencies = append(page.Emergencies, v)
	}
	rows.Close()
	return page, nil
}

func (a *App) activeLeaseForTenant(ctx context.Context, tenantID int64) (activeLeaseContext, error) {
	var v activeLeaseContext
	today := businessToday(a.location)
	err := a.db.QueryRowContext(ctx, `SELECT l.id,l.property_id,l.room_id,p.name,rm.number FROM leases l JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE l.tenant_id=? AND l.status IN ('active','expiring') AND ?>=l.start_date AND ?<l.end_date ORDER BY l.start_date DESC LIMIT 1`, tenantID, today, today).Scan(&v.LeaseID, &v.PropertyID, &v.RoomID, &v.Property, &v.Room)
	return v, err
}

func activeLeaseForTenantTx(ctx context.Context, tx *sql.Tx, tenantID int64, today time.Time) (activeLeaseContext, error) {
	var v activeLeaseContext
	err := tx.QueryRowContext(ctx, `SELECT l.id,l.property_id,l.room_id,p.name,rm.number FROM leases l JOIN properties p ON p.id=l.property_id JOIN rooms rm ON rm.id=l.room_id WHERE l.tenant_id=? AND l.status IN ('active','expiring') AND ?>=l.start_date AND ?<l.end_date ORDER BY l.start_date DESC LIMIT 1 FOR UPDATE`, tenantID, today, today).Scan(&v.LeaseID, &v.PropertyID, &v.RoomID, &v.Property, &v.Room)
	return v, err
}

func (a *App) ownerCreateFacility(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	propertyID, e := parseID(r.FormValue("property_id"))
	name := strings.TrimSpace(r.FormValue("name"))
	location := strings.TrimSpace(r.FormValue("location"))
	opens, oe := clockMinutes(r.FormValue("opens_at"))
	closes, ce := clockMinutes(r.FormValue("closes_at"))
	slot, _ := strconv.Atoi(r.FormValue("slot_minutes"))
	capacity, _ := strconv.Atoi(r.FormValue("capacity"))
	if e != nil || len(name) < 2 || len(name) > 150 || len(location) < 2 || len(location) > 190 || oe != nil || ce != nil || opens >= closes || slot < 15 || slot > 240 || capacity < 1 || capacity > 100 {
		http.Error(w, "Data fasilitas tidak valid. Jam tutup harus setelah jam buka.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), nil)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var valid int
	if err = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM properties WHERE id=? AND owner_id=? AND deleted_at IS NULL`, propertyID, s.User.ID).Scan(&valid); err != nil || valid != 1 {
		http.Error(w, "Properti tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	res, err := tx.ExecContext(r.Context(), `INSERT INTO shared_facilities(owner_id,property_id,name,location,description,opens_at,closes_at,slot_minutes,capacity,requires_approval) VALUES(?,?,?,?,?,?,?,?,?,?)`, s.User.ID, propertyID, name, location, nullableString(strings.TrimSpace(r.FormValue("description"))), r.FormValue("opens_at"), r.FormValue("closes_at"), slot, capacity, r.FormValue("requires_approval") == "1")
	if err != nil {
		if isDuplicate(err) {
			http.Error(w, "Nama fasilitas sudah digunakan pada properti ini.", http.StatusConflict)
			return
		}
		a.serverError(w, r, err)
		return
	}
	id, _ := res.LastInsertId()
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.facility.create", "shared_facility", id, nil, map[string]any{"name": name, "property_id": propertyID}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=fasilitas", http.StatusSeeOther)
}

func (a *App) tenantCreateFacilityReservation(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	facilityID, e := parseID(r.FormValue("facility_id"))
	date, de := parseBusinessDate(r.FormValue("reservation_date"))
	start, se := clockMinutes(r.FormValue("starts_at"))
	end, ee := clockMinutes(r.FormValue("ends_at"))
	purpose := strings.TrimSpace(r.FormValue("purpose"))
	today := businessToday(a.location)
	if e != nil || de != nil || se != nil || ee != nil || date.Before(today) || date.After(today.AddDate(0, 2, 0)) || start >= end || end-start > 240 || len(purpose) > 500 {
		http.Error(w, "Jadwal reservasi tidak valid. Reservasi maksimal 60 hari ke depan dan 4 jam.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	lease, err := activeLeaseForTenantTx(r.Context(), tx, s.User.ID, today)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Reservasi hanya tersedia saat kontrak aktif.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var opens, closes, slot, capacity int
	var requires bool
	err = tx.QueryRowContext(r.Context(), `SELECT TIME_TO_SEC(opens_at) DIV 60,TIME_TO_SEC(closes_at) DIV 60,slot_minutes,capacity,requires_approval FROM shared_facilities WHERE id=? AND property_id=? AND status='active' FOR UPDATE`, facilityID, lease.PropertyID).Scan(&opens, &closes, &slot, &capacity, &requires)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Fasilitas tidak tersedia untuk hunian Anda.", http.StatusNotFound)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if start < opens || end > closes || (end-start)%slot != 0 {
		http.Error(w, "Waktu harus berada dalam jam operasional dan mengikuti durasi slot fasilitas.", http.StatusUnprocessableEntity)
		return
	}
	status := "pending"
	if !requires {
		var conflicts int
		err = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM facility_reservations WHERE facility_id=? AND reservation_date=? AND status='approved' AND starts_at<? AND ends_at>?`, facilityID, date, r.FormValue("ends_at"), r.FormValue("starts_at")).Scan(&conflicts)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
		if conflicts >= capacity {
			http.Error(w, "Slot fasilitas sudah penuh. Pilih waktu lain.", http.StatusConflict)
			return
		}
		status = "approved"
	}
	number, _ := referenceNumber("RSV")
	res, err := tx.ExecContext(r.Context(), `INSERT INTO facility_reservations(reservation_number,facility_id,tenant_id,lease_id,property_id,room_id,reservation_date,starts_at,ends_at,purpose,status) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, number, facilityID, s.User.ID, lease.LeaseID, lease.PropertyID, lease.RoomID, date, r.FormValue("starts_at"), r.FormValue("ends_at"), nullableString(purpose), status)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := res.LastInsertId()
	if err = notifyOwnerForPropertyTx(r.Context(), tx, lease.PropertyID, fmt.Sprintf("smart.reservation.created.%d", id), "Reservasi fasilitas baru", s.User.Name+" mengajukan reservasi fasilitas.", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if status == "approved" {
		_ = notifyUserTx(r.Context(), tx, s.User.ID, fmt.Sprintf("smart.reservation.auto.%d", id), "Reservasi disetujui", "Reservasi fasilitas otomatis disetujui.", "/penghuni/layanan")
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.reservation.create", "facility_reservation", id, nil, map[string]any{"status": status, "date": r.FormValue("reservation_date")}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=reservasi", http.StatusSeeOther)
}

func (a *App) ownerUpdateFacilityReservation(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	next := r.FormValue("status")
	note := strings.TrimSpace(r.FormValue("note"))
	if e != nil || len(note) < 3 || !map[string]bool{"approved": true, "rejected": true, "completed": true, "cancelled": true}[next] {
		http.Error(w, "Status atau catatan tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID, facilityID, capacity int64
	var old string
	var date time.Time
	var start, end string
	err = tx.QueryRowContext(r.Context(), `SELECT fr.tenant_id,fr.facility_id,fr.status,fr.reservation_date,DATE_FORMAT(fr.starts_at,'%H:%i'),DATE_FORMAT(fr.ends_at,'%H:%i'),sf.capacity FROM facility_reservations fr JOIN shared_facilities sf ON sf.id=fr.facility_id JOIN properties p ON p.id=fr.property_id WHERE fr.id=? AND p.owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &facilityID, &old, &date, &start, &end, &capacity)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !reservationTransition(old, next) {
		http.Error(w, "Perubahan status reservasi tidak diizinkan.", http.StatusConflict)
		return
	}
	if next == "completed" {
		today := businessToday(a.location)
		endMinutes, _ := clockMinutes(end)
		if date.After(today) || (sameBusinessDate(date, today) && businessClockMinutes(a.location) < endMinutes) {
			http.Error(w, "Reservasi belum melewati waktu selesai.", http.StatusConflict)
			return
		}
	}
	if next == "approved" {
		var conflicts int64
		if err = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM facility_reservations WHERE facility_id=? AND id<>? AND reservation_date=? AND status='approved' AND starts_at<? AND ends_at>?`, facilityID, id, date, end, start).Scan(&conflicts); err != nil {
			a.serverError(w, r, err)
			return
		}
		if conflicts >= capacity {
			http.Error(w, "Slot sudah penuh oleh reservasi lain.", http.StatusConflict)
			return
		}
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE facility_reservations SET status=?,owner_note=?,reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6) WHERE id=?`, next, note, s.User.ID, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("smart.reservation.status.%d.%s", id, next), "Status reservasi fasilitas", note, "/penghuni/layanan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.reservation.status", "facility_reservation", id, map[string]any{"status": old}, map[string]any{"status": next, "note": note}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=reservasi", http.StatusSeeOther)
}

func (a *App) tenantCancelFacilityReservation(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var old string
	var date time.Time
	err = tx.QueryRowContext(r.Context(), `SELECT property_id,status,reservation_date FROM facility_reservations WHERE id=? AND tenant_id=? FOR UPDATE`, id, s.User.ID).Scan(&propertyID, &old, &date)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !map[string]bool{"pending": true, "approved": true}[old] || date.Before(businessToday(a.location)) {
		http.Error(w, "Reservasi tidak dapat dibatalkan.", http.StatusConflict)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE facility_reservations SET status='cancelled' WHERE id=?`, id); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyOwnerForPropertyTx(r.Context(), tx, propertyID, fmt.Sprintf("smart.reservation.cancelled.%d", id), "Reservasi dibatalkan penghuni", s.User.Name+" membatalkan reservasi fasilitas.", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.reservation.cancel", "facility_reservation", id, map[string]any{"status": old}, map[string]any{"status": "cancelled"}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=batal-reservasi", http.StatusSeeOther)
}

func (a *App) tenantCreateVisitorPass(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("visitor_name"))
	phone := strings.TrimSpace(r.FormValue("visitor_phone"))
	purpose := strings.TrimSpace(r.FormValue("purpose"))
	plate := strings.ToUpper(strings.TrimSpace(r.FormValue("vehicle_plate")))
	date, de := parseBusinessDate(r.FormValue("visit_date"))
	start, se := clockMinutes(r.FormValue("starts_at"))
	end, ee := clockMinutes(r.FormValue("ends_at"))
	today := businessToday(a.location)
	if len(name) < 3 || len(name) > 150 || len(phone) > 40 || len(purpose) < 3 || len(purpose) > 500 || len(plate) > 30 || de != nil || se != nil || ee != nil || date.Before(today) || date.After(today.AddDate(0, 2, 0)) || start >= end || end-start > 480 {
		http.Error(w, "Data kunjungan tidak valid. Kunjungan maksimal 60 hari ke depan dan 8 jam.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	lease, err := activeLeaseForTenantTx(r.Context(), tx, s.User.ID, today)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Izin tamu hanya dapat dibuat saat kontrak aktif.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, _ := referenceNumber("TAMU")
	res, err := tx.ExecContext(r.Context(), `INSERT INTO visitor_passes(pass_number,tenant_id,lease_id,property_id,room_id,visitor_name,visitor_phone,visit_date,starts_at,ends_at,purpose,vehicle_plate) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, number, s.User.ID, lease.LeaseID, lease.PropertyID, lease.RoomID, name, nullableString(phone), date, r.FormValue("starts_at"), r.FormValue("ends_at"), purpose, nullableString(plate))
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := res.LastInsertId()
	if err = notifyOwnerForPropertyTx(r.Context(), tx, lease.PropertyID, fmt.Sprintf("smart.visitor.created.%d", id), "Izin tamu baru", s.User.Name+" mengajukan kunjungan "+name+".", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.visitor.create", "visitor_pass", id, nil, map[string]any{"visitor": name, "visit_date": r.FormValue("visit_date")}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=tamu", http.StatusSeeOther)
}

func (a *App) ownerUpdateVisitorPass(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	next := r.FormValue("status")
	note := strings.TrimSpace(r.FormValue("note"))
	if e != nil || !map[string]bool{"approved": true, "rejected": true, "checked_in": true, "checked_out": true, "cancelled": true, "expired": true}[next] || len(note) > 1000 {
		http.Error(w, "Status izin tamu tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	if (next == "rejected" || next == "cancelled") && len(note) < 3 {
		http.Error(w, "Alasan wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var old, code string
	var visitDate time.Time
	var startsAt, endsAt string
	err = tx.QueryRowContext(r.Context(), `SELECT v.tenant_id,v.status,COALESCE(v.access_code,''),v.visit_date,DATE_FORMAT(v.starts_at,'%H:%i'),DATE_FORMAT(v.ends_at,'%H:%i') FROM visitor_passes v JOIN properties p ON p.id=v.property_id WHERE v.id=? AND p.owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &old, &code, &visitDate, &startsAt, &endsAt)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !visitorTransition(old, next) {
		http.Error(w, "Perubahan status izin tamu tidak diizinkan.", http.StatusConflict)
		return
	}
	if next == "checked_in" {
		startMinutes, _ := clockMinutes(startsAt)
		endMinutes, _ := clockMinutes(endsAt)
		nowMinutes := businessClockMinutes(a.location)
		if !sameBusinessDate(visitDate, businessToday(a.location)) || nowMinutes < startMinutes-30 || nowMinutes > endMinutes {
			http.Error(w, "Tamu hanya dapat ditandai masuk mulai 30 menit sebelum jadwal hingga waktu selesai.", http.StatusConflict)
			return
		}
	}
	if next == "approved" && code == "" {
		key, kerr := randomKey(5)
		if kerr != nil {
			a.serverError(w, r, kerr)
			return
		}
		code = strings.ToUpper(key)
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE visitor_passes SET status=?,owner_note=?,access_code=IF(?='',access_code,?),reviewed_by=IF(? IN ('approved','rejected'),?,reviewed_by),reviewed_at=IF(? IN ('approved','rejected'),UTC_TIMESTAMP(6),reviewed_at),checked_in_at=IF(?='checked_in',UTC_TIMESTAMP(6),checked_in_at),checked_out_at=IF(?='checked_out',UTC_TIMESTAMP(6),checked_out_at) WHERE id=?`, next, nullableString(note), code, code, next, s.User.ID, next, next, next, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	body := "Izin tamu diperbarui menjadi " + smartStatusLabel(next) + "."
	if next == "approved" {
		body += " Kode akses: " + code
	}
	if note != "" {
		body += " " + note
	}
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("smart.visitor.status.%d.%s", id, next), "Status izin tamu", body, "/penghuni/layanan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.visitor.status", "visitor_pass", id, map[string]any{"status": old}, map[string]any{"status": next}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=tamu", http.StatusSeeOther)
}

func (a *App) tenantCancelVisitorPass(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var old string
	err = tx.QueryRowContext(r.Context(), `SELECT property_id,status FROM visitor_passes WHERE id=? AND tenant_id=? FOR UPDATE`, id, s.User.ID).Scan(&propertyID, &old)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !map[string]bool{"pending": true, "approved": true}[old] {
		http.Error(w, "Izin tamu tidak dapat dibatalkan.", http.StatusConflict)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE visitor_passes SET status='cancelled',owner_note='Dibatalkan penghuni' WHERE id=?`, id); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyOwnerForPropertyTx(r.Context(), tx, propertyID, fmt.Sprintf("smart.visitor.cancelled.%d", id), "Izin tamu dibatalkan", s.User.Name+" membatalkan izin tamu.", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.visitor.cancel", "visitor_pass", id, map[string]any{"status": old}, map[string]any{"status": "cancelled"}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=batal-tamu", http.StatusSeeOther)
}

func (a *App) ownerCreateParcel(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	leaseID, e := parseID(r.FormValue("lease_id"))
	courier := strings.TrimSpace(r.FormValue("courier"))
	last4 := strings.ToUpper(strings.TrimSpace(r.FormValue("tracking_last4")))
	notes := strings.TrimSpace(r.FormValue("notes"))
	if e != nil || len(courier) < 2 || len(courier) > 100 || len(last4) > 8 || len(notes) > 500 {
		http.Error(w, "Data paket tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID, propertyID, roomID int64
	err = tx.QueryRowContext(r.Context(), `SELECT l.tenant_id,l.property_id,l.room_id FROM leases l JOIN properties p ON p.id=l.property_id WHERE l.id=? AND p.owner_id=? AND l.status IN ('active','expiring') FOR UPDATE`, leaseID, s.User.ID).Scan(&tenantID, &propertyID, &roomID)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Penghuni aktif tidak ditemukan.", http.StatusNotFound)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, _ := referenceNumber("PKT")
	res, err := tx.ExecContext(r.Context(), `INSERT INTO parcel_deliveries(parcel_number,tenant_id,lease_id,property_id,room_id,courier,tracking_last4,notes,status,logged_by,received_at) VALUES(?,?,?,?,?,?,?,?, 'notified',?,UTC_TIMESTAMP(6))`, number, tenantID, leaseID, propertyID, roomID, courier, nullableString(last4), nullableString(notes), s.User.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := res.LastInsertId()
	body := "Paket dari " + courier + " telah diterima dan menunggu di pengelola."
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("smart.parcel.arrived.%d", id), "Paket telah tiba", body, "/penghuni/layanan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.parcel.create", "parcel_delivery", id, nil, map[string]any{"courier": courier, "tenant_id": tenantID}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=paket", http.StatusSeeOther)
}

func (a *App) tenantCollectParcel(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var old string
	err = tx.QueryRowContext(r.Context(), `SELECT property_id,status FROM parcel_deliveries WHERE id=? AND tenant_id=? FOR UPDATE`, id, s.User.ID).Scan(&propertyID, &old)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !map[string]bool{"arrived": true, "notified": true}[old] {
		http.Error(w, "Paket tidak tersedia untuk dikonfirmasi.", http.StatusConflict)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE parcel_deliveries SET status='collected',collected_at=UTC_TIMESTAMP(6) WHERE id=?`, id); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyOwnerForPropertyTx(r.Context(), tx, propertyID, fmt.Sprintf("smart.parcel.collected.%d", id), "Paket telah diambil", s.User.Name+" mengonfirmasi paket sudah diambil.", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.parcel.collect", "parcel_delivery", id, map[string]any{"status": old}, map[string]any{"status": "collected"}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=paket-diambil", http.StatusSeeOther)
}

func (a *App) ownerUpdateParcel(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	next := r.FormValue("status")
	if e != nil || !map[string]bool{"collected": true, "returned": true}[next] {
		http.Error(w, "Status paket tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var old string
	err = tx.QueryRowContext(r.Context(), `SELECT pd.tenant_id,pd.status FROM parcel_deliveries pd JOIN properties p ON p.id=pd.property_id WHERE pd.id=? AND p.owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &old)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !map[string]bool{"arrived": true, "notified": true}[old] {
		http.Error(w, "Status paket tidak dapat diperbarui.", http.StatusConflict)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE parcel_deliveries SET status=?,collected_at=IF(?='collected',UTC_TIMESTAMP(6),collected_at) WHERE id=?`, next, next, id); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("smart.parcel.status.%d.%s", id, next), "Status paket diperbarui", "Paket ditandai "+smartStatusLabel(next)+" oleh pemilik.", "/penghuni/layanan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.parcel.status", "parcel_delivery", id, map[string]any{"status": old}, map[string]any{"status": next}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=status-paket", http.StatusSeeOther)
}

func (a *App) tenantCreateResidentRequest(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	kind := r.FormValue("request_type")
	notes := strings.TrimSpace(r.FormValue("notes"))
	dateText := strings.TrimSpace(r.FormValue("requested_date"))
	var date any
	today := businessToday(a.location)
	if dateText != "" {
		d, e := parseBusinessDate(dateText)
		if e != nil || d.Before(today) {
			http.Error(w, "Tanggal rencana tidak valid.", http.StatusUnprocessableEntity)
			return
		}
		date = d
	}
	if !map[string]bool{"renewal": true, "move_out": true, "room_move": true, "other": true}[kind] || len(notes) < 5 || len(notes) > 2000 || (kind == "move_out" && dateText == "") {
		http.Error(w, "Jenis, tanggal, atau catatan permintaan tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	lease, err := activeLeaseForTenantTx(r.Context(), tx, s.User.ID, today)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Permintaan hunian hanya tersedia saat kontrak aktif.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	var pending int
	if err = tx.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM resident_requests WHERE tenant_id=? AND lease_id=? AND request_type=? AND status IN ('submitted','reviewing','approved')`, s.User.ID, lease.LeaseID, kind).Scan(&pending); err != nil {
		a.serverError(w, r, err)
		return
	}
	if pending > 0 {
		http.Error(w, "Permintaan sejenis masih sedang diproses.", http.StatusConflict)
		return
	}
	number, _ := referenceNumber("LYN")
	res, err := tx.ExecContext(r.Context(), `INSERT INTO resident_requests(request_number,tenant_id,lease_id,property_id,room_id,request_type,requested_date,notes) VALUES(?,?,?,?,?,?,?,?)`, number, s.User.ID, lease.LeaseID, lease.PropertyID, lease.RoomID, kind, date, notes)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := res.LastInsertId()
	if err = notifyOwnerForPropertyTx(r.Context(), tx, lease.PropertyID, fmt.Sprintf("smart.request.created.%d", id), "Permintaan penghuni baru", s.User.Name+" mengirim permintaan "+smartRequestLabel(kind)+".", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.resident_request.create", "resident_request", id, nil, map[string]any{"type": kind}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=permintaan", http.StatusSeeOther)
}

func (a *App) ownerUpdateResidentRequest(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	next := r.FormValue("status")
	response := strings.TrimSpace(r.FormValue("response"))
	if e != nil || len(response) < 3 || len(response) > 2000 || !map[string]bool{"reviewing": true, "approved": true, "rejected": true, "completed": true}[next] {
		http.Error(w, "Status dan tanggapan owner wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var old, kind string
	err = tx.QueryRowContext(r.Context(), `SELECT rr.tenant_id,rr.status,rr.request_type FROM resident_requests rr JOIN properties p ON p.id=rr.property_id WHERE rr.id=? AND p.owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &old, &kind)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !residentRequestTransition(old, next) {
		http.Error(w, "Perubahan status permintaan tidak diizinkan.", http.StatusConflict)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE resident_requests SET status=?,owner_response=?,reviewed_by=?,reviewed_at=UTC_TIMESTAMP(6) WHERE id=?`, next, response, s.User.ID, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	body := smartRequestLabel(kind) + ": " + response
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("smart.request.status.%d.%s", id, next), "Permintaan hunian "+smartStatusLabel(next), body, "/penghuni/layanan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.resident_request.status", "resident_request", id, map[string]any{"status": old}, map[string]any{"status": next, "response": response}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=permintaan", http.StatusSeeOther)
}

func (a *App) tenantCancelResidentRequest(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	if e != nil {
		http.NotFound(w, r)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var propertyID int64
	var old string
	err = tx.QueryRowContext(r.Context(), `SELECT property_id,status FROM resident_requests WHERE id=? AND tenant_id=? FOR UPDATE`, id, s.User.ID).Scan(&propertyID, &old)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !map[string]bool{"submitted": true, "reviewing": true}[old] {
		http.Error(w, "Permintaan tidak dapat dibatalkan.", http.StatusConflict)
		return
	}
	if _, err = tx.ExecContext(r.Context(), `UPDATE resident_requests SET status='cancelled' WHERE id=?`, id); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyOwnerForPropertyTx(r.Context(), tx, propertyID, fmt.Sprintf("smart.request.cancelled.%d", id), "Permintaan dibatalkan penghuni", s.User.Name+" membatalkan permintaan hunian.", "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.resident_request.cancel", "resident_request", id, map[string]any{"status": old}, map[string]any{"status": "cancelled"}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=batal-permintaan", http.StatusSeeOther)
}

func (a *App) tenantCreateEmergencyAlert(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "tenant")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	category := r.FormValue("category")
	description := strings.TrimSpace(r.FormValue("description"))
	if !map[string]bool{"security": true, "fire": true, "medical": true, "electrical": true, "water": true, "other": true}[category] || len(description) < 5 || len(description) > 2000 {
		http.Error(w, "Kategori dan keterangan darurat tidak valid.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	lease, err := activeLeaseForTenantTx(r.Context(), tx, s.User.ID, businessToday(a.location))
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Laporan darurat memerlukan kontrak aktif.", http.StatusConflict)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	number, _ := referenceNumber("SOS")
	res, err := tx.ExecContext(r.Context(), `INSERT INTO emergency_alerts(alert_number,tenant_id,lease_id,property_id,room_id,category,description) VALUES(?,?,?,?,?,?,?)`, number, s.User.ID, lease.LeaseID, lease.PropertyID, lease.RoomID, category, description)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	id, _ := res.LastInsertId()
	if err = notifyOwnerForPropertyTx(r.Context(), tx, lease.PropertyID, fmt.Sprintf("smart.emergency.created.%d", id), "PERINGATAN DARURAT — "+lease.Room, s.User.Name+": "+description, "/pemilik/smart-kos"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.emergency.create", "emergency_alert", id, nil, map[string]any{"category": category}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/penghuni/layanan?berhasil=darurat", http.StatusSeeOther)
}

func (a *App) ownerUpdateEmergencyAlert(w http.ResponseWriter, r *http.Request) {
	s := a.requireRole(w, r, "owner")
	if s == nil {
		return
	}
	if err := r.ParseForm(); err != nil || !a.sessions.ValidCSRF(s, r.FormValue("csrf_token")) {
		http.Error(w, "Permintaan tidak valid.", http.StatusBadRequest)
		return
	}
	id, e := parseID(r.PathValue("id"))
	next := r.FormValue("status")
	response := strings.TrimSpace(r.FormValue("response"))
	if e != nil || len(response) < 3 || len(response) > 2000 || !map[string]bool{"acknowledged": true, "resolved": true, "cancelled": true}[next] {
		http.Error(w, "Status dan tanggapan darurat wajib diisi.", http.StatusUnprocessableEntity)
		return
	}
	tx, err := a.db.BeginTx(r.Context(), &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	defer tx.Rollback()
	var tenantID int64
	var old string
	err = tx.QueryRowContext(r.Context(), `SELECT ea.tenant_id,ea.status FROM emergency_alerts ea JOIN properties p ON p.id=ea.property_id WHERE ea.id=? AND p.owner_id=? FOR UPDATE`, id, s.User.ID).Scan(&tenantID, &old)
	if errors.Is(err, sql.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if !emergencyTransition(old, next) {
		http.Error(w, "Perubahan status darurat tidak diizinkan.", http.StatusConflict)
		return
	}
	_, err = tx.ExecContext(r.Context(), `UPDATE emergency_alerts SET status=?,owner_response=?,acknowledged_by=IF(?='acknowledged',?,acknowledged_by),acknowledged_at=IF(?='acknowledged',UTC_TIMESTAMP(6),acknowledged_at),resolved_by=IF(?='resolved',?,resolved_by),resolved_at=IF(?='resolved',UTC_TIMESTAMP(6),resolved_at) WHERE id=?`, next, response, next, s.User.ID, next, next, s.User.ID, next, id)
	if err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = notifyUserTx(r.Context(), tx, tenantID, fmt.Sprintf("smart.emergency.status.%d.%s", id, next), "Status bantuan darurat", response, "/penghuni/layanan"); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = insertAuditTx(r.Context(), tx, s.User.ID, "smart.emergency.status", "emergency_alert", id, map[string]any{"status": old}, map[string]any{"status": next, "response": response}, a.sessions.HashIP(r)); err != nil {
		a.serverError(w, r, err)
		return
	}
	if err = tx.Commit(); err != nil {
		a.serverError(w, r, err)
		return
	}
	http.Redirect(w, r, "/pemilik/smart-kos?berhasil=darurat", http.StatusSeeOther)
}

func notifyUserTx(ctx context.Context, tx *sql.Tx, userID int64, event, title, body, url string) error {
	if err := createNotificationTx(ctx, tx, userID, event, title, body, url); err != nil {
		return err
	}
	var email string
	if err := tx.QueryRowContext(ctx, `SELECT email FROM users WHERE id=? AND status='active'`, userID).Scan(&email); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `INSERT IGNORE INTO notification_deliveries(user_id,channel,event_key,recipient,subject,body) VALUES(?,'email',?,?,?,?)`, userID, event, email, title, body)
	return err
}

func notifyOwnerForPropertyTx(ctx context.Context, tx *sql.Tx, propertyID int64, event, title, body, url string) error {
	var ownerID int64
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM properties WHERE id=?`, propertyID).Scan(&ownerID); err != nil {
		return err
	}
	return notifyUserTx(ctx, tx, ownerID, event, title, body, url)
}

func clockMinutes(value string) (int, error) {
	v := strings.TrimSpace(value)
	if len(v) != 5 || v[2] != ':' {
		return 0, fmt.Errorf("format waktu harus JJ:MM")
	}
	t, err := time.Parse("15:04", v)
	if err != nil {
		return 0, err
	}
	return t.Hour()*60 + t.Minute(), nil
}

func businessClockMinutes(location *time.Location) int {
	now := time.Now().In(location)
	return now.Hour()*60 + now.Minute()
}

func reservationTransition(from, to string) bool {
	return map[string]map[string]bool{"pending": {"approved": true, "rejected": true, "cancelled": true}, "approved": {"completed": true, "cancelled": true}}[from][to]
}
func visitorTransition(from, to string) bool {
	return map[string]map[string]bool{"pending": {"approved": true, "rejected": true, "cancelled": true}, "approved": {"checked_in": true, "cancelled": true, "expired": true}, "checked_in": {"checked_out": true}}[from][to]
}
func residentRequestTransition(from, to string) bool {
	return map[string]map[string]bool{"submitted": {"reviewing": true, "approved": true, "rejected": true}, "reviewing": {"approved": true, "rejected": true}, "approved": {"completed": true}}[from][to]
}
func emergencyTransition(from, to string) bool {
	return map[string]map[string]bool{"open": {"acknowledged": true, "resolved": true, "cancelled": true}, "acknowledged": {"resolved": true, "cancelled": true}}[from][to]
}

func smartStatusLabel(status string) string {
	if v, ok := map[string]string{"pending": "menunggu persetujuan", "approved": "disetujui", "rejected": "ditolak", "checked_in": "tamu masuk", "checked_out": "tamu keluar", "cancelled": "dibatalkan", "expired": "kedaluwarsa", "completed": "selesai", "submitted": "diajukan", "reviewing": "sedang ditinjau", "open": "terbuka", "acknowledged": "ditanggapi", "resolved": "selesai"}[status]; ok {
		return v
	}
	return status
}
func smartRequestLabel(kind string) string {
	if v, ok := map[string]string{"renewal": "perpanjangan kontrak", "move_out": "rencana keluar", "room_move": "pindah kamar", "other": "layanan lainnya"}[kind]; ok {
		return v
	}
	return kind
}

func smartCategoryLabel(value string) string {
	if label, ok := map[string]string{
		"security": "Keamanan", "fire": "Kebakaran", "medical": "Medis",
		"electrical": "Kelistrikan", "water": "Air", "other": "Lainnya",
	}[value]; ok {
		return label
	}
	return value
}
