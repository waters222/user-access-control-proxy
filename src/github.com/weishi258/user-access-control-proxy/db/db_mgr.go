package db

import (
	"database/sql"
	. "github.com/weishi258/user-access-control-proxy/log"
	"go.uber.org/zap"
	"github.com/pkg/errors"
	"github.com/weishi258/user-access-control-proxy/util"
)

const(
	ADMIN_NAME = "admin"
	ADMIN_PASSWROD = "admin"
	ADMIN_GROUP_NAME = "admin"
	SALT_LEN = 32
	PASSWORD_HASH_LEN = 40
)


type DBMgr struct{
	db	*sql.DB
}
func getLogger() *zap.Logger{
	return GetLogger().With(zap.String("pkg", "DBMgr"))
}

func InitDB(dbFile string) (mgr *DBMgr, err error){
	logger := getLogger()
	mgr = &DBMgr{}
	if mgr.db, err = sql.Open("sqlite3", dbFile); err != nil{
		logger.Error("Init SQLite3 failed", zap.String("error", err.Error()))
		return nil, err
	}
	return mgr, nil
}

func (c *DBMgr) bootstrapDB() (err error){
	if _, err = c.db.Exec(StmtMap[STMT_INIT_USERS]); err != nil{
		return errors.Wrap(err, "Create users table failed")
	}
	if _, err = c.db.Exec(StmtMap[STMT_INIT_GROUPS]); err != nil{
		return errors.Wrap(err, "Create groups table failed")
	}
	if _, err = c.db.Exec(StmtMap[STMT_INIT_SESSIONS]); err != nil{
		return errors.Wrap(err, "Create sessions table failed")
	}
	if _, err = c.db.Exec(StmtMap[STMT_INIT_RULES]); err != nil{
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

	// create admin rule
	return nil
}