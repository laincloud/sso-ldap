package user

import (
	"strconv"
	"strings"
	"time"

	"github.com/mijia/sweb/log"
	//	"github.com/mqu/openldap"

	"github.com/laincloud/sso-ldap/ssoldap/user/ldap"
	"github.com/laincloud/sso/ssolib/models/iuser"
)

var baseTime = time.Date(2000, 1, 1, 12, 1, 1, 0, time.UTC)

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
			case "employeeID":
				ret.Id = getIdByEId(v)
				//			case "mail":
				//				ret.Email = v
			case "userPrincipalName":
				ret.Email = v
			case "mailNickname":
				ret.Name = v
			case "whenCreated":
				ret.Created = v
			case "whenChanged":
				ret.Updated = v
			}
		}
	}
	log.Debug("end search ldap")
	return ret, nil
}

func getIdByEId(employeeID string) int {
	// 为了节省内存，我们暂且认为员工编号是 20********** 共12位数字
	// 考虑到 int 的范围 是 2147483647 ~ -2147483648
	if len(employeeID) != 12 {
		panic("unexpected employeeID")
	}
	if !strings.HasPrefix(employeeID, "20") {
		log.Error("unexpected employeeID:" + employeeID)
	}
	// 将员工的前八位映射成当前日期距离 20000101 的天数
	year, err := strconv.Atoi(employeeID[0:4])
	if err != nil {
		panic(err)
	}
	month, err := strconv.Atoi(employeeID[4:6])
	if err != nil {
		panic(err)
	}
	day, err := strconv.Atoi(employeeID[6:8])
	if err != nil {
		panic(err)
	}
	idTime := time.Date(year, time.Month(month), day, 12, 1, 1, 0, time.UTC)
	idUTime := idTime.Unix()
	baseUTime := baseTime.Unix()
	between := idUTime - baseUTime
	days := int(between / 3600 / 24)
	sDays := strconv.Itoa(days)
	if len(sDays) > 8 {
		log.Error(year, " ", month, " ", day, " ", idUTime, " ", baseUTime)
		panic("impossible")
	}
	last, err := strconv.Atoi(employeeID[8:12])
	if err != nil {
		log.Error(employeeID)
		panic(err)
	}
	id := days*10000 + last
	return id
}

func getEIdById(id int) string {
	last := strconv.Itoa((id % 10000) + 10000)
	between := id / 10000 * 3600 * 24
	baseUTime := baseTime.Unix()
	idUTime := baseUTime + int64(between)
	idTime := time.Unix(idUTime, 0)
	first := idTime.Format("20060102")
	return first + last[1:]
}
