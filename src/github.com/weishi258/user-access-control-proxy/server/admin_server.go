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
	"github.com/weishi258/user-access-control-proxy/db"
	"github.com/pkg/errors"
)
const(
	SESSION_KEY = "sessionKey"
	SESSION_KEY_LEN = 40
	PROXY_USER_KEY = "proxy-user"
)
type AdminServer struct{
	server    		*http.Server
	localRouter		*mux.Router
	addr			string
	group			*db.Group
	dbMgr 			*db.DBMgr
}


type Route struct {
	Name        string
	Method      string
	Pattern     string
	Handler     http.HandlerFunc
}


func NewAdminServer(dbMgr *db.DBMgr) (ret *AdminServer, err error){
	logger := getLogger()
	ret = &AdminServer{}
	ret.dbMgr = dbMgr
	ret.addr = os.Getenv("LISTENING_ADDR")
	router := mux.NewRouter().StrictSlash(true)
	router.MatcherFunc(ret.MatchFunc).HandlerFunc(ret.HandlerFunc).Methods("OPTION", "GET", "POST", "PUT", "DELETE")
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})
	ret.server = &http.Server{Addr: ret.addr, Handler: handlers.CORS(headersOk, originsOk, methodsOk)(router)}
	ret.server.SetKeepAlivesEnabled(false)
	logger.Info(fmt.Sprintf("Allowed origin: %s", os.Getenv("ORIGIN_ALLOWED")))


	ret.localRouter = mux.NewRouter().StrictSlash(true)
	for _, route := range RestAdmin{
		ret.localRouter.Methods(route.Method,"OPTION").Path(route.Pattern).Name(route.Name).Handler(route.Handler)
	}
	for _, route := range RestUser{
		ret.localRouter.Methods(route.Method,"OPTION").Path(route.Pattern).Name(route.Name).Handler(route.Handler)
	}
	// get guest group id

	if ret.group, err = ret.dbMgr.GetGroupByName(db.GUEST_GROUP_NAME); err != nil{
		return nil, errors.Wrap(err, "Can not get guest group from database")
	}

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
	var err error
	var sessionKey string
	defer func(){
		if err != nil{
			http.ServeFile(w, r, fmt.Sprintf("html/errors/%d.html", http.StatusInternalServerError))
		}
	}()
	logger := getLogger()
	logger.Debug(fmt.Sprintf("HandlerFunc with uri: %s", r.RequestURI))
	// first remove any header for login info
	r.Header.Del(PROXY_USER_KEY)
	// lets check what kind of user it is
	var cookie *http.Cookie
	if cookie, err = r.Cookie(SESSION_KEY); err == nil{
		sessionKey = cookie.Value
	}


	// mock username
	userName := ""
	// cookie is empty so try to get it from header
	if len(sessionKey) != SESSION_KEY_LEN {
		sessionKey = r.Header.Get(SESSION_KEY)
	}
	groupId := c.group.Id
	if len(sessionKey) == SESSION_KEY_LEN{
		// get from database

	}
	var groupRules []db.Rule
	if groupRules, err = c.dbMgr.GetGroupRules(groupId); err != nil{
		logger.Error("Get group rules failed", zap.Int("group_id", groupId))
		return
	}
	for _, rule := range groupRules{
		if rule.Match(r.RequestURI) {
			permission := rule.GetPermission()
			bHasPermission := false
			switch r.Method{
			case "GET":
				bHasPermission = permission.Get
			case "PUT":
				bHasPermission = permission.Put
			case "POST":
				bHasPermission = permission.Post
			case "DELETE":
				bHasPermission = permission.Delete
			}
			if bHasPermission{
				if len(userName) > 0 {
					r.Header.Set(PROXY_USER_KEY, userName)
				}
				if rule.IsRemote(){
					c.handleProxy(w, r, rule.ComposeProxyUrl(r.RequestURI))
				}else{
					logger.Debug("server local for url", zap.String("url", rule.ComposeProxyUrl(r.RequestURI)))
					c.localRouter.ServeHTTP(w, r)
				}
				return
			}

		}
	}

	http.ServeFile(w, r, fmt.Sprintf("html/errors/%d.html", http.StatusForbidden))
}


func (c *AdminServer) handleProxy(w http.ResponseWriter, r *http.Request, url string){
	logger := getLogger()
	logger.Debug("handleProxy to remote", zap.String("url", url))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Remote OK"))
}