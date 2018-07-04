package ldap

import (
	"errors"

	"github.com/mijia/sweb/log"
	"gopkg.in/ldap.v2"
	"strings"
)

var (
	ErrUserNotFound  = errors.New("USER NOT FOUND")
	ErrUserUncertain = errors.New("USER UNCERTAIN")
)

func (c *LdapClient) Auth(feature, passwd string) (bool, error) {
	return c.AuthEmail(feature, passwd)
}


func AuthUPN(mail string, passwd string) (bool, error) {
	addr := strings.Split(ldapUrl,"/")
	ldap, err := ldap.Dial("tcp", addr[2])
	if err != nil {
		log.Fatal(err)
		return false, err
	}
	err = ldap.Bind(mail, passwd)
	if err != nil {
		log.Debug(err)
		return false, nil
	} else {
		// Close will panic if bind fails
		ldap.Close()
		log.Debug("success")
		return true, nil
	}
}


/*func AuthUPN(mail string, passwd string) (bool, error) {
	ldap, err := openldap.Initialize(ldapUrl)
	if err != nil {
		return false, err
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(mail, passwd)
	if err != nil {
		log.Debug(err)
		return false, nil
	} else {
		// Close will panic if bind fails
		ldap.Close()
		log.Debug("success")
		return true, nil
	}
}
*/


func (c *LdapClient) GetUserPrincipalName(filter string) (string, error) {
	log.Debug("Get cn:", filter)
	result, err := c.search(filter)
	if err != nil {
		log.Debug(err)
		return "", err
	}
	if len(result.Entries) == 0 {
		log.Debug(ErrUserNotFound)
		return "", ErrUserNotFound
	} else if len(result.Entries) > 1 {
		log.Debug(err)
		return "", ErrUserUncertain
	}
	for _, entry := range result.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "userPrincipalName" {
				log.Debug(attr.Values)
				return attr.Values[0], nil
			}
		}
	}
	panic("should already return")
	return "", nil
}


/*func (c *LdapClient) GetUserPrincipalName(filter string) (string, error) {
	log.Debug("Get cn:", filter)
	result, err := c.search(filter)
	if err != nil {
		log.Debug(err)
		return "", err
	}
	if len(result.Entries) == 0 {
		log.Debug(ErrUserNotFound)
		return "", ErrUserNotFound
	} else if len(result.Entries) > 1 {
		log.Debug(err)
		return "", ErrUserUncertain
	}
	for _, entry := range result.Entries {
		for _, attr := range entry.Attributes {
			if attr.Name == "userPrincipalName" {
				log.Debug(attr.Values)
				return attr.Values[0], nil
			}
		}
	}
	panic("should already return")
	return "", nil
}
*/

func (c *LdapClient) AuthEmail(email, passwd string) (bool, error) {
	return AuthUPN(email, passwd)
}


/*func (c *LdapClient) AuthEmail(email, passwd string) (bool, error) {
	return AuthUPN(email, passwd)
}
*/

func (c *LdapClient) AuthFilter(filter, passwd string) (bool, error) {
	upn, err := c.GetUserPrincipalName(filter)
	log.Debug(upn, " ", err)
	if err != nil {
		log.Debug("return false")
		return false, err
	}
	return AuthUPN(upn, passwd)
}

/*func (c *LdapClient) AuthFilter(filter, passwd string) (bool, error) {
	upn, err := c.GetUserPrincipalName(filter)
	log.Debug(upn, " ", err)
	if err != nil {
		log.Debug("return false")
		return false, err
	}
	return AuthUPN(upn, passwd)
}*/
