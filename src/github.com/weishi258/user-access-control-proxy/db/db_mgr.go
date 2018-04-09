package db

import (
	"database/sql"
	. "github.com/weishi258/user-access-control-proxy/log"
	"go.uber.org/zap"
	"github.com/pkg/errors"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"fmt"
	"sync"
	"github.com/weishi258/user-access-control-proxy/util"
)

const(
	ADMIN_NAME = "admin"
	ADMIN_PASSWROD = "admin"
	ADMIN_GROUP_NAME = "admin"
	SALT_LEN = 32
	PASSWORD_HASH_LEN = 40
	USER_GROUP_NAME = "user"
	USER_URL = "user"
	GUEST_GROUP_NAME = "guest"
	ADMIN_URL = "admin"
)


type DBMgr struct{
	db			*sql.DB
	stmt		[]*sql.Stmt

	// for fast cache
	groupRuleMux 		sync.Mutex
	groupRules			map[int][]Rule
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
	mgr.groupRules = make(map[int][]Rule)
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
		if err := c.Close(); err != nil{
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
	if res, err = tx.Exec(StmtMap[STMT_INSERT_GROUP], ADMIN_GROUP_NAME, "Admin Group", 0); err != nil{
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
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], adminGrpId, "^/"+ADMIN_URL+"$", EncPermission(ProxyPermission{true, true, true, true}), "", 1, "admin panel access rule"); err != nil{
		return errors.Wrap(err, "Can not insert admin's admin panel web access rule")
	}
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], adminGrpId, "^/"+ADMIN_URL+"/.*$", EncPermission(ProxyPermission{true, true, true, true}), "", 2, "admin panel access rule"); err != nil{
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

	// create users group
	if res, err = tx.Exec(StmtMap[STMT_INSERT_GROUP], USER_GROUP_NAME, "User Group", 0); err != nil{
		return errors.Wrap(err, "Can not create user group")
	}
	var userGrpId int64
	if userGrpId, err = res.LastInsertId(); err != nil{
		return errors.Wrap(err, "Get user group id failed")
	}

	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], userGrpId, "^/"+USER_URL+"/?.*$", EncPermission(ProxyPermission{true, true, true, true}), "", 1, "user panel html access rule"); err != nil{
		return errors.Wrap(err, "Can not insert user's user panel web access rule")
	}
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], userGrpId, "^/"+USER_URL+"/statics/.+\\.(json|png|jpg|svg|css)$", EncPermission(ProxyPermission{true, false, false, false}), "", 2, "user panel statics access rule"); err != nil{
		return errors.Wrap(err, "Can not insert user's user panel statics access rule")
	}

	// create guest group
	if res, err = tx.Exec(StmtMap[STMT_INSERT_GROUP], GUEST_GROUP_NAME, "Guest Group", 0); err != nil{
		return errors.Wrap(err, "Can not create guest group")
	}
	var guestGrpId int64
	if guestGrpId, err = res.LastInsertId(); err != nil{
		return errors.Wrap(err, "Get guest group id failed")
	}
	// create guest rules
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], guestGrpId, "^/"+USER_URL+"/?$", EncPermission(ProxyPermission{true, false, false, false}), "", 1, "user panel html access rule"); err != nil{
		return errors.Wrap(err, "Can not insert guest's user panel web access rule")
	}
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], guestGrpId, "^/"+USER_URL+"/rest/login$", EncPermission(ProxyPermission{true, true, false, false}), "", 1, "user panel html access rule"); err != nil{
		return errors.Wrap(err, "Can not insert guest's user panel web access rule")
	}
	if res, err = tx.Exec(StmtMap[STMT_INSERT_RULES], guestGrpId, "^/"+USER_URL+"/statics/.+\\.(json|png|jpg|svg|css)$", EncPermission(ProxyPermission{true, false, false, false}), "", 2, "user panel statics access rule"); err != nil{
		return errors.Wrap(err, "Can not insert guest's user panel statics access rule")
	}
	return nil
}

func (c *DBMgr)GetGroupByName(name string) (ret *Group, err error){
	logger := getLogger()
	var rows *sql.Rows
	if rows, err = c.db.Query(`select id, name, desc, mutable from `+TABLE_GROUPS+` where name = ?`, name); err != nil{
		logger.Error("Query group by name failed", zap.String("Error", err.Error()))
		return nil, errors.Wrap(err, "Get group by name failed")
	}
	defer rows.Close()
	if !rows.Next(){
		logger.Debug("Group is not exists", zap.String("name", name))
		return nil, errors.New(fmt.Sprintf("Group %s is not exists", name))
	}
	ret = &Group{}
	if err = rows.Scan(&ret.Id, &ret.Name, &ret.Desc, &ret.Mutable); err != nil{
		logger.Error("Scan table "+TABLE_GROUPS+" failed", zap.String("error", err.Error()))
		return nil, err
	}
	return ret, nil
}
func (c *DBMgr)GetGroupRules(groupId int) (rules []Rule, err error){
	logger := getLogger()
	c.groupRuleMux.Lock()
	defer c.groupRuleMux.Unlock()
	groupRules, _ := c.groupRules[groupId]
	if len(groupRules) == 0{
		var rows *sql.Rows
		if rows, err = c.db.Query(`select id, group_id, rule,  permission, weight from `+TABLE_RULES+` where group_id = ? order by weight asc`, groupId); err != nil{
			logger.Error("Query rules by group_id failed", zap.String("Error", err.Error()))
			return groupRules, err
		}
		defer rows.Close()
		for rows.Next(){
			rule := Rule{}
			if err = rows.Scan(&rule.Id, &rule.GroupId, &rule.Rule, &rule.Permission, &rule.Weight); err != nil{
				logger.Error("Scan group rule failed", zap.String("error", err.Error()))
				return groupRules, err
			}
			groupRules = append(groupRules, rule)
		}
		c.groupRules[groupId] = groupRules
	}
	rules = make([]Rule, len(groupRules))
	copy(rules, groupRules)
	return rules, nil
}