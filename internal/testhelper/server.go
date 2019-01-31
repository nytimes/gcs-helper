package testhelper

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type ServerTest struct {
	TestCase       string
	Method         string
	Addr           string
	ReqHeader      http.Header
	ExpectedStatus int
	ExpectedHeader http.Header
	ExpectedBody   interface{}
}

func (st ServerTest) Run(t *testing.T) {
	req, _ := http.NewRequest(st.Method, st.Addr, nil)
	for name := range st.ReqHeader {
		req.Header.Set(name, st.ReqHeader.Get(name))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != st.ExpectedStatus {
		t.Errorf("wrong status code\nwant %d\ngot  %d", st.ExpectedStatus, resp.StatusCode)
	}
	if expectedBody, ok := st.ExpectedBody.(string); ok {
		if string(data) != expectedBody {
			t.Errorf("wrong body\nwant %q\ngot  %q", expectedBody, string(data))
		}
	} else if st.ExpectedBody != nil {
		var body interface{}
		err = json.Unmarshal(data, &body)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, st.ExpectedBody) {
			t.Log(cmp.Diff(body, st.ExpectedBody))
			t.Errorf("wrong body returned\nwant %#v\ngot  %#v", st.ExpectedBody, body)
		}
	}
	for name := range st.ExpectedHeader {
		expected := st.ExpectedHeader.Get(name)
		got := resp.Header.Get(name)
		if expected != got {
			t.Errorf("header %q: wrong value\nwant %q\ngot  %q", name, expected, got)
		}
	}
}
