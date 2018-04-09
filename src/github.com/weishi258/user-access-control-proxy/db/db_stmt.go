package db
const (
	TABLE_USERS    = "users"
	TABLE_GROUPS   = "groups"
	TABLE_SESSIONS = "sessions"
	TABLE_RULES    = "rules"

)

const(
	STMT_DROP_USERS = iota
	STMT_DROP_GROUPS
	STMT_DROP_SESSIONS
	STMT_DROP_GROUPS_RULES

	STMT_INIT_USERS
	STMT_INIT_GROUPS
	STMT_INIT_SESSIONS
	STMT_INIT_RULES
	STMT_INIT_GROUPS_RULES


)
const (
	STMT_INSERT_USERS = iota
	STMT_INSERT_GROUP
	STMT_INSERT_RULES
	STMT_INSERT_SESSIONS


	STMT_SELECT_GROUP_BY_NAME
)


var StmtInitMap = map[int]string{
	STMT_DROP_USERS: "drop table "+TABLE_USERS+";",
	STMT_DROP_GROUPS : "drop table "+TABLE_GROUPS+";",
	STMT_DROP_SESSIONS : "drop table "+TABLE_SESSIONS+";",
	STMT_DROP_GROUPS_RULES : "drop table "+ TABLE_RULES +";",

	STMT_INIT_USERS : "create table "+TABLE_USERS+"(id integer not null primary key, name text, password char(40), salt char(32), group_id integer not null); " +
		"create unique index "+TABLE_USERS+"_name_index on "+TABLE_USERS+" (name);",
	STMT_INIT_GROUPS : "create table "+TABLE_GROUPS+"(id integer not null primary key, name text, desc text, mutable tinyint not null);"+
						"create unique index "+TABLE_GROUPS+"_name_index on "+TABLE_GROUPS+" (name);",
	STMT_INIT_SESSIONS : "create table "+TABLE_SESSIONS+"(sess char(40) not null primary key, user_id integer not null, last_time integer not null);",
	STMT_INIT_RULES : "create table "+ TABLE_RULES +"(id integer not null primary key, group_id integer, rule text not null,  permission tinyint not null, proxy text, weight integer not null, desc text);" +
		"create index "+TABLE_RULES+"_group_index on "+TABLE_RULES+" (group_id);",

}
var StmtMap = map[int]string{

	STMT_INSERT_GROUP :         "insert into "+TABLE_GROUPS+"(name, desc, mutable) values(?, ?, ?);",
	STMT_INSERT_USERS :         "insert into "+TABLE_USERS+"(name, password, salt, group_id) values(?, ?, ?, ?);",
	STMT_INSERT_RULES:          "insert into "+ TABLE_RULES +"(group_id, rule, permission, proxy, weight, desc) values(?, ?, ?, ?, ?, ?);",
	STMT_INSERT_SESSIONS :      "insert into "+TABLE_SESSIONS+"(sess, user_id, last_time) values(?, ?, ?);",
	STMT_SELECT_GROUP_BY_NAME : "select id, name, desc from "+TABLE_GROUPS+" where name = ?",
}
