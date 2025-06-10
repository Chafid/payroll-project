package utils

import (
	"database/sql"
	"log"
	"net"

	"github.com/google/uuid"
)

func LogAudit(db *sql.DB, action, table string, recordID string, userID uuid.UUID, ip net.IP, changeData []byte) {
	ipAddress := ip.String()
	_, err := db.Exec(`
		INSERT INTO audit_logs (
			table_name, record_id, action, user_id, changed_ip, change_data
		) VALUES ($1, $2, $3, $4, $5::inet, $6)
	`, table, recordID, action, userID, ipAddress, changeData)

	if err != nil {
		log.Printf("[AuditLog] Failed to insert audit log: %v", err)
	}
}
