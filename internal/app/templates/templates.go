package templates

import (
	"bytes"
	"html/template"
	"net/http"
	"reflect"
)

func Prepare(additionalFiles ...string) *template.Template {
	//for key, val := range additionalFiles {
	//	additionalFiles[key] = "../../" + val
	//}

	var parseFiles, err = template.ParseFiles(additionalFiles...)

	if err != nil {
		panic(err)
	}

	return template.Must(parseFiles, err)
}

func ContentAdd(reader *http.Request, tmpl *template.Template, content any) {
	var InfoToSend = reader.Context().Value("InfoToSend").(map[string]interface{})

	if reflect.ValueOf(content).Kind() == reflect.Map {
		for k, v := range InfoToSend {
			content.(map[string]interface{})[k] = v
		}
	}

	var newBytesBuffer = new(bytes.Buffer)
	tmpl.Execute(newBytesBuffer, content)

	InfoToSend["Content"] = template.HTML(newBytesBuffer.String())
}

var Index = Prepare("www/main/index.html")
var LoginPage = Prepare("www/main/login.html")
var RegisterPage = Prepare("www/main/register.html")
var Forum = Prepare("www/main/forum.html")
var Topics = Prepare("www/main/topics.html", "www/common/topic.html")

var TopicPage = Prepare("www/main/topic.html", "www/common/message.html")

var ProfileSettings = Prepare("www/main/profile-settings.html", "www/common/not-authorized.html")
var CreateNewTopic = Prepare("www/main/create-new-topic.html", "www/common/not-authorized.html")
var Profile = Prepare("www/main/profile.html", "www/common/not-authorized.html")

var AdminMain = Prepare("www/admin/index.html", "www/admin/main.html", "www/common/user-admin-block.html", "www/common/user-admin-settings.html")
var AdminCategories = Prepare("www/admin/index.html", "www/admin/categories.html")
var AdminUsers = Prepare("www/admin/index.html", "www/admin/users.html", "www/common/user-admin-block.html", "www/common/user-admin-settings.html")
