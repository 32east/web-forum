package templates

import (
	"bytes"
	"html/template"
	"reflect"
)

func ParseFiles(page string) (*template.Template, error) {
	return template.ParseFiles(page, "frontend/not-authorized.html", "frontend/topic-textarea.html")
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

var Index = template.Must(ParseFiles("frontend/index.html"))
var LoginPage = template.Must(ParseFiles("frontend/login.html"))
var RegisterPage = template.Must(ParseFiles("frontend/register.html"))
var ProfileSettings = template.Must(ParseFiles("frontend/profile-settings.html"))
var Forum = template.Must(ParseFiles("frontend/forum.html"))
var Topics = template.Must(ParseFiles("frontend/topics.html"))
var TopicPage = template.Must(ParseFiles("frontend/topic.html"))
var CreateNewTopic = template.Must(ParseFiles("frontend/create-new-topic.html"))
var FAQ = template.Must(ParseFiles("frontend/faq.html"))
var Users = template.Must(ParseFiles("frontend/users.html"))
