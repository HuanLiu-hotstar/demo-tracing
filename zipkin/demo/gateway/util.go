package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Req struct {
	ID string `json:"id"`
}
type Resp struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

func getbody(r *http.Request) (*Req, error) {
	req := &Req{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return req, err
	}
	defer r.Body.Close()
	if err := json.Unmarshal(body, req); err != nil {
		return req, err
	}
	return req, nil

}
func Write(w http.ResponseWriter, code int32, msg string) {
	resp := Resp{
		Code: code,
		Msg:  msg,
	}
	bye, _ := json.Marshal(resp)
	w.Write(bye)
}
