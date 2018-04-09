package main

import (
	"flag"
	"os"
	"fmt"
	"github.com/weishi258/user-access-control-proxy/log"
	"github.com/weishi258/user-access-control-proxy/db"
	"go.uber.org/zap"
	"github.com/weishi258/user-access-control-proxy/server"
	"os/signal"
	"syscall"
)
const VERSION = "0.1a"

var dbMgr *db.DBMgr
var sigChan chan os.Signal

func main(){

	// waiting loop for signal
	sigChan = make(chan os.Signal, 5)
	done := make(chan bool)

	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGKILL,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGINT)

	var err error

	var printVer bool
	var dbFilePath string
	var logLevel string
	var logMode bool

	flag.BoolVar(&printVer, "version", false, "print server version")
	flag.StringVar(&dbFilePath, "db", "main.db", "SQLite3 db file path")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.BoolVar(&logMode, "production", false, "Log output production mode")
	flag.Parse()

	defer func(){
		if err != nil{
			os.Exit(1)
		}else{
			os.Exit(0)
		}
	}()

	if printVer {
		fmt.Printf("UserAccessControlProxy Version %s", VERSION)
		return
	}

	logger := log.InitZapLogger(logLevel, logMode)
	if dbMgr, err = db.InitDB(dbFilePath); err != nil{
		logger.Fatal("Start database manager failed", zap.String("error", err.Error()))
		return
	}
	var adminServer *server.AdminServer
	if adminServer, err = server.NewAdminServer(dbMgr); err != nil{
		logger.Fatal(fmt.Sprintf("Create UserAccessControlProxy failed: %s", err.Error()))
		return
	}
	adminServer.Start(sigChan)
	defer adminServer.Shutdown()

	go func() {
		_ = <-sigChan
		//logger.Info("UserAccessControlProxy caught signal for exit",
		//	zap.Any("signal", sig))
		done <- true
	}()
	<-done
}
