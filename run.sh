#!/bin/bash

# TODO: remove sleep after lain has a more stable calico ip allocation

sleep 5

# for ldaps
cp /lain/app/newcert.crt /usr/local/share/ca-certificates/newcert.crt
update-ca-certificates

DOMAIN=${LAIN_DOMAIN:-"lain.local"}
source ./secrets

DEBUG=${DEBUG:-"false"}

exec ./sso-0.1.linux.amd64 -domain="@example.com" -from="sso-ldap@$DOMAIN" -mysql="$MYSQL" -site="https://sso-ldap.$DOMAIN" -smtp="$SMTP" -web=":80" -sentry="$SENTRY" -ldapurl="$LDAPURL" -ldapuser="$LDAPUSER" -ldappasswd="$LDAPPASSWORD" -ldapbase="$LDAPBASE" -debug="$DEBUG"
