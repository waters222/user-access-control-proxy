package server

import (
	"net/http"
	"encoding/json"
)

type RestError struct {
	Error 		string
}

func ReturnError(w http.ResponseWriter, code int, err string) error{
	if response, err := json.Marshal(RestError{err}); err != nil {
		w.WriteHeader(code)
		w.Write(nil)
		return err
	} else {
		w.WriteHeader(code)
		w.Write(response)
		return nil
	}
}
