package www

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"web-forum/api/admin"
	"web-forum/api/auth"
	"web-forum/api/profile"
	"web-forum/api/topics"
	"web-forum/internal"
	"web-forum/system"
	"web-forum/system/db"
	"web-forum/system/rdb"
	"web-forum/www/handlers"
	initialize_functions "web-forum/www/initialize-functions"
	"web-forum/www/middleware"
	"web-forum/www/services/category"
)

var ctx = context.Background()

func RegisterStaticFiles(path string, urlPath string) {
	const errorFunction = "RegisterStaticFiles"
	file, err := os.Open(path)

	if err != nil {
		system.FatalLog(errorFunction, err)
	}

	fileServer, errDirRead := file.Readdir(-1)

	if errDirRead != nil {
		system.FatalLog(errorFunction, errDirRead)
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

func RegisterURLs() {
	const errorFunction = "handlers.RegisterURLs"
	//RegisterStaticFiles("frontend/template/imgs", "images")

	for _, val := range []string{"imgs", "styles", "scripts"} {
		fileServer := http.FileServer(http.Dir(fmt.Sprintf("frontend/%s", val)))
		http.Handle(fmt.Sprintf("/%s/", val), http.StripPrefix("/"+val, fileServer))
	}

	middleware.Page("/", "Главная страница", handlers.MainPage)

	middleware.Page("/faq", "FAQ", handlers.FAQPage)
	middleware.Page("/users", "Юзеры", handlers.UsersPage)
	middleware.Page("/login", "Авторизация", handlers.LoginPage)
	middleware.Page("/register", "Регистрация", handlers.HandleRegisterPage)
	middleware.Page("/topic/create", "Создание нового топика", handlers.TopicCreate)
	middleware.Page("/profile/settings", "Настройки аккаунта", handlers.HandleProfileSettings)

	middleware.Page("/admin", "Админ панель", handlers.AdminMainPage)
	middleware.Page("/admin/categories", "Категории - Админ панель", handlers.AdminCategoriesPage)
	middleware.Page("/admin/users", "Юзеры - Админ панель", handlers.AdminUsersPage)

	middleware.API("/api/v1/auth/login", auth.HandleLogin)
	middleware.API("/api/v1/auth/register", auth.HandleRegister)
	middleware.API("/api/v1/auth/logout", auth.HandleLogout)
	middleware.API("/api/v1/auth/refresh-token", auth.HandleRefreshToken)

	middleware.API("/api/v1/topics/send-message", topics.HandleMessage)
	middleware.API("/api/v1/topics/create", topics.HandleTopicCreate)

	middleware.API("/api/v1/profile/settings", profile.HandleSettings)

	middleware.AdminAPI("/api/v1/admin/users/edit", "POST", admin.HandleProfileSettings)
	middleware.AdminAPI("/api/v1/admin/message/delete", "POST", admin.HandleMessageDelete)
	middleware.AdminAPI("/api/v1/admin/category/create", "POST", admin.HandleCategoryCreate)
	middleware.AdminAPI("/api/v1/admin/category/edit", "POST", admin.HandleCategoryEdit)
	middleware.AdminAPI("/api/v1/admin/category/delete", "POST", admin.HandleCategoryDelete)

	initialize_functions.Categorys()
	initialize_functions.Topics()
	initialize_functions.Profiles()

	category.GetAll() // Инициализируем

	var countInfo = internal.CountStruct{}
	db.Postgres.QueryRow(ctx, `select 
    	(select count(*) from users),
    	(select count(*) from topics),
    	(select count(*) from messages);
    	`).Scan(&countInfo.Users, &countInfo.Topics, &countInfo.Messages)

	rdb.RedisDB.Set(ctx, "count:users", countInfo.Users, 0)
	rdb.RedisDB.Set(ctx, "count:topics", countInfo.Topics, 0)
	rdb.RedisDB.Set(ctx, "count:messages", countInfo.Messages, 0)

	httpErr := http.ListenAndServe(":8081", nil)

	system.FatalLog(errorFunction, httpErr)
}
