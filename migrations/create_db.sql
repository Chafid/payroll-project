-- create_db.sql

-- Enable UUID generation functions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Drop existing tables if they exist (for dev reset)
DROP TABLE IF EXISTS reimbursements, overtimes, attendances, payslips, attendance_periods, audit_logs,  users, employee_levels CASCADE;

-- Employee level table
CREATE TABLE employee_levels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    base_salary NUMERIC NOT NULL
);

-- Users table - employees and admin use the same table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'employee')),
    level_id UUID REFERENCES employee_levels(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    created_by UUID,
    updated_by UUID
);

-- Attendances periods - only updated/created by admin
CREATE TABLE attendance_periods (
    id TEXT PRIMARY KEY,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_ip INET,
    updated_ip INET
);

-- Employee daily attendance - 1 per day max, no weekends)
CREATE TABLE attendances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    date DATE NOT NULL,
    period_id TEXT NOT NULL REFERENCES attendance_periods(id),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_ip INET,
    updated_ip INET,
    UNIQUE(user_id, date)
);

-- Overtime submissions
CREATE TABLE overtimes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    hours NUMERIC(4, 2) CHECK (hours > 0 AND hours <= 3),
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_ip INET,
    updated_ip INET,
    UNIQUE(user_id, date)
);

-- Reimbursement submissions
CREATE TABLE reimbursements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount NUMERIC(12, 2) NOT NULL CHECK (amount > 0),
    description TEXT,
    date DATE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    created_by UUID REFERENCES users(id),
    updated_by UUID REFERENCES users(id),
    created_ip INET,
    updated_ip INET
);

-- Payslip table - created once payroll is processed
CREATE TABLE payslips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    attendance_periods_id TEXT NOT NULL REFERENCES attendance_periods(id) ON DELETE CASCADE,
    base_salary NUMERIC(12, 2) NOT NULL,
    attendance_days INTEGER NOT NULL,
    attendance_amount NUMERIC(12, 2) NOT NULL,
    overtime_hours NUMERIC(8, 2) NULL,
    overtime_amount NUMERIC(12, 2) NULL,
    reimbursement_amount NUMERIC(12, 2) NULL,
    total_take_home NUMERIC(12, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    created_by UUID REFERENCES users(id),
    created_ip INET,
    UNIQUE(user_id, attendance_periods_id)
);

-- Audit log table
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL,
    action TEXT NOT NULL, -- 'INSERT', 'UPDATE', 'DELETE'
    user_id UUID NOT NULL REFERENCES users(id),
    changed_ip INET,
    change_data JSONB,
    created_at TIMESTAMPTZ DEFAULT now()
);
