package repository

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"simrs-asoy/backend/internal/domain"
)

type AppointmentRepository struct {
	db *sqlx.DB
}

func NewAppointmentRepository(db *sqlx.DB) *AppointmentRepository {
	return &AppointmentRepository{db: db}
}

func (r *AppointmentRepository) FindAll(date, status string, doctorID string, page, limit int) ([]domain.Appointment, int, error) {
	offset := (page - 1) * limit
	whereClauses := []string{"a.deleted_at IS NULL"}
	args := []interface{}{}
	argIndex := 1

	if date != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.appointment_date = $%d", argIndex))
		args = append(args, date)
		argIndex++
	}

	if status != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if doctorID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("a.doctor_id = $%d", argIndex))
		args = append(args, doctorID)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM appointments a WHERE %s`, whereSQL)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT a.id, a.patient_id, a.poli_id, a.doctor_id, TO_CHAR(a.appointment_date, 'YYYY-MM-DD') AS appointment_date,
		       a.start_time::text, a.end_time::text, a.status, a.channel, COALESCE(a.catatan, '') AS catatan, a.checked_in_at, a.checked_in_by,
		       a.visit_id, a.created_at,
		       p.nama_pasien AS patient_name, p.no_rm, COALESCE(p.no_hp, '') AS no_hp,
		       pu.name AS poli_name, COALESCE(u.full_name, '') AS doctor_name
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN poli_units pu ON a.poli_id = pu.id
		LEFT JOIN users u ON a.doctor_id = u.id
		WHERE %s
		ORDER BY a.appointment_date ASC, a.start_time ASC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, limit, offset)
	var appointments []domain.Appointment
	err = r.db.Select(&appointments, query, args...)
	if err != nil {
		return nil, 0, err
	}

	return appointments, total, nil
}

func (r *AppointmentRepository) FindByID(id string) (*domain.Appointment, error) {
	query := `
		SELECT a.id, a.patient_id, a.poli_id, a.doctor_id, TO_CHAR(a.appointment_date, 'YYYY-MM-DD') AS appointment_date,
		       a.start_time::text, a.end_time::text, a.status, a.channel, COALESCE(a.catatan, '') AS catatan, a.checked_in_at, a.checked_in_by,
		       a.visit_id, a.created_at,
		       p.nama_pasien AS patient_name, p.no_rm, COALESCE(p.no_hp, '') AS no_hp,
		       pu.name AS poli_name, COALESCE(u.full_name, '') AS doctor_name
		FROM appointments a
		JOIN patients p ON a.patient_id = p.id
		JOIN poli_units pu ON a.poli_id = pu.id
		LEFT JOIN users u ON a.doctor_id = u.id
		WHERE a.id = $1 AND a.deleted_at IS NULL
	`
	var a domain.Appointment
	err := r.db.Get(&a, query, id)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *AppointmentRepository) Create(a *domain.Appointment) error {
	query := `
		INSERT INTO appointments (patient_id, poli_id, doctor_id, appointment_date, start_time, end_time,
		                          channel, catatan, status, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'TERJADWAL', $9, NOW(), NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, a.PatientID, a.PoliID, a.DoctorID, a.AppointmentDate, a.StartTime, a.EndTime,
		a.Channel, a.Catatan, a.DoctorID).Scan(&a.ID, &a.CreatedAt)
}

func (r *AppointmentRepository) Reschedule(id, newDate, startTime, endTime, notes string) error {
	query := `
		UPDATE appointments 
		SET appointment_date = $1, start_time = $2, end_time = $3, catatan = $4, status = 'RESCHEDULE', updated_at = NOW()
		WHERE id = $5
	`
	_, err := r.db.Exec(query, newDate, startTime, endTime, notes, id)
	return err
}

func (r *AppointmentRepository) Cancel(id, notes string) error {
	query := `UPDATE appointments SET status = 'BATAL', catatan = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, notes, id)
	return err
}

func (r *AppointmentRepository) CheckIn(id, checkedInBy, visitID string) error {
	query := `
		UPDATE appointments 
		SET status = 'HADIR', checked_in_at = NOW(), checked_in_by = $1, visit_id = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := r.db.Exec(query, checkedInBy, visitID, id)
	return err
}

// -------------------------------------------------------------
// Doctor Schedules & Leaves (Fase 1)
// -------------------------------------------------------------

func (r *AppointmentRepository) GetDoctorSchedules(doctorID string, poliID int, dayOfWeek int) ([]domain.DoctorSchedule, error) {
	where := []string{"ds.status_aktif = TRUE"}
	args := []interface{}{}
	idx := 1

	if doctorID != "" {
		where = append(where, fmt.Sprintf("ds.doctor_id = $%d", idx))
		args = append(args, doctorID)
		idx++
	}
	if poliID > 0 {
		where = append(where, fmt.Sprintf("ds.poli_id = $%d", idx))
		args = append(args, poliID)
		idx++
	}
	if dayOfWeek >= 0 {
		where = append(where, fmt.Sprintf("ds.day_of_week = $%d", idx))
		args = append(args, dayOfWeek)
		idx++
	}

	query := fmt.Sprintf(`
		SELECT ds.id, ds.doctor_id, u.full_name AS doctor_name, ds.poli_id, p.name AS poli_name,
		       ds.day_of_week, ds.start_time::TEXT, ds.end_time::TEXT, ds.max_quota, ds.status_aktif, ds.created_at
		FROM doctor_schedules ds
		JOIN users u ON ds.doctor_id = u.id
		JOIN poli_units p ON ds.poli_id = p.id
		WHERE %s
		ORDER BY ds.day_of_week ASC, ds.start_time ASC
	`, strings.Join(where, " AND "))

	var list []domain.DoctorSchedule
	err := r.db.Select(&list, query, args...)
	return list, err
}

func (r *AppointmentRepository) CreateDoctorSchedule(s *domain.DoctorSchedule) error {
	query := `
		INSERT INTO doctor_schedules (doctor_id, poli_id, day_of_week, start_time, end_time, max_quota, status_aktif, created_at)
		VALUES ($1, $2, $3, $4::TIME, $5::TIME, $6, TRUE, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, s.DoctorID, s.PoliID, s.DayOfWeek, s.StartTime, s.EndTime, s.MaxQuota).Scan(&s.ID, &s.CreatedAt)
}

func (r *AppointmentRepository) DeleteDoctorSchedule(id int) error {
	_, err := r.db.Exec("UPDATE doctor_schedules SET status_aktif = FALSE WHERE id = $1", id)
	return err
}

func (r *AppointmentRepository) GetDoctorLeaves(doctorID string, date string) ([]domain.DoctorLeave, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if doctorID != "" {
		where = append(where, fmt.Sprintf("dl.doctor_id = $%d", idx))
		args = append(args, doctorID)
		idx++
	}
	if date != "" {
		where = append(where, fmt.Sprintf("$%d::DATE BETWEEN dl.start_date AND dl.end_date", idx))
		args = append(args, date)
		idx++
	}

	query := fmt.Sprintf(`
		SELECT dl.id, dl.doctor_id, u.full_name AS doctor_name,
		       dl.start_date::TEXT, dl.end_date::TEXT, dl.reason,
		       dl.substitute_doctor_id, COALESCE(sub.full_name, '') AS substitute_doctor_name,
		       dl.created_at
		FROM doctor_leaves dl
		JOIN users u ON dl.doctor_id = u.id
		LEFT JOIN users sub ON dl.substitute_doctor_id = sub.id
		WHERE %s
		ORDER BY dl.start_date DESC
	`, strings.Join(where, " AND "))

	var list []domain.DoctorLeave
	err := r.db.Select(&list, query, args...)
	return list, err
}

func (r *AppointmentRepository) CreateDoctorLeave(l *domain.DoctorLeave) error {
	query := `
		INSERT INTO doctor_leaves (doctor_id, start_date, end_date, reason, substitute_doctor_id, created_at)
		VALUES ($1, $2::DATE, $3::DATE, $4, $5, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, l.DoctorID, l.StartDate, l.EndDate, l.Reason, l.SubstituteDoctorID).Scan(&l.ID, &l.CreatedAt)
}
