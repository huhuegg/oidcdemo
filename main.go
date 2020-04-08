package main

import (
	"log"
	"net/http"
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"fmt"
	"encoding/json"
	"github.com/satori/go.uuid"
	"strings"
	"os"
	"github.com/sevlyar/go-daemon"
)

const (
	SERVICE_NAME = "oidc_demo"
)

var (
	scopes = []string{}
	callbackUrl = ""
)

var (
	provideUrl = os.Getenv("OIDC_PRIVIDE_URL") //http://keycloak.test.com:8080/auth/realms/master
	clientID = os.Getenv("OIDC_CLIENT_ID")
	clientSecret = os.Getenv("OIDC_CLIENT_SECRET")
	callbackPattern = os.Getenv("OIDC_CALLBACK_PATTERN")
	host = os.Getenv("OIDC_HOST")
	oidc_scopes = os.Getenv("OIDC_SCOPES")
	certFile = os.Getenv("OIDC_HOST_CRET_FILE")
	keyFile = os.Getenv("OIDC_HOST_KEY_FILE")
	logPath = os.Getenv("OIDC_LOG_PATH")
)

func main() {
	err := checkEnv()
	if err != nil {
		log.Printf("checkEnv failed, err:%s\n",err)
	}

	//daemon
	pidFile := fmt.Sprintf("%s/%s.pid",logPath,SERVICE_NAME)
	logFile := fmt.Sprintf("%s/%s.log",logPath,SERVICE_NAME)
	log.Println("# Start Daemon #")
	log.Printf("Log: %s\n",logFile)
	cntxt := &daemon.Context{
		PidFileName: pidFile,
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		WorkDir:     "./",
		Umask:       027,
		Args:        []string{},
	}

	child, err := cntxt.Reborn()
	if err != nil {
		log.Printf("Unable to run, err:%s\n", err)
		return
	}
	if child != nil {
		return
	}
	defer cntxt.Release()

	serviceLogic()
}

func checkEnv() error {
	if provideUrl == "" {
		return fmt.Errorf("OIDC_PRIVIDE_URL missing")
	}
	if clientID == "" {
		return fmt.Errorf("OIDC_CLIENT_ID missing")
	}
	if clientSecret == "" {
		return fmt.Errorf("OIDC_CLIENT_SECRET missing")
	}
	if callbackPattern == "" {
		return fmt.Errorf("OIDC_CALLBACK_PATTERN missing")
	}
	if host == "" {
		return fmt.Errorf("OIDC_HOST missing")
	}
	if oidc_scopes == "" {
		return fmt.Errorf("OIDC_SCOPES missing")
	}
	scopes = strings.Split(oidc_scopes,",")
	if certFile == "" {
		return fmt.Errorf("OIDC_HOST_CRET_FILE missing")
	}
	if keyFile == "" {
		return fmt.Errorf("OIDC_HOST_KEY_FILE missing")
	}
	if logPath == "" {
		logPath = "/tmp"
	}
	callbackUrl = fmt.Sprintf("https://%s%s",host,callbackPattern)
	return nil
}

func serviceLogic() {
	ctx := context.Background()

	fmt.Printf("provideUrl: %s\n",provideUrl)
	provider, err := oidc.NewProvider(ctx, provideUrl)
	if err != nil {
		fmt.Println("provideDiscover error")
		log.Fatal(err)
	}

	endPoint := provider.Endpoint()
	fmt.Printf("EndPoint:\n%s\n",jsonStr(endPoint))
	config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     provider.Endpoint(),
		RedirectURL:  callbackUrl,
		Scopes: scopes,
	}

	state := randomStr()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s := config.AuthCodeURL(state)
		fmt.Printf("debug AuthCodeUrl: \n%s\n",s)
		http.Redirect(w, r, s, http.StatusFound)
	})

	http.HandleFunc(callbackPattern, func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Query().Get("state") != state {
			http.Error(w, "state did not match", http.StatusBadRequest)
			return
		}

		oauth2Token, err := config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			http.Error(w, "Failed to exchange token: "+err.Error(), http.StatusInternalServerError)
			return
		}

		oauth2Token.AccessToken = oauth2Token.Extra("access_token").(string)

		fmt.Println("#UserInfo#")
		userInfo, err := provider.UserInfo(ctx, oauth2.StaticTokenSource(oauth2Token))
		if err != nil {
			http.Error(w, "Failed to get userinfo: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(stringToBytes(jsonStr(userInfo)))
	})

	s := &http.Server{
		Addr: ":443",
		Handler: nil,
	}
	log.Printf("listening on https://%s/", host)
	e := s.ListenAndServeTLS(certFile, keyFile)
	if e != nil {
		log.Fatal("ListenAndServeTLS: ", e)
	}
}



func jsonStr(i interface{}) string {
	_, str := toJson(i)
	return str
}

func toJson(i interface{}) (error, string) {
	data, error := json.Marshal(&i)
	if error != nil {
		return error, ""
	} else {
		return nil, string(data)
	}
}

func randomStr() string {
	id,err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return strings.Replace(id.String(), "-", "", -1)
}

func stringToBytes(str string) []byte {
	return []byte(str)
}