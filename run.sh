#!/bin/bash

# TODO: remove sleep after lain has a more stable calico ip allocation

sleep 5

set -ex

DOMAIN=${LAIN_DOMAIN:-"lain.local"}
source ./secrets

if [ "$LDAPS" -eq 1 ]
then
	# for ldaps
	cp /lain/app/newcert.crt /usr/local/share/ca-certificates/newcert.crt
	update-ca-certificates
fi

DEBUG=${DEBUG:-"false"}
email=${EMAIL:-"@example.com"}


exec ./sso-ldap-0.1.linux.amd64 -domain=$email -from="sso-ldap@$DOMAIN" -mysql="$MYSQL" -site="https://sso-ldap.$DOMAIN" -smtp="$SMTP" -web=":80" -sentry="$SENTRY" -ldapurl="$LDAPURL" -ldapuser="$LDAPUSER" -ldappasswd="$LDAPPASSWORD" -ldapbase="$LDAPBASE" -debug="$DEBUG"
