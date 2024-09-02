package www

import (
	"database/sql"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"web-forum/api/auth"
	"web-forum/api/profile"
	"web-forum/api/topics"
	"web-forum/frontend/web"
	"web-forum/www/handlers"
)

func RegisterStaticFiles(path string, urlPath string) {
	file, err := os.Open(path)

	if err != nil {
		log.Fatal(err)
	}

	fileServer, errDirRead := file.Readdir(-1)

	if errDirRead != nil {
		log.Fatal(errDirRead)
	}

	for _, value := range fileServer {
		if value.IsDir() {
			RegisterStaticFiles(path+"/"+value.Name(), urlPath+"/"+value.Name())
		} else {
			http.HandleFunc("/staticfiles/"+urlPath+"/"+value.Name(), func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, path+"/"+value.Name())
			})
		}
	}
}

func RegisterURLs(db *sql.DB, rdb *redis.Client) {
	//RegisterStaticFiles("frontend/template/imgs", "images")

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

		handlers.HandleMainPage(&w, r)
	})

	http.HandleFunc("/faq", func(w http.ResponseWriter, r *http.Request) { handlers.HandleFAQPage(&w, r) })
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) { handlers.HandleUsersPage(&w, r) })
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) { handlers.HandleLoginPage(&w, r) })
	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) { handlers.HandleRegisterPage(&w, r) })
	http.HandleFunc("/topic/create", func(w http.ResponseWriter, r *http.Request) { handlers.HandleTopicCreate(&w, r) })
	http.HandleFunc("/profile/settings", func(w http.ResponseWriter, r *http.Request) { handlers.HandleProfileSettings(&w, r) })

	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) { auth.HandleLogin(&w, r, db, rdb) })
	http.HandleFunc("/api/register", func(w http.ResponseWriter, r *http.Request) { auth.HandleRegister(&w, r, db) })
	http.HandleFunc("/api/logout", func(w http.ResponseWriter, r *http.Request) { auth.HandleLogout(&w, r, db, rdb) })
	http.HandleFunc("/api/send-message", func(w http.ResponseWriter, r *http.Request) { topics.HandleMessage(&w, r, db, rdb) })
	http.HandleFunc("/api/profile/settings", func(w http.ResponseWriter, r *http.Request) { profile.HandleSettings(&w, r, db, rdb) })
	http.HandleFunc("/api/topics/create", func(w http.ResponseWriter, r *http.Request) { topics.HandleTopicCreate(&w, r, db, rdb) })

	web.InitializeForumsPages(db, rdb)
	web.InitializeTopicsPages(db, rdb)

	http.ListenAndServe(":80", nil)
}
