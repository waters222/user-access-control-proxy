package server

import (
	"net/http"
	"go.uber.org/zap"
)

var RestUser = []Route{
	{
		"user_test",
		"get",
		"/user/rest/login",
		UserLogin,
	},

}



func UserLogin(w http.ResponseWriter, r *http.Request){
	logger := getLogger()

	var err error
	var code int
	var response []byte
	defer func(){
		if err != nil{
			ReturnError(w, code, err.Error())
		}else{
			w.WriteHeader(http.StatusOK)
			w.Write(response)
			logger.Info("Test ok")
		}
	}()

	userName := r.Header.Get(PROXY_USER_KEY)
	if len(userName) > 0{
		logger.Info("Test has user", zap.String("userName", userName))
	}else{
		logger.Info("Test has guest")
	}



}