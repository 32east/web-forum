package api

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func RegisterURLs(db *sql.DB, rdb *redis.Client) {
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) { HandleLogin(&w, r, db, rdb) })
	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) { HandleRegister(&w, r, db) })
	http.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) { HandleLogout(&w, r, db, rdb) })
	http.HandleFunc("/api/send-message", func(w http.ResponseWriter, r *http.Request) { HandleMessage(&w, r, db, rdb) })
	http.HandleFunc("/api/profile/settings", func(w http.ResponseWriter, r *http.Request) { HandleSettings(&w, r, db, rdb) })
	http.HandleFunc("/api/topics/create", func(w http.ResponseWriter, r *http.Request) { HandleTopicCreate(&w, r, db, rdb) })
}
