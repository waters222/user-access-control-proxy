package db

import (
	"database/sql"
	. "github.com/weishi258/user-access-control-proxy/log"
	"go.uber.org/zap"
	"github.com/pkg/errors"
	"github.com/weishi258/user-access-control-proxy/util"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

const(
	ADMIN_NAME = "admin"
	ADMIN_PASSWROD = "admin"
	ADMIN_GROUP_NAME = "admin"
	SALT_LEN = 32
	PASSWORD_HASH_LEN = 40
	GUEST_GROUP_NAME = "guest"
	ADMIN_URL = "admin"
)


type DBMgr struct{
	db			*sql.DB
	stmt		[]*sql.Stmt
}
func getLogger() *zap.Logger{
	//return GetLogger().With(zap.String("Origin", "DBMgr"))
	return GetLogger()
}

func InitDB(dbFile string) (mgr *DBMgr, err error){
	logger := getLogger()

	bHasDB := false
	if _, err := os.Stat(dbFile); err != nil{
		if !os.IsNotExist(err){
			os.Remove(dbFile)
		}
	}else{
		bHasDB = true
	}
	mgr = &DBMgr{}
	if mgr.db, err = sql.Open("sqlite3", dbFile); err != nil{
		logger.Error("Init SQLite3 failed", zap.String("error", err.Error()))
		return nil, err
	}

	if !bHasDB{
		logger.Info("Database not initialized, starting populate database.")
		if err = mgr.bootstrapDB(); err != nil{
			logger.Error("Bootstrap database failed, remove db", zap.String("dbFile", dbFile))
			os.Remove(dbFile)
			return nil, errors.Wrap(err, "Bootstrap database failed")
		}
	}

	if err = mgr.prepareSQL(); err != nil{
		return nil, errors.Wrap(err, "Prepare stmt failed")
	}
	return mgr, nil
}
func (c* DBMgr)Close() error{
	if c.db != nil{
		if err := c.db.Close(); err != nil{
			return err
		}else{
			c.db = nil
			return nil
		}

	}else{
		return errors.New("DB handler is null")
	}

}
func (c* DBMgr)prepareSQL() (err  error){
	logger := getLogger()
	c.stmt = make([]*sql.Stmt, len(StmtMap))
	for key, sql := range StmtMap {
		if c.stmt[key], err = c.db.Prepare(sql); err != nil{
			err = errors.Wrapf(err, "Prepare statement %s failed", sql)
			return err
		}
		logger.Debug("Prepare SQL statement successful", zap.String("sql", sql))
	}
	return nil
}

func (c *DBMgr) bootstrapDB() (err error){
	if _, err = c.db.Exec(StmtInitMap[STMT_INIT_USERS]); err != nil{
		return errors.Wrap(err, "Create users table failed")
	}
	if _, err = c.db.Exec(StmtInitMap[STMT_INIT_GROUPS]); err != nil{
		return errors.Wrap(err, "Create groups table failed")
	}
	if _, err = c.db.Exec(StmtInitMap[STMT_INIT_SESSIONS]); err != nil{
		return errors.Wrap(err, "Create sessions table failed")
	}
	if _, err = c.db.Exec(StmtInitMap[STMT_INIT_RULES]); err != nil{
		return errors.Wrap(err, "Create rules table failed")
	}
	// lets create admin group and basic rule
	var tx *sql.Tx
	if tx, err = c.db.Begin(); err != nil{
		return errors.Wrap(err, "Can not begin transaction")
	}
	defer func(){
		if err != nil{
			tx.Rollback()
		}else{
			tx.Commit()
		}
	}()
	// create admin group
	var res sql.Result
	if res, err = tx.Exec(StmtMap[STMT_INSERT_GROUP], ADMIN_GROUP_NAME, "Admin Group"); err != nil{
		return errors.Wrap(err, "Can not create admin group")
	}
	var rowsAffected int64
	if rowsAffected, err = res.RowsAffected(); err != nil || rowsAffected == 0{
		err = errors.New("Can not create admin group because zero row affected")
		return err
	}
	// create first admin
	var adminGrpId int64
	if adminGrpId, err = res.LastInsertId(); err != nil{
		return errors.Wrap(err, "Get admin group id failed")
	}
	// create admin rules
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], adminGrpId, "^"+ADMIN_URL+"$", util.EncPermission(util.ProxyPermission{true, true, true, true}), "local", 1, "admin panel access rule"); err != nil{
		return errors.Wrap(err, "Can not insert admin's admin panel web access rule")
	}
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], adminGrpId, "^"+ADMIN_URL+"/.*$", util.EncPermission(util.ProxyPermission{true, true, true, true}), "local", 2, "admin panel access rule"); err != nil{
		return errors.Wrap(err, "Can not insert admin's admin panel full access rule")
	}
	// generate salt
	salt := util.RandomCharString(SALT_LEN)
	passwordHashed := util.EncodeToSha1String(ADMIN_PASSWROD + salt)
	if res, err = tx.Exec(StmtMap[STMT_INSERT_USERS],  "admin", passwordHashed, salt, adminGrpId); err != nil{
		return errors.Wrap(err, "Can not create admin user")
	}

	if rowsAffected, err = res.RowsAffected(); err != nil || rowsAffected == 0{
		err = errors.New("Can not create admin user because zero row affected")
		return err
	}

	// create guest group
	if res, err = tx.Exec(StmtMap[STMT_INSERT_GROUP], GUEST_GROUP_NAME, "Guest Group"); err != nil{
		return errors.Wrap(err, "Can not create guest group")
	}
	var guestGrpId int64
	if guestGrpId, err = res.LastInsertId(); err != nil{
		return errors.Wrap(err, "Get guest group id failed")
	}
	// create guest rules
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], guestGrpId, "^"+ADMIN_URL+"$", util.EncPermission(util.ProxyPermission{true, false, false, false}), "local", 1, "admin panel html access rule"); err != nil{
		return errors.Wrap(err, "Can not insert guest's admin panel web access rule")
	}
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], guestGrpId, "^"+ADMIN_URL+"/statics/.+\\.(json|png|jpg|svg|css)$", util.EncPermission(util.ProxyPermission{true, false, false, false}), "local", 2, "admin panel statics access rule"); err != nil{
		return errors.Wrap(err, "Can not insert guest's admin panel statics access rule")
	}
	return nil
}