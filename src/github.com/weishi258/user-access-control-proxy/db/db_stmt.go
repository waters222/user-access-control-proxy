package db

const(
	STMT_INIT_USERS = iota
	STMT_INIT_GROUPS
	STMT_INIT_SESSIONS
	STMT_INIT_RULES
	STMT_INIT_GROUPS_RULES

	STMT_INSERT_USERS
	STMT_INSERT_GROUP
	STMT_INSERT_RULES
	STMT_INSERT_GROUPS_RULES
	STMT_INSERT_SESSIONS

)

var StmtMap = map[int]string{
	STMT_INIT_USERS = "create table users(id integer not null primary key, name text, password char(40), salt char(32), group_id integer not null);",
	STMT_INIT_GROUPS = "create table groups(id integer not null primary key, name text, desc text);",
	STMT_INIT_SESSIONS = "create table sessions(sess text nul null primary key, user_id integer not null, last_time integer not null);",
	STMT_INIT_RULES = "create table rules(id integer not null primary key, rule text not null, desc text);",
	STMT_INIT_GROUPS_RULES = "create table groups_rules(group_id integer not null primary key, rule_id integer not null,  permission integer not null);",

	STMT_INSERT_GROUP = "insert into groups(name, desc) values(?, ?);",
	STMT_INSERT_USERS = "insert into users(name, password, salt, group_id) values(?, ?, ?, ?);",
	STMT_INSERT_RULES = "insert into rules(rule, desc) values(?, ?,);"
	STMT_INSERT_GROUPS_RULES = "insert into groups_rules(group_id, rule_id, permission) values(?, ?, ?);",
	STMT_INSERT_SESSIONS = "insert into session(sess, user_id, last_time) values(?, ?, ?);",


}
