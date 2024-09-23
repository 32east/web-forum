package tests

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestAuth(t *testing.T) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8080/api/v1/auth/login", nil)

	if err != nil {
		t.Fatal(err)
	}

	res, errDo := client.Do(req)

	if errDo != nil {
		t.Fatal(errDo)
	}

	body, errRead := io.ReadAll(res.Body)

	if errRead != nil {
		t.Fatal(errRead)
	}

	var output map[string]interface{}
	unmarshalErr := json.Unmarshal(body, &output)

	if unmarshalErr != nil {
		t.Fatal(unmarshalErr)
	}

	if res.StatusCode != 200 {
		t.Errorf("Failed on test 1: status code is %v", res.StatusCode)
	} else {
		t.Logf("Success on test 1: status code is %v", res.StatusCode)
	}

	val, ok := output["success"]

	if !ok {
		t.Errorf("Failed on test 2: success is nil")
	} else {
		t.Logf("Success on test 2: success is NOT nil")
	}

	if val == false {
		t.Logf("Success on test 3: success is %v", val)
	} else {
		t.Errorf("Failed on test 3: success is %v", val)
	}

	req, _ = http.NewRequest("POST", "http://localhost:8080/api/v1/auth/login", nil)
	res, _ = client.Do(req)

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Failed on test 4: status code is %v", res.StatusCode)
	} else {
		t.Logf("Success on test 4: status code is %v", res.StatusCode)
	}

	body, errRead = io.ReadAll(res.Body)

	if errRead != nil {
		t.Errorf("Failed on test 5: %s", errRead)
	} else {
		t.Logf("Success on test 5: body readed %s", string(body))
	}

	var outputPost map[string]interface{}
	unmarshalPostErr := json.Unmarshal(body, &outputPost)

	if unmarshalPostErr != nil {
		t.Fatal(unmarshalPostErr)
	}

	if val == false {
		t.Logf("Success on test 6: success is %v because of no account", val)
	} else {
		t.Errorf("Failed on test 6: success is %v", val)
	}

	req.PostForm = url.Values{}
	req.PostForm.Add("login", "reil")
	req.PostForm.Add("password", "12345678")

	res, _ = client.Do(req)

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.StatusCode != http.StatusOK {
		t.Errorf("Failed on test 7: status code is %v", res.StatusCode)
	} else {
		t.Logf("Success on test 7: status code is %v", res.StatusCode)
	}

	body, errRead = io.ReadAll(res.Body)

	if errRead != nil {
		t.Errorf("Failed on test 8: %s", errRead)
	} else {
		t.Logf("Success on test 8: body readed %s", string(body))
	}

	var outputPostLogin map[string]interface{}
	unmarshalPostLoginErr := json.Unmarshal(body, &outputPostLogin)

	if unmarshalPostLoginErr != nil {
		t.Fatal(unmarshalPostLoginErr)
	}

	if val == true {
		t.Logf("Success on test 9: success is %v", val)
	} else {
		t.Errorf("Failed on test 9: success is %v", val)
	}
}
