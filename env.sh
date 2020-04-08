#!/bin/sh

export OIDC_PRIVIDE_URL=http://keycloak.test.com:8080/auth/realms/master
export OIDC_CLIENT_ID=client-mac
export OIDC_CLIENT_SECRET=4a7d05e7-060b-4b11-8eaa-cad33936ec31
export OIDC_CALLBACK_PATTERN=/auth/callback
export OIDC_HOST=mac.test.com
export OIDC_SCOPES="openid,profile,email"
export OIDC_HOST_CRET_FILE=/etc/ssl/certs/ssl.cer
export OIDC_HOST_KEY_FILE=/etc/ssl/certs/ssl.key
export OIDC_LOG_PATH=/tmp
