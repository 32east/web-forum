package templates

import (
	"bytes"
	"html/template"
	"net/http"
	"reflect"
)

func Prepare(additionalFiles ...string) *template.Template {
	parseFiles, err := template.ParseFiles(additionalFiles...)

	if err != nil {
		panic(err)
	}

	return template.Must(parseFiles, err)
}

func ContentAdd(reader *http.Request, tmpl *template.Template, content any) {
	InfoToSend := reader.Context().Value("InfoToSend").(map[string]interface{})

	if reflect.ValueOf(content).Kind() == reflect.Map {
		for k, v := range InfoToSend {
			content.(map[string]interface{})[k] = v
		}
	}

	newBytesBuffer := new(bytes.Buffer)
	tmpl.Execute(newBytesBuffer, content)

	InfoToSend["Content"] = template.HTML(newBytesBuffer.String())
}

var Index = Prepare("frontend/forum/index.html")
var LoginPage = Prepare("frontend/forum/login.html")
var RegisterPage = Prepare("frontend/forum/register.html")
var Forum = Prepare("frontend/forum/forum.html")
var Topics = Prepare("frontend/forum/topics.html", "frontend/common/topic.html")
var FAQ = Prepare("frontend/forum/faq.html")
var Users = Prepare("frontend/forum/users.html")

var TopicPage = Prepare("frontend/forum/topic.html", "frontend/common/message.html")

var ProfileSettings = Prepare("frontend/forum/profile-settings.html", "frontend/common/not-authorized.html")
var CreateNewTopic = Prepare("frontend/forum/create-new-topic.html", "frontend/common/not-authorized.html")
var Profile = Prepare("frontend/forum/profile.html", "frontend/common/not-authorized.html")

var AdminMain = Prepare("frontend/admin/index.html", "frontend/admin/main.html", "frontend/common/user-admin-panel.html", "frontend/common/user-admin-settings.html")
var AdminCategories = Prepare("frontend/admin/index.html", "frontend/admin/categories.html")
var AdminUsers = Prepare("frontend/admin/index.html", "frontend/admin/users.html", "frontend/common/user-admin-panel.html", "frontend/common/user-admin-settings.html")
