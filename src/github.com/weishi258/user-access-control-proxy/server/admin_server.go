package server

import (
	"net/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"os"
	"go.uber.org/zap"
	"syscall"
	"time"
	. "github.com/weishi258/user-access-control-proxy/log"
	"context"
	"fmt"
)

type AdminServer struct{
	server    	*http.Server
	addr		string
}

func NewAdminServer() (ret *AdminServer, err error){
	logger := getLogger()
	ret = &AdminServer{}
	ret.addr = os.Getenv("LISTENING_ADDR")
	router := mux.NewRouter().StrictSlash(true)
	router.MatcherFunc(ret.MatchFunc).HandlerFunc(ret.HandlerFunc).Methods("OPTION", "GET", "POST", "PUT", "DELETE")
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	ret.server = &http.Server{Addr: ret.addr, Handler: handlers.CORS(headersOk, originsOk, methodsOk)(router)}
	ret.server.SetKeepAlivesEnabled(false)

	logger.Info(fmt.Sprintf("Allowed origin: %s", os.Getenv("ORIGIN_ALLOWED")))
	return ret, nil
}

func getLogger() *zap.Logger{
	//return GetLogger().With(zap.String("Origin", "AdminServer"))
	return GetLogger()
}

func (c *AdminServer) Start(sigChan chan os.Signal) {
	go func() {
		logger := getLogger()
		logger.Info("UserAccessControlProxy Starting", zap.String("addr", c.addr))
		if err := c.server.ListenAndServe(); err != nil {
			logger.Info("UserAccessControlProxy stopped", zap.String("cause", err.Error()))
			sigChan <- syscall.SIGQUIT
		}
	}()
}

func (c *AdminServer) Shutdown() {
	logger := getLogger()
	logger.Info("UserAccessControlProxy is shutting down")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	c.server.Shutdown(ctx)
	logger.Info("UserAccessControlProxy is shutdown successful")
}

func (c *AdminServer) MatchFunc(r *http.Request, rm *mux.RouteMatch) bool{
	return true
}

func (c *AdminServer) HandlerFunc(w http.ResponseWriter, r *http.Request){
	logger := getLogger()
	logger.Debug(fmt.Sprintf("HandlerFunc with uri: %s", r.RequestURI))
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("hi there %s", r.RequestURI)))
}