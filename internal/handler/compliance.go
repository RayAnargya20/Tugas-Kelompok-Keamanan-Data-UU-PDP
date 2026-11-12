package handler

import (
	"aaa_bekend/internal/lmodel"
	"encoding/json"
	"net/http"
	"time"
)

func GetMyData(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())

	var profile lmodel.User
	err := db.DB.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = $1", user.ID).Scan(
		&profile.ID, &profile.Username, &profile.Email, &profile.CreatedAt,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(profile)
}

func UpdateMyData(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())
	var req struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err := db.DB.Exec("UPDATE users SET username = $1, email = $2 WHERE id = $3", req.Username, req.Email, user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Data updated successfully",
		"username": req.Username,
	})
}

func GetMyAuditLog(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())
	rows, err := db.DB.Query("SELECT id, action, path, ip_address, created_at FROM audit_logs WHERE user_id = $1 ORDER BY created_at DESC", user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id int
		var action, path, ipAddr string
		var createdAt time.Time
		if err := rows.Scan(&id, &action, &path, &ipAddr, &createdAt); err == nil {
			logs = append(logs, map[string]interface{}{
				"id":         id,
				"action":     action,
				"path":       path,
				"ip_address": ipAddr,
				"created_at": createdAt,
			})
		}
	}

	if logs == nil {
		logs = []map[string]interface{}{}
	} 

	json.NewEncoder(w).Encode(map[string]interface{}{
		"user":  user.Username,
		"total": len(logs),
		"logs":  logs,
	})
}

func ExportMyData(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())
	var profile lmodel.User
	db.DB.QueryRow("SELECT id, username, email, created_at FROM users WHERE id = $1", user.ID).Scan(
		&profile.ID, &profile.Username, &profile.Email, &profile.CreatedAt,
	)
	// Mengambil data proyek atau data terkait pengguna jika ada
	var projects []map[string]interface{}
	rowsProj, err := db.DB.Query("SELECT id, name, description FROM projects WHERE user_id = $1", user.ID)
	if err == nil {
		defer rowsProj.Close()
		for rowsProj.Next() {
			var pid int
			var pname, pdesc string
			if rowsProj.Scan(&pid, &pname, &pdesc) == nil {
				projects = append(projects, map[string]interface{}{"id": pid, "name": pname, "description": pdesc})
			}
		}
	}
	if projects == nil {
		projects = []map[string]interface{}{}
	}
	// Mengambil audit log pengguna
	var logs []map[string]interface{}
	rowsLog, err := db.DB.Query("SELECT id, action, path, created_at FROM audit_logs WHERE user_id = $1", user.ID)
	if err == nil {
		defer rowsLog.Close()
		for rowsLog.Next() {
			var lid int
			var laction, lpath string
			var ltime time.Time
			if rowsLog.Scan(&lid, &laction, &lpath, &ltime) == nil {
				logs = append(logs, map[string]interface{}{"id": lid, "action": laction, "path": lpath, "created_at": ltime})
			}
		}
	}
	if logs == nil {
		logs = []map[string]interface{}{}
	}
	export := map[string]interface{}{
		"exported_at": time.Now(),
		"profile":     profile,
		"projects":    projects,
		"audit_logs":  logs,
	}

	w.Header().Set("Content-Disposition", "attachment; filename=my_data_export.json")
	json.NewEncoder(w).Encode(export)
}
func DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	user, _ := lmodel.UserFromContext(r.Context())
	_, err := db.DB.Exec("DELETE FROM users WHERE id = $1", user.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Your account and all data have been permanently deleted",
	})
}
