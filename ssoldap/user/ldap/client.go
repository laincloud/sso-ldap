package ldap

import (
	"errors"
	"strings"

	"github.com/mijia/sweb/log"
	"github.com/mqu/openldap"
)

var (
	ErrUserNotFound  = errors.New("USER NOT FOUND")
	ErrUserUncertain = errors.New("USER UNCERTAIN")
)

type LdapClient struct {
	baseDn   string
	ldapUrl  string
	user     string
	password string
	Ldap     *openldap.Ldap
}

func NewClient(url string, cn string, password string, base string) (*LdapClient, error) {
	client := &LdapClient{
		baseDn:   base,
		ldapUrl:  url,
		user:     cn,
		password: password,
	}

	ldap, err := client.newConn()
	if err != nil {
		log.Errorf("failed to init ldap conn, err %+v", err)
		return nil, err
	}
	log.Info("init ldap client successfully")
	client.Ldap = ldap
	return client, nil
}

func (c *LdapClient) newConn() (*openldap.Ldap, error) {
	ldap, err := openldap.Initialize(c.ldapUrl)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(c.user, c.password)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return ldap, nil
}

func (c *LdapClient) ReConnect() {
	c.Close()
	conn, err := c.newConn()
	if err != nil {
		log.Errorf("failed to reconnect ldap, err %+v", err)
		return
	}
	c.Ldap = conn
}

func (c *LdapClient) Close() error {
	if c.Ldap != nil {
		return c.Ldap.Close()
	}
	return nil
}

func (c *LdapClient) SearchForUser(filter string) (*openldap.LdapSearchResult, error) {
	// LDAP_SCOPE_BASE, LDAP_SCOPE_ONELEVEL, LDAP_SCOPE_SUBTREE
	scope := openldap.LDAP_SCOPE_SUBTREE
	attributes := []string{"*"}
	//	attributes := []string{"cn", "whenCreated", "employeeID", "whenChanged",
	//		"userPrincipalName", "mail", "mailNickname"}
	log.Debug("begin to call ldap lib")
	result, err := c.Ldap.SearchAll(c.baseDn, scope, filter, attributes)
	log.Debug(result)
	log.Debug("end to call ldap lib")
	if err != nil {
		log.Error(err)
		//LDAP::Search() error : -1 (Can't contact LDAP server)
		if strings.Index(err.Error(), "LDAP::Search() error : -1 (Can't contact LDAP server)") >= 0 {
			c.ReConnect()
			result, err = c.Ldap.SearchAll(c.baseDn, scope, filter, attributes)
		} else {
			return nil, err
		}
	}
	if result.Count() > 1 {
		return result, ErrUserUncertain
	} else if result.Count() == 0 {
		return nil, ErrUserNotFound
	} else {
		return result, err
	}
}

func (c *LdapClient) SearchForOU(OUs string) (*openldap.LdapSearchResult, error) {
	base := OUs
	scope := openldap.LDAP_SCOPE_BASE
	attributes := []string{"name", "ou"}
	filter := "(&(objectClass=organizationalUnit)(objectClass=top))"
	return c.Ldap.SearchAll(base, scope, filter, attributes)
}

func (c *LdapClient) Auth(mail string, passwd string) (bool, error) {
	ldap, err := openldap.Initialize(c.ldapUrl)
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
