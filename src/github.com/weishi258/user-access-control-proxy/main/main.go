package main

import (
	"flag"
	"os"
	"fmt"
	"github.com/weishi258/user-access-control-proxy/log"
	"github.com/weishi258/user-access-control-proxy/db"
	"go.uber.org/zap"
)
const VERSION = "0.1a"

func main(){

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
	//var dbMgr *db.DBMgr
	if _, err = db.InitDB(dbFilePath); err != nil{
		logger.Error("Start database manager failed", zap.String("error", err.Error()))
		return
	}

}