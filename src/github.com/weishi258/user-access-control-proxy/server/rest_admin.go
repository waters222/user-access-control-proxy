package server

import (
	"net/http"
	"go.uber.org/zap"
)

var RestAdmin = []Route{
	{
		"admin_test",
		"get",
		"/admin/rest/test",
		AdminTest,
	},

}



func AdminTest(w http.ResponseWriter, r *http.Request){
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
	logger.Info("Test has admin", zap.String("userName", userName))


}