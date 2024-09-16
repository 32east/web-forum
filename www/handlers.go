package www

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"web-forum/api/admin"
	"web-forum/api/auth"
	"web-forum/api/profile"
	"web-forum/api/topics"
	"web-forum/system"
	"web-forum/www/handlers"
	initialize_functions "web-forum/www/initialize-functions"
	"web-forum/www/middleware"
)

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
	middleware.Page("/admin/categories/create", "Категории - Админ панель", handlers.AdminCategoryCreate)

	middleware.API("/api/login", auth.HandleLogin)
	middleware.API("/api/register", auth.HandleRegister)
	middleware.API("/api/logout", auth.HandleLogout)
	middleware.API("/api/refresh-token", auth.HandleRefreshToken)

	middleware.API("/api/send-message", topics.HandleMessage)
	middleware.API("/api/profile/settings", profile.HandleSettings)
	middleware.API("/api/topics/create", topics.HandleTopicCreate)

	middleware.API("/api/admin/category/create", admin.HandleCategoryCreate)
	// middleware.API("/api/admin/category/edit", admin.HandleCategoryCreate)

	initialize_functions.Categorys()
	initialize_functions.Topics()
	initialize_functions.Profiles()

	httpErr := http.ListenAndServe(":80", nil)

	system.FatalLog(errorFunction, httpErr)
}
