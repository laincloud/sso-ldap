package user

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/mijia/sweb/log"
	"golang.org/x/net/context"

	"github.com/laincloud/sso-ldap/ssoldap/user/ldap"
	"github.com/laincloud/sso/ssolib/models/iuser"
	"github.com/laincloud/sso/ssolib/utils"
)

// the feature in this file is either the email(also as upn), or the email prefix
// if the user is in the ldap backend, only the email and id are valid in the user table; in this code as a example, the mobile is also valid, but the name should be the prefix of the UPN(email) by default, so not consider the old data.
// if the user is not in the ldap, all the fields in the user table is valid

const (
	LDAPREALMNAME = "sso-ldap"
)

// for compatible with sso database
const createUserTableSQL = `
CREATE TABLE IF NOT EXISTS user (
	id INT NOT NULL AUTO_INCREMENT,
	name VARCHAR(32) NULL DEFAULT NULL,
	fullname VARCHAR(128) CHARACTER SET utf8 NULL DEFAULT NULL,
	email VARCHAR(64) NULL DEFAULT NULL,
	password VARBINARY(60) NULL DEFAULT NULL,
	mobile VARCHAR(11) NULL DEFAULT NULL,
	created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY (id),
	UNIQUE KEY (name),
	UNIQUE KEY (email)
) DEFAULT CHARSET=latin1
`

var EMAIL_SUFFIX string

var (
	ErrForbidden = errors.New("functions not developed")
)

type UserBack struct {
	C  *ldap.LdapClient
	DB *sqlx.DB
}

func New(url, cn, passwd string, mysqlDSN string, email string, ldapBase string) *UserBack {
	client, err := ldap.NewClient(url, cn, passwd, ldapBase)
	log.Debug(url, " ", cn, " ", passwd)
	if err != nil {
		panic(err)
	}
	db, err := utils.InitMysql(mysqlDSN)
	if err != nil {
		panic(err)
	}
	EMAIL_SUFFIX = email
	return &UserBack{
		C:  client,
		DB: db,
	}
}

func (ub *UserBack) InitDatabase() {
	tx := ub.DB.MustBegin()
	tx.MustExec(createLdapGroupTableSQL)
	tx.MustExec(createUserTableSQL)
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func (ub *UserBack) ListUsers(ctx context.Context) ([]iuser.User, error) {
	// 加一个参数, 来防止太长的返回值

	// FIXME 临时解决方案
	db := ctx.Value("db").(*sqlx.DB)
	userIds := []int{}
	//	err := db.Select(&userIds, "SELECT DISTINCT user_id FROM user_group")
	err := db.Select(&userIds, "SELECT id FROM user")
	ret := make([]iuser.User, len(userIds))

	log.Debug(userIds)
	for i, v := range userIds {
		log.Debug(i, v)
		ret[i], err = ub.GetUser(int(v))
		log.Debug(err)
	}
	return ret, nil
}

func (ub *UserBack) GetUser(id int) (iuser.User, error) {
	upn, err := ub.getUPNById(id)
	if err != nil {
		return nil, err
	}
	user, err := ub.Search("userPrincipalName=" + upn)
	if err != nil {
		if err != iuser.ErrUserNotFound {
			return user, err
		} else {
			user, err = ub.getUserFromMysql(id)
			if err != nil {
				log.Error(err)
				return nil, err
			}
		}
	}
	user.SetBackend(ub)

	return user, nil
}

func (ub *UserBack) GetUserByName(name string) (iuser.User, error) {
	return ub.GetUserByEmail(name + EMAIL_SUFFIX)
}

func (ub *UserBack) GetUserByEmail(email string) (iuser.User, error) {
	log.Debug(email)
	user, err := ub.Search("userPrincipalName=" + email)
	if err != nil {
		if err != iuser.ErrUserNotFound {
			return user, err
		} else {
			user, err = ub.getUserByEmailFromMysql(email)
			if err != nil {
				return nil, err
			}
		}
	}
	user.SetBackend(ub)
	return user, nil
}

func (ub *UserBack) CreateUser(user iuser.User, passwordHashed bool) error {
	return ErrForbidden
}

func (ub *UserBack) DeleteUser(user iuser.User) error {
	return ErrForbidden
}

// deprecated
func (ub *UserBack) AuthPassword(sub, passwd string) (bool, error) {
	log.Debug(sub)
	id, err := ub.UserSubToId(sub)
	if err != nil {
		log.Error(err)
		return false, err
	}
	u, err := ub.GetUser(id)
	log.Debug(id)
	if err != nil {
		log.Debug(err)
		return false, err
	}
	b, _ := json.Marshal(u)
	log.Debug(string(b))
	return ub.C.Auth(u.(*User).Email, passwd)
}

func (ub *UserBack) AuthPasswordByFeature(feature, passwd string) (bool, iuser.User, error) {
	if !strings.HasSuffix(feature, EMAIL_SUFFIX) {
		feature = feature + EMAIL_SUFFIX
	}
	success, err := ub.C.Auth(feature, passwd)
	log.Debug(err)
	if success {
		u, err := ub.GetUserByEmail(feature)
		return true, u, err
	} else {
		u, err := ub.getUserByEmailFromMysql(feature)
		if err == nil && u != nil {
			if u.VerifyPassword([]byte(passwd)) {
				return true, u, err
			}
		}
	}
	return false, nil, nil
}

func (ub *UserBack) GetUserByFeature(f string) (iuser.User, error) {
	if strings.HasSuffix(f, EMAIL_SUFFIX) {
		return ub.GetUserByEmail(f)
	} else {
		return ub.GetUserByEmail(f + EMAIL_SUFFIX)
	}
}

func (ub *UserBack) Name() string {
	return LDAPREALMNAME
}

func (ub *UserBack) SupportedVerificationMethods() []string {
	ret := []string{}
	ret = append(ret, iuser.PASSWORD)
	return ret
}

func (ub *UserBack) UserIdToSub(id int) string {
	return LDAPREALMNAME + fmt.Sprint(id)
}

func (ub *UserBack) UserSubToId(sub string) (int, error) {
	if !strings.HasPrefix(sub, LDAPREALMNAME) {
		return -1, iuser.ErrInvalidSub
	} else {
		return strconv.Atoi(sub[len(LDAPREALMNAME):])
	}
}

func (ub *UserBack) getUserFromMysql(id int) (*User, error) {
	user := User{}
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE id=?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (ub *UserBack) getUserByEmailFromMysql(email string) (*User, error) {
	user := User{}
	err := ub.DB.Get(&user, "SELECT * FROM user WHERE email=?", email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, iuser.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil

}
