package user

import (
	"database/sql"
	"strings"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso-ldap/ssoldap/user/ldap"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

func (ub *UserBack) Search(filter string) (*User, error) {
	ret := &User{}

	log.Debug("begin search ldap")
	result, err := ub.C.SearchForUser(filter)
	log.Debug("end with results")
	if err != nil {
		log.Debug(err)
		if err == ldap.ErrUserNotFound {
			err = iuser.ErrUserNotFound
		}
		return ret, err
	}

	for _, entry := range result.Entries() {
		ret.dn = entry.Dn()
		for _, attr := range entry.Attributes() {
			//			log.Debug(attr.Name())
			v := attr.Values()[0]
			//			log.Debug(v)
			switch attr.Name() {
			case "cn":
				ret.FullName = v
			case "userPrincipalName":
				ret.Email = v
				ret.Name = getUserNameByUPN(v)
				ret.Id, err = ub.getIdByUPN(v)
			case "whenCreated":
				ret.Created = v
			case "whenChanged":
				ret.Updated = v
			}
		}
	}
	log.Debug("end search ldap")
	ret.Mobile = ub.GetMobileByEmail(ret.Email)
	return ret, nil
}

// the UPN is email, and the name is the prefix of the email
// if the usage of ldap is diffrent, you must fix
func getUserNameByUPN(upn string) string {
	atIndex := strings.Index(upn, "@")
	return upn[0:atIndex]
}

func (ub *UserBack) GetMobileByEmail(email string) string {
	// for example, just use the mysql backend
	var mobile string
	err := ub.DB.Get(&mobile, "SELECT mobile FROM user WHERE email=?", email)
	if err != nil {
		return ""
	}
	return mobile
}

// if the upn is not in mysql, insert and return the id
func (ub *UserBack) getIdByUPN(upn string) (int, error) {
	item := User{}
	tx := ub.DB.MustBegin()
	err := tx.Get(&item, "SELECT * FROM user WHERE email=?", upn)
	if err == sql.ErrNoRows {
		result, err1 := tx.Exec("INSERT INTO user (email) "+"VALUES(?)", upn)
		if err2 := tx.Commit(); err2 != nil {
			log.Error(err2)
			return -1, err2
		}
		if err1 != nil {
			log.Error(err1)
			return -1, err1
		}
		if id, err := result.LastInsertId(); err != nil {
			log.Error(err)
			return -1, err
		} else {
			return int(id), nil
		}
	} else if err != nil {
		log.Error(err)
		return -1, err
	} else {
		return item.Id, nil
	}
}

func (ub *UserBack) getUPNById(id int) (string, error) {
	item := User{}
	err := ub.DB.Get(&item, "SELECT * FROM user WHERE id=?", id)
	return item.Email, err
}
