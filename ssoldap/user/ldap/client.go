package ldap

import (
	"strings"

	"github.com/mijia/sweb/log"
	"github.com/mqu/openldap"
)

var (
	BASE string = ""
)

var (
	// for the first ldap search
	cUSER     string
	cPASSWORD string

	// for authenticating the user
	ldapUrl string
)

type LdapClient struct {
	Ldap *openldap.Ldap
}

func NewClient(url string, cn string, password string, base string) (*LdapClient, error) {
	BASE = base
	ldap, err := openldap.Initialize(url)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	ldapUrl = url
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	cUSER = cn
	cPASSWORD = password
	err = ldap.Bind(cn, password)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return &LdapClient{
		Ldap: ldap,
	}, nil
}

func (c *LdapClient) ReConnect() {
	c.Close()
	temp, _ := NewClient(ldapUrl, cUSER, cPASSWORD, BASE)
	c.Ldap = temp.Ldap
}

func (c *LdapClient) Close() error {
	return c.Ldap.Close()
}

func (c *LdapClient) search(filter string) (*openldap.LdapSearchResult, error) {
	// LDAP_SCOPE_BASE, LDAP_SCOPE_ONELEVEL, LDAP_SCOPE_SUBTREE
	scope := openldap.LDAP_SCOPE_SUBTREE
	attributes := []string{"cn", "userPrincipalName"}
	return c.Ldap.SearchAll(BASE, scope, filter, attributes)
}

func (c *LdapClient) SearchForUser(filter string) (*openldap.LdapSearchResult, error) {
	// LDAP_SCOPE_BASE, LDAP_SCOPE_ONELEVEL, LDAP_SCOPE_SUBTREE
	scope := openldap.LDAP_SCOPE_SUBTREE
	attributes := []string{"cn", "whenCreated", "employeeID", "whenChanged",
		"userPrincipalName", "mail", "mailNickname"}
	log.Debug("begin to call ldap lib")
	result, err := c.Ldap.SearchAll(BASE, scope, filter, attributes)
	log.Debug("end to call ldap lib")
	if err != nil {
		log.Error(err)
		//LDAP::Search() error : -1 (Can't contact LDAP server)
		if strings.Index(err.Error(), "LDAP::Search() error : -1 (Can't contact LDAP server)") >= 0 {
			c.ReConnect()
			result, err = c.Ldap.SearchAll(BASE, scope, filter, attributes)
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
