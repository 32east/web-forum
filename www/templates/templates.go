package templates

import (
	"bytes"
	"html/template"
	"reflect"
)

func Prepare(additionalFiles ...string) *template.Template {
	parseFiles, err := template.ParseFiles(additionalFiles...)

	if err != nil {
		panic(err)
	}

	return template.Must(parseFiles, err)
}

func ContentAdd(infoToSend *map[string]interface{}, tmpl *template.Template, content any) {
	if reflect.ValueOf(content).Kind() == reflect.Map {
		for k, v := range *infoToSend {
			content.(map[string]interface{})[k] = v
		}
	}

	newBytesBuffer := new(bytes.Buffer)
	tmpl.Execute(newBytesBuffer, content)

	(*infoToSend)["Content"] = template.HTML(newBytesBuffer.String())
}

var Index = Prepare("frontend/forum/index.html")
var LoginPage = Prepare("frontend/forum/login.html")
var RegisterPage = Prepare("frontend/forum/register.html")
var Forum = Prepare("frontend/forum/forum.html")
var Topics = Prepare("frontend/forum/topics.html")
var FAQ = Prepare("frontend/forum/faq.html")
var Users = Prepare("frontend/forum/users.html")

var TopicPage = Prepare("frontend/forum/topic.html", "frontend/forum/topic-textarea.html")

var ProfileSettings = Prepare("frontend/forum/profile-settings.html", "frontend/forum/not-authorized.html")
var CreateNewTopic = Prepare("frontend/forum/create-new-topic.html", "frontend/forum/not-authorized.html")
var Profile = Prepare("frontend/forum/profile.html", "frontend/forum/not-authorized.html")

var AdminMain = Prepare("frontend/admin/index.html", "frontend/admin/main.html")
var AdminCategories = Prepare("frontend/admin/index.html", "frontend/admin/categories.html")
