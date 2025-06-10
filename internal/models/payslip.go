package models

import (
	"database/sql"
	"time"
)

type Payslip struct {
	ID                  string          `json:"id"`
	UserID              string          `json:"user_id"`
	Username            string          `json:"username"`
	AttendancePeriodID  string          `json:"attendance_period_id"`
	BaseSalary          float64         `json:"base_salary"`
	AttendanceAmount    float64         `json:"attendance_amount"`
	AttendanceDays      int             `json:"attendance_days"`
	OvertimeHours       sql.NullFloat64 `json:"overtime_hours"`
	OvertimeAmount      sql.NullFloat64 `json:"overtime_amount"`
	ReimbursementAmount sql.NullFloat64 `json:"reimbursement_amount"`
	TotalTakeHome       float64         `json:"total_take_home"`
	CreatedAt           time.Time       `json:"created_at"`
}

type PayslipDetailResponse struct {
	Payslip        Payslip             `json:"payslip"`
	Attendance     AttendanceBreakdown `json:"attendance"`
	Overtime       OvertimeBreakdown   `json:"overtime"`
	Reimbursements []Reimbursement     `json:"reimbursements"`
}

type AttendanceBreakdown struct {
	WorkingDays      int     `json:"working_days"`
	AttendanceDays   int     `json:"attendance_days"`
	AttendanceAmount float64 `json:"attendance_amount"`
}

type OvertimeBreakdown struct {
	OvertimeHours  float64 `json:"overtime_hours"`
	HourlyRate     float64 `json:"hourly_rate"`
	OvertimeRate   float64 `json:"overtime_rate"`
	OvertimeAmount float64 `json:"overtime_amount"`
}

type Reimbursement struct {
	ID          string    `json:"id"`
	Date        string    `json:"date"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	SubmittedAt time.Time `json:"submitted_at"`
}
