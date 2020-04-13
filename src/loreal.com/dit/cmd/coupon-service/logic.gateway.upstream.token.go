package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

//APIAccount - Upstream API Account
var APIAccount *UpstreamAccount

func (a *App) getAPIAccount() *UpstreamAccount {
	if APIAccount != nil {
		return APIAccount
	}
	APIAccount = a.NewUpstreamAccount()
	return APIAccount
}

//UpstreamTokenStore - token store for upstream system
type UpstreamTokenStore struct {
	Scope            string    `json:"scope,omitempty"`
	TokenType        string    `json:"token_type,omitempty"`
	AccessToken      string    `json:"access_token,omitempty"`
	RefreshToken     string    `json:"refresh_token,omitempty"`
	ExpiresIn        int       `json:"expires_in,omitempty"`
	JTI              string    `json:"jti,omitempty"`
	Error            string    `json:"error,omitempty"`
	ErrorDescription string    `json:"error_description,omitempty"`
	RefreshAt        time.Time `json:"-"`
}

//Reset - reset store
func (s *UpstreamTokenStore) Reset() {
	s.Scope = ""
	s.TokenType = ""
	s.AccessToken = ""
	s.RefreshToken = ""
	s.ExpiresIn = 0
	s.JTI = ""
	s.Error = ""
	s.ErrorDescription = ""
}

//UpstreamAccount - account for upstream system
type UpstreamAccount struct {
	TokenURL     string
	ClientID     string
	ClientSecret string
	UserName     string
	Password     string
	store        *UpstreamTokenStore
	mutex        *sync.Mutex
}

//NewUpstreamAccount - create upstream account from config
func (a *App) NewUpstreamAccount() *UpstreamAccount {
	const oauthPath = "oauth/token"
	return &UpstreamAccount{
		TokenURL:     a.Config.UpstreamURL + oauthPath,
		ClientID:     a.Config.UpstreamClientID,
		ClientSecret: a.Config.UpstreamClientSecret,
		UserName:     a.Config.UpstreamUserName,
		Password:     a.Config.UpstreamPassword,
		store:        &UpstreamTokenStore{},
		mutex:        &sync.Mutex{},
	}
}

//GetToken - get or refresh Token
func (a *UpstreamAccount) GetToken(forced bool) string {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if a.store.RefreshToken == "" {
		return a.getToken()
	}
	expiresAt := a.store.RefreshAt.Add(time.Second * time.Duration(a.store.ExpiresIn-10))
	if !forced && expiresAt.After(time.Now()) {
		return a.store.AccessToken
	}
	token := a.refreshToken()
	if token != "" {
		return token
	}
	return a.getToken()
}

func (a *UpstreamAccount) getToken() string {
	payload := url.Values{}
	payload.Add("grant_type", "password")
	payload.Add("username", a.UserName)
	payload.Add("password", a.Password)

	req, _ := http.NewRequest(http.MethodPost, a.TokenURL, strings.NewReader(payload.Encode()))
	req.SetBasicAuth(a.ClientID, a.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[ERR] - [UpstreamTokenStore][getToken], http.do err: %v", err)
		return ""
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&a.store); err != nil {
		log.Printf("[ERR] - [UpstreamTokenStore][getToken], decode err: %v", err)
		return ""
	}
	if a.store.Error != "" {
		log.Printf("[ERR] - [UpstreamTokenStore][getToken], token err: %s, %s", a.store.Error, a.store.ErrorDescription)
		//a.store.Reset()
		return ""
	}
	a.store.RefreshAt = time.Now()
	return a.store.AccessToken
}

func (a *UpstreamAccount) refreshToken() string {
	payload := url.Values{}
	payload.Add("grant_type", "refresh_token")
	payload.Add("refresh_token", a.store.RefreshToken)

	req, _ := http.NewRequest(http.MethodPost, a.TokenURL, strings.NewReader(payload.Encode()))
	req.SetBasicAuth(a.ClientID, a.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("[ERR] - [UpstreamTokenStore][refreshToken], http.do err: %v", err)
		return ""
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&a.store); err != nil {
		log.Printf("[ERR] - [UpstreamTokenStore][refreshToken], decode err: %v", err)
		a.store.Reset()
		return ""
	}
	if a.store.Error != "" {
		log.Printf("[ERR] - [UpstreamTokenStore][refreshToken], token err: %s, %s", a.store.Error, a.store.ErrorDescription)
		a.store.Reset()
		return ""
	}
	a.store.RefreshAt = time.Now()
	return a.store.AccessToken
}
