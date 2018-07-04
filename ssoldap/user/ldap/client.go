package ldap

import (
	"github.com/mijia/sweb/log"
	"gopkg.in/ldap.v2"
	"strings"
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
	Ldap *ldap.Conn
	//Ldap *openldap.Ldap
}

func NewClient(url string, cn string, password string, base string) (*LdapClient, error){
	BASE = base
	ldapUrl = url
	cUSER = cn
	cPASSWORD = password
	addr := strings.Split(ldapUrl,"/")
	l, err := ldap.Dial("tcp", addr[2])
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	err = l.Bind(cn,password)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	log.Debug("init ldap client successfully")
	return &LdapClient{
		Ldap: l,
	}, nil
}


/*func oldNewClient(url string, cn string, password string, base string) (*LdapClient, error) {
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
	log.Debug("init ldap client successfully")
	return &LdapClient{
		Ldap: ldap,
	}, nil
}
*/

func  (c *LdapClient) ReConnect() {
	c.Close()
	log.Debug("reconnect")
	temp, err := NewClient(ldapUrl, cUSER, cPASSWORD, BASE)
	log.Debug(temp, err)
	c.Ldap = temp.Ldap
}

/*func (c *LdapClient) oldReConnect() {
	c.Close()
	log.Debug("reconnect")
	temp, err := NewClient(ldapUrl, cUSER, cPASSWORD, BASE)
	log.Debug(temp, err)
	c.Ldap = temp.Ldap
}*/


func (c *LdapClient) Close() error {
	c.Ldap.Close()
	return nil
}
/*
func (c *LdapClient) oldClose() error {
	return c.Ldap.Close()
}
*/



func (c *LdapClient) search(filter string) (*ldap.SearchResult, error) {
	base := BASE
	scope := ldap.ScopeWholeSubtree
	attributes := []string{"cn", "userPrincipalName"}
	searchRequest := ldap.NewSearchRequest(
		base, // The base dn to search
		scope, ldap.NeverDerefAliases, 0, 0, false,
		filter, // The filter to apply
		attributes,                    // A list attributes to retrieve
		nil,
	)
	sr, err := c.Ldap.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return sr, nil
}


/*func (c *LdapClient) oldsearch(filter string) (*openldap.LdapSearchResult, error) {
	// LDAP_SCOPE_BASE, LDAP_SCOPE_ONELEVEL, LDAP_SCOPE_SUBTREE
	scope := openldap.LDAP_SCOPE_SUBTREE
	attributes := []string{"cn", "userPrincipalName"}
	return c.Ldap.SearchAll(BASE, scope, filter, attributes)
}
*/



func (c *LdapClient) SearchForUser(filter string) (*ldap.SearchResult, error) {
	log.Debug("begin to call ldap lib")
	base := BASE
	scope := ldap.ScopeWholeSubtree
	attributes := []string{"cn", "whenCreated", "employeeID", "whenChanged", "userPrincipalName", "mail", "mailNickname"}
	searchRequest := ldap.NewSearchRequest(
		base, // The base dn to search
		scope, ldap.NeverDerefAliases, 0, 0, false,
		filter, // The filter to apply
		attributes,                    // A list attributes to retrieve
		nil,
	)
	result, err := c.Ldap.Search(searchRequest)
	log.Debug(result)
	log.Debug("end to call ldap lib")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(result.Entries) > 1 {
		return result, ErrUserUncertain
	} else if len(result.Entries) == 0 {
		return nil, ErrUserNotFound
	} else {
		return result, err
	}
}


/*
func (c *LdapClient) oldSearchForUser(filter string) (*openldap.LdapSearchResult, error) {
	// LDAP_SCOPE_BASE, LDAP_SCOPE_ONELEVEL, LDAP_SCOPE_SUBTREE
	scope := openldap.LDAP_SCOPE_SUBTREE
	attributes := []string{"*"}
	//	attributes := []string{"cn", "whenCreated", "employeeID", "whenChanged",
	//		"userPrincipalName", "mail", "mailNickname"}
	log.Debug("begin to call ldap lib")
	result, err := c.Ldap.SearchAll(BASE, scope, filter, attributes)
	log.Debug(result)
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
}*/

func (c *LdapClient) SearchForOU(OUs string) (*ldap.SearchResult, error) {
	base := OUs
	scope := ldap.ScopeBaseObject
	attributes := []string{"name", "ou"}
	filter := "(&(objectClass=organizationalUnit)(objectClass=top))"
	searchRequest := ldap.NewSearchRequest(
		base, // The base dn to search
		scope, ldap.NeverDerefAliases, 0, 0, false,
		filter, // The filter to apply
		attributes,                    // A list attributes to retrieve
		nil,
	)
	return c.Ldap.Search(searchRequest)
}

/*func (c *LdapClient) oldSearchForOU(OUs string) (*openldap.LdapSearchResult, error) {
	base := OUs
	scope := openldap.LDAP_SCOPE_BASE
	attributes := []string{"name", "ou"}
	filter := "(&(objectClass=organizationalUnit)(objectClass=top))"
	return c.Ldap.SearchAll(base, scope, filter, attributes)
}
*/