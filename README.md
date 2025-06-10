# Payroll Management System

A simplified backend payroll system built with Go (Golang), PostgreSQL, and Gin. It handles employee salary computation based on attendance, overtime, and reimbursements, and supports JWT-based authentication with admin and employee roles.

## ✨ Features

- JWT-based authentication (admin and employee roles)
- Monthly payroll computation
- Attendance, overtime, and reimbursement tracking
- Payslip generation for employees
- Payroll summary for admin
- Audit logging of payroll runs

## 🔧 Technologies Used

- Go (Golang)
- PostgreSQL
- Gin web framework
- JWT for authentication
- Docker (optional, for PostgreSQL setup)

## 📦 Project Structure

```
├── cmd/
│   └── main.go              # Main application entry point
├── internal/
│   ├── handlers/            # Gin route handlers
│   │   ├── admin_payslip_summary.go
│   │   └── attendance_period.go
│   │   ├── attendance.go
│   │   └── auth.go
│   │   ├── employee_payslip.go
│   │   └── overtime.go
│   │   ├── payroll.go
│   │   └── reimbursement.go
│   ├── middleware/          # JWT auth middleware
│   │   └── auth.go
│   └── test/                # black-box tests
│   │   ├── admin_payslip_summary_test.go
│   │   └── attendance_period_test.go
│   │   ├── attendance_test.go
│   │   └── auth_test.go
│   │   ├── employee_payslip_test.go
│   │   └── overtime_test.go
│   │   ├── payroll_test.go
│   │   └── reimbursement_test.go
├── model/                   # Data models (e.g., User, Payslip)
├── migrations/              # SQL migration files
├── utils/                   # utilities functions
├── go.mod
├── go.sum
└── README.md

```

## 🔐 Authentication

This application uses **JWT (JSON Web Tokens)** for authentication:

- On successful login, a JWT is issued.
- Include the JWT token in the `Authorization` header as:
  ```
  Authorization: Bearer <your_token>
  ```
- Middleware extracts `userID` and `isAdmin` from the token and injects them into the context.

## 📑 API Endpoints (Sample)

### Admin
- `POST /admin/run-payroll/:period_id` — Run payroll for period
- `GET /admin/payslip-summary/:period_id` — Get summary of payslips
- `POST /admin/attendance-period/` — Run payroll for period

### Employee
- `GET /employee/payslip/:period_id` — Get employee payslip
- `POST /employee/attendance` — Submit attendance
- `POST /employee/overtime` — Submit overtime
- `POST /employee/reimbursement` — Submit reimbursement

### Auth
- `POST /login` — Login to receive JWT

## 🧪 Testing

The project includes scaffolding for automated tests using Go’s built-in `testing` package. You can run tests using:

```
go test ./...
```

### Test Files
- `test/employee_payslip_test.go`
- `test/admin_summary_test.go`
- `test/auth_test.go`
- `test/admin_payslip_summary_test.go`
- `test/attendance_period_test.go`
- `test/overtime_test.go`
- `test/payroll_test.go`
- `test/reimbursement_test.go`

## 🏁 Getting Started

### 1. Clone the Repository
```bash
git clone https://github.com/yourusername/payroll-project.git
cd payroll-project
```

### 2. Set up the Database

Use the migration SQL scripts in `migrations/` to create the necessary tables in PostgreSQL. Example with `psql`:

```bash
psql -U youruser -d yourdb -f migrations/create_db.sql
```
Optional: You can populate the users table with fake employee data up to 100 and set up employee level

```bash
psql -U youruser -d yourdb -f migrations/seed_users.sql
```

Or use a migration tool like [golang-migrate](https://github.com/golang-migrate/migrate).

### 3. Configure Environment Variables

Create a `.env` or use environment variables directly:
```
DB_URL=postgres://user:pass@localhost:5432/payroll_db
JWT_SECRET=your-secret-key
DB_USER=<your db user>
DB_PASSWORD=<your db password>
DB_NAME=payroll_db
DB_HOST=localhost
DB_PORT=5432
PORT=8000
DB_SSLMODE=disable
```

### 4. Run the App

```bash
go run main.go
```

## 📘 Documentation

### Payroll Computation Rules
- Base salary depends on employee level
- Prorated salary based on attendance
- Overtime is paid at 2x hourly rate
- Reimbursements are added directly
- Attendance period must be full month (e.g., 2025-06-01 to 2025-06-30)
- Attendance period id is constructed from the month and year (MM-YYYY) for readability and easier maintenance. ie: `06-2025`

### Code Organization
- Business logic is kept in handlers
- DB queries are written inline for simplicity
- `model/` includes response struct like `PayslipDetailResponse`, `AttendanceBreakdown`, `OvertimeBreakdown`, and `Reimbursement` which breakdown reimbursement items

### 🧱 Software Architecture

The project follows a simplified layered architecture:

- **Handlers (`handlers/`)**  
  Handle incoming HTTP requests and responses, parse input, and format output. Also perform minor validation and context extraction (e.g., from JWT).
  
- **Models (`model/`)**  
  Define data structures used across the application, such as `User`, `Payslip`, `Attendance`, and various response DTOs.
  
- **Middleware (`middleware/`)**  
  Handles cross-cutting concerns like JWT authentication, ensuring only authorized access to routes.
  
- **Migrations (`migrations/`)**  
  Contains SQL schema definitions and seed data to bootstrap the database.

- **Test (`test/`)**  
  Includes scaffolding and unit tests for key features and API endpoints.

- **Entry Point (`cmd/main.go`)**  
  Sets up the Gin engine, initializes routes, connects to the database, and runs the application.

#### Data Flow Overview

1. **Authentication**  
   User logs in and receives a JWT token.
2. **Protected Routes**  
   JWT middleware decodes token, sets `userID` and `role` in context.
3. **Handlers**  
   Based on role (admin or employee), handlers process request and run business logic.
4. **Database Interaction**  
   Inline SQL queries fetch and aggregate data (e.g., attendance, overtime, reimbursements).
5. **Payslip Generation**  
   When admin runs payroll, summary data is aggregated and inserted into the `payslips` table.




## ✅ Feature list
- [x] JWT auth
- [x] Admin & employee endpoints
- [x] Payroll engine
- [x] Audit log
- [x] Payslip summary for admin
- [x] Testing scaffolds
- [x] README documentation

## Points for improvement
- [ ] Approval flow for overtime and reimbursement
- [ ] User registration for new employees

## 📄 License
MIT
