package transport

import (
	"context"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	admin "web-forum/internal/app/api/admin"
	"web-forum/internal/app/api/auth"
	"web-forum/internal/app/api/profile"
	"web-forum/internal/app/api/topics"
	"web-forum/internal/app/database/db"
	"web-forum/internal/app/database/rdb"
	"web-forum/internal/app/handlers"
	"web-forum/internal/app/middleware"
	"web-forum/internal/app/models"
	"web-forum/internal/app/services/category"
	url_init "web-forum/internal/app/transport/url-init"
	"web-forum/pkg/stuff"
)

var ctx = context.Background()

func RegisterURLs() {
	const errorFunction = "handlers.RegisterURLs"

	for _, val := range []string{"imgs", "styles", "scripts"} {
		var fileServer = http.FileServer(http.Dir(fmt.Sprintf("www/staticfiles/%s", val)))
		http.Handle(fmt.Sprintf("/%s/", val), http.StripPrefix("/"+val, fileServer))
	}

	middleware.Page("/", "Главная страница", handlers.MainPage)
	middleware.Page("/login", "Авторизация", handlers.LoginPage)
	middleware.Page("/register", "Регистрация", handlers.HandleRegisterPage)
	middleware.Page("/topic/create", "Создание нового топика", handlers.TopicCreate)
	middleware.Page("/profile/settings", "Настройки аккаунта", handlers.HandleProfileSettings)

	middleware.AdminPage("/admin", "Админ панель", handlers.AdminMainPage)
	middleware.AdminPage("/admin/categories", "Категории - Админ панель", handlers.AdminCategoriesPage)
	middleware.AdminPage("/admin/users", "Юзеры - Админ панель", handlers.AdminUsersPage)

	middleware.API("/api/v1/auth/login", auth.HandleLogin)
	middleware.API("/api/v1/auth/register", auth.HandleRegister)
	middleware.API("/api/v1/auth/logout", auth.HandleLogout)
	middleware.API("/api/v1/auth/refresh-token", auth.HandleRefreshToken)

	middleware.API("/api/v1/topics/send-message", topics.HandleMessage)
	middleware.API("/api/v1/topics/create", topics.HandleTopicCreate)

	middleware.API("/api/v1/profile/settings", profile.HandleSettings)

	middleware.AdminAPI("/api/v1/admin/category/create", "POST", admin.HandleCategoryCreate)

	middleware.AdminAPI("/api/v1/admin/users/edit", "POST", admin.HandleProfileSettings)
	middleware.AdminAPI("/api/v1/admin/category/edit", "POST", admin.HandleCategoryEdit)

	middleware.AdminAPI("/api/v1/admin/category/delete", "POST", admin.HandleCategoryDelete)
	middleware.AdminAPI("/api/v1/admin/message/delete", "POST", admin.HandleMessageDelete)
	middleware.AdminAPI("/api/v1/admin/topics/delete", "POST", admin.HandleTopicDelete)
	middleware.AdminAPI("/api/v1/admin/users/delete", "POST", admin.HandleProfileDelete)

	url_init.Categorys()
	url_init.Topics()
	url_init.Profiles()

	category.GetAll() // Инициализируем

	var countInfo = models.CountStruct{}
	db.Postgres.QueryRow(ctx, `select 
    	(select count(*) from users),
    	(select count(*) from topics),
    	(select count(*) from messages);
    	`).Scan(&countInfo.Users, &countInfo.Topics, &countInfo.Messages)

	rdb.RedisDB.Set(ctx, "count:users", countInfo.Users, 0)
	rdb.RedisDB.Set(ctx, "count:topics", countInfo.Topics, 0)
	rdb.RedisDB.Set(ctx, "count:messages", countInfo.Messages, 0)

	httpErr := http.ListenAndServe(":8080", nil)

	stuff.FatalLog(errorFunction, httpErr)
}
