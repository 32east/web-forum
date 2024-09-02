package web

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"net/http"
)

func ExecuteNewServer(db *sql.DB, rdb *redis.Client) {
	InitializeForumsPages(db, rdb)
	InitializeTopicsPages(db, rdb)

	// CSS файлы
	fileServer := http.FileServer(http.Dir("frontend/template/imgs"))
	http.Handle("/imgs/", http.StripPrefix("/imgs", fileServer))

	// CSS файлы
	fileServer = http.FileServer(http.Dir("frontend/template/styles"))
	http.Handle("/styles/", http.StripPrefix("/styles", fileServer))

	// JS файлы
	fileServer = http.FileServer(http.Dir("frontend/template/scripts"))
	http.Handle("/scripts/", http.StripPrefix("/scripts", fileServer))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		HandleMainPage(&w, r, db, rdb)
	})

	http.HandleFunc("/faq", func(w http.ResponseWriter, r *http.Request) { HandleFAQPage(&w, r, db, rdb) })
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) { HandleUsersPage(&w, r, db, rdb) })
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { HandleLoginPage(&w, r, db, rdb) })
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) { HandleRegisterPage(&w, r, db, rdb) })
	http.HandleFunc("/topic/create", func(w http.ResponseWriter, r *http.Request) { HandleTopicCreate(&w, r, db, rdb) })
	http.HandleFunc("/profile/settings", func(w http.ResponseWriter, r *http.Request) { HandleProfileSettings(&w, r, db, rdb) })

	http.ListenAndServe(":80", nil)
}
