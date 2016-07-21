package user

import (
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

const (
	LDAPREALMNAME = "sso-ldap"
)

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
	if err := tx.Commit(); err != nil {
		panic(err)
	}
}

func (ub *UserBack) ListUsers(ctx context.Context) ([]iuser.User, error) {
	// 加一个参数, 来防止太长的返回值

	// FIXME 临时解决方案
	db := ctx.Value("db").(*sqlx.DB)
	userIds := []int{}
	err := db.Select(&userIds, "SELECT DISTINCT user_id FROM user_group")
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
	employeeId := getEIdById(id)
	user, err := ub.Search("sAMAccountName=" + employeeId)
	if err != nil {
		return user, err
	}
	user.SetBackend(ub)

	return user, nil
}

func (ub *UserBack) GetUserByName(name string) (iuser.User, error) {
	return ub.GetUserByEmail(name + EMAIL_SUFFIX)
}

func (ub *UserBack) GetUserByEmail(email string) (iuser.User, error) {
	user, err := ub.Search("userPrincipalName=" + email)
	if err != nil {
		if err == ldap.ErrUserNotFound {
			return user, iuser.ErrUserNotFound
		}
		return user, err
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
	return false, nil, nil
}

func (ub *UserBack) GetUserByFeature(f string) (iuser.User, error) {
	return ub.GetUserByEmail(f)
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
