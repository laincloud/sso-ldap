package ldap

import (
	"errors"

	"github.com/mijia/sweb/log"
	"github.com/mqu/openldap"
)

var (
	ErrUserNotFound  = errors.New("USER NOT FOUND")
	ErrUserUncertain = errors.New("USER UNCERTAIN")
)

func (c *LdapClient) Auth(feature, passwd string) (bool, error) {
	return c.AuthEmail(feature, passwd)
}

func AuthUPN(mail string, passwd string) (bool, error) {
	ldap, err := openldap.Initialize(ldapUrl)
	if err != nil {
		return false, err
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(mail, passwd)
	if err != nil {
		return false, nil
	} else {
		// Close will panic if bind fails
		ldap.Close()
		log.Debug("success")
		return true, nil
	}
}

func (c *LdapClient) GetUserPrincipalName(filter string) (string, error) {
	log.Debug("Get cn:", filter)
	result, err := c.search(filter)
	if err != nil {
		log.Debug(err)
		return "", err
	}
	if result.Count() == 0 {
		log.Debug(ErrUserNotFound)
		return "", ErrUserNotFound
	} else if result.Count() > 1 {
		log.Debug(err)
		return "", ErrUserUncertain
	}
	for _, entry := range result.Entries() {
		for _, attr := range entry.Attributes() {
			if attr.Name() == "userPrincipalName" {
				log.Debug(attr.Values())
				return attr.Values()[0], nil
			}
		}
	}
	panic("should already return")
	return "", nil
}

func (c *LdapClient) AuthEmail(email, passwd string) (bool, error) {
	return AuthUPN(email, passwd)
}

func (c *LdapClient) AuthFilter(filter, passwd string) (bool, error) {
	upn, err := c.GetUserPrincipalName(filter)
	log.Debug(upn, " ", err)
	if err != nil {
		log.Debug("return false")
		return false, err
	}
	return AuthUPN(upn, passwd)
}
