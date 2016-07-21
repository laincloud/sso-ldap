package user

import (
	_ "fmt"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso/ssolib/models/iuser"
)

type User struct {
	Id       int
	Name     string
	FullName string
	Email    string
	Mobile   string
	Created  string
	Updated  string

	dn      string
	backend iuser.UserBackend
}

type UserProfile struct {
	Name     string `json:"name"`
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
}

func (up *UserProfile) GetName() string {
	return up.Name
}

func (up *UserProfile) GetEmail() string {
	return up.Email
}

func (up *UserProfile) GetMobile() string {
	return up.Mobile
}

func (u *User) GetName() string {
	return u.Name
}

func (u *User) GetFullName() string {
	return u.FullName
}

func (u *User) GetId() int {
	return u.Id
}

func (u *User) GetEmail() string {
	return u.Email
}

func (u *User) GetMobile() string {
	log.Debug(u.Mobile)
	return u.Mobile
}

func (u *User) SetBackend(b iuser.UserBackend) {
	u.backend = b
}

func (u *User) GetSub() string {
	return u.backend.(*UserBack).UserIdToSub(u.Id)
}

func (u *User) GetProfile() iuser.UserProfile {
	return &UserProfile{
		Name:     u.GetName(),
		FullName: u.GetFullName(),
		Email:    u.GetEmail(),
		Mobile:   u.GetMobile(),
	}
}

func (u *User) GetPublicProfile() iuser.UserProfile {
	return &UserProfile{
		Name:     u.GetName(),
		FullName: u.GetFullName(),
		Email:    u.GetEmail(),
	}
}
