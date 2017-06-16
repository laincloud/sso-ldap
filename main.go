package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"github.com/mijia/sweb/log"

	"github.com/laincloud/sso-ldap/ssoldap"
	"github.com/laincloud/sso-ldap/ssoldap/user"
	"github.com/laincloud/sso/ssolib"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {

	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	var webAddr, mysqlDSN, siteURL, smtpAddr, emailFrom, emailSuffix string
	var prikeyfile, pubkeyfile string
	var legalNets string
	var isDebug bool
	var sentryDSN string

	var ldapUrl, ldapUser, ldapPassword, ldapBase string

	flag.StringVar(&webAddr, "web", ":14000", "The address which SSO service is listening on")
	flag.StringVar(&mysqlDSN, "mysql", "user:password@tcp(127.0.0.1:3306)/dbname",
		"Data source name of mysql connection")
	flag.StringVar(&siteURL, "site", "http://sso.example.com", "Base URL of SSO site")

	// TODO fix the ssolib independent on smtp
	flag.StringVar(&smtpAddr, "smtp", "mail.example.com:25", "SMTP address for sending mail")
	flag.StringVar(&emailFrom, "from", "sso@example.com", "Email address to send register mail from")
	flag.StringVar(&emailSuffix, "domain", "@example.com", "Valid email suffix")
	flag.BoolVar(&isDebug, "debug", false, "Debug mode switch")
	flag.StringVar(&prikeyfile, "private", "certs/server.key", "private key file for jwt")
	flag.StringVar(&pubkeyfile, "public", "certs/server.pem", "public key file for jwt")
	flag.StringVar(&legalNets, "legalnet", "", "legal net segment for registry")
	flag.StringVar(&sentryDSN, "sentry", "http://7:6@sentry.lain.cloud/3", "sentry Data Source Name")
	flag.StringVar(&ldapUrl, "ldapurl", "http://ldap.lain.cloud/", "ldap address")

	// give a ldap user for search some info for ssolib treat all users as public
	// also let the user be the init Admin of sso-ldap
	flag.StringVar(&ldapUser, "ldapuser", "test", "some ldap user")
	flag.StringVar(&ldapPassword, "ldappasswd", "test", "the password of the ldap user")
	flag.StringVar(&ldapBase, "ldapbase", "", "the ldap search base")

	flag.Parse()

	user.InitAdmin = ldapUser

	if isDebug {
		log.EnableDebug()
	}

	log.Debug(ldapUrl, ldapUser, ldapPassword, mysqlDSN, emailSuffix, ldapBase)

	userback := user.New(ldapUrl, ldapUser, ldapPassword, mysqlDSN, emailSuffix, ldapBase)

	server := ssolib.NewServer(mysqlDSN, siteURL, smtpAddr, emailFrom, emailSuffix, isDebug, prikeyfile, pubkeyfile, sentryDSN, false)

	server.SetUserBackend(userback)

	log.Error(server.ListenAndServe(webAddr, ssoldap.AddHandlers))
}
