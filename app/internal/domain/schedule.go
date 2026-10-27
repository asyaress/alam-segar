package domain

import (
	"errors"
	"time"
)

type LeaseCharge struct {
	Number       int
	DueDate      time.Time
	ServiceStart time.Time
	ServiceEnd   time.Time
	Amount       int64
}

type RevenuePeriod struct {
	Number          int
	RecognitionDate time.Time
	ServiceStart    time.Time
	ServiceEnd      time.Time
	Amount          int64
}

func AnchoredDate(year int, month time.Month, anchorDay int, location *time.Location) (time.Time, error) {
	if anchorDay < 1 || anchorDay > 31 {
		return time.Time{}, errors.New("tanggal acuan harus antara 1 dan 31")
	}
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, location).Day()
	day := anchorDay
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, location), nil
}

func AddMonthsAnchored(base time.Time, months int, originalAnchorDay int) (time.Time, error) {
	target := time.Date(base.Year(), base.Month()+time.Month(months), 1, 0, 0, 0, 0, base.Location())
	return AnchoredDate(target.Year(), target.Month(), originalAnchorDay, base.Location())
}

func SplitInstallments(total int64, count int) ([]int64, error) {
	if total <= 0 {
		return nil, errors.New("total harus lebih dari nol")
	}
	if count <= 0 || count > 12 {
		return nil, errors.New("jumlah angsuran harus antara 1 dan 12")
	}
	base := total / int64(count)
	parts := make([]int64, count)
	for i := range parts {
		parts[i] = base
	}
	parts[count-1] += total - base*int64(count)
	return parts, nil
}

func MonthsBetweenAnchored(start, end time.Time, anchorDay int) (int, error) {
	if !end.After(start) {
		return 0, errors.New("tanggal selesai harus setelah tanggal mulai")
	}
	for months := 1; months <= 240; months++ {
		candidate, err := AddMonthsAnchored(start, months, anchorDay)
		if err != nil {
			return 0, err
		}
		if candidate.Equal(end) {
			return months, nil
		}
		if candidate.After(end) {
			return 0, errors.New("tanggal selesai harus mengikuti tanggal acuan kontrak")
		}
	}
	return 0, errors.New("periode kontrak maksimal 240 bulan")
}

func BuildLeasePlan(start, end time.Time, anchorDay, dueDay int, cycle string, agreedRent int64, installmentCount int) ([]LeaseCharge, []RevenuePeriod, error) {
	months, err := MonthsBetweenAnchored(start, end, anchorDay)
	if err != nil {
		return nil, nil, err
	}
	if dueDay < 1 || dueDay > 31 || agreedRent <= 0 {
		return nil, nil, errors.New("tanggal jatuh tempo atau tarif tidak valid")
	}
	var total int64
	var charges []LeaseCharge
	switch cycle {
	case "monthly":
		if installmentCount != 1 {
			return nil, nil, errors.New("kontrak bulanan tidak memakai cicilan tahunan")
		}
		if agreedRent > (1<<63-1)/int64(months) {
			return nil, nil, errors.New("total nilai kontrak terlalu besar")
		}
		total = agreedRent * int64(months)
		for index := 0; index < months; index++ {
			serviceStart, _ := AddMonthsAnchored(start, index, anchorDay)
			serviceEnd, _ := AddMonthsAnchored(start, index+1, anchorDay)
			due, _ := AnchoredDate(serviceStart.Year(), serviceStart.Month(), dueDay, start.Location())
			charges = append(charges, LeaseCharge{Number: index + 1, DueDate: due, ServiceStart: serviceStart, ServiceEnd: serviceEnd, Amount: agreedRent})
		}
	case "annual":
		if months != 12 {
			return nil, nil, errors.New("kontrak tahunan harus tepat 12 bulan sesuai tanggal acuan")
		}
		valid := map[int]bool{1: true, 2: true, 4: true, 6: true, 12: true}
		if !valid[installmentCount] {
			return nil, nil, errors.New("skema cicilan tahunan harus 1, 2, 4, 6, atau 12 kali")
		}
		total = agreedRent
		parts, _ := SplitInstallments(total, installmentCount)
		for index, amount := range parts {
			monthOffset := index * 12 / installmentCount
			base, _ := AddMonthsAnchored(start, monthOffset, anchorDay)
			due, _ := AnchoredDate(base.Year(), base.Month(), dueDay, start.Location())
			charges = append(charges, LeaseCharge{Number: index + 1, DueDate: due, ServiceStart: start, ServiceEnd: end, Amount: amount})
		}
	default:
		return nil, nil, errors.New("jenis kontrak harus bulanan atau tahunan")
	}
	revenueParts, _ := SplitInstallments(total, months)
	revenue := make([]RevenuePeriod, 0, months)
	for index, amount := range revenueParts {
		periodStart, _ := AddMonthsAnchored(start, index, anchorDay)
		periodEnd, _ := AddMonthsAnchored(start, index+1, anchorDay)
		revenue = append(revenue, RevenuePeriod{Number: index + 1, RecognitionDate: periodEnd.AddDate(0, 0, -1), ServiceStart: periodStart, ServiceEnd: periodEnd, Amount: amount})
	}
	return charges, revenue, nil
}
