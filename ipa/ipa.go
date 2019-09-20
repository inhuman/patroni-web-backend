package ipa

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/robbert229/jwt"
	"github.com/ubccr/goipa"
	"github.com/vmware/vic/pkg/errors"
	"jgit.me/tools/patroni-web-backend/config"
	"net/http"
	"time"
)

const JwtSalt = "M3eCjiBmp9bC7attoQDJa1HYn4KoY7QadT2lDi3M3b2ThoLTz9J2Sx6FK4Jsg8LBDdkk2vefk98nLQdP"

var JwtAlgorithm = jwt.HmacSha256(JwtSalt)

type User struct {
	Uid    string        `json:"uid"`
	Groups []string      `json:"groups"`
	Name   ipa.IpaString `json:"name"`
	Token  string        `json:"token"`
}

var Client *ipa.Client

func Init() {
	httpClient := &http.Client{
		Timeout: 1 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	Client = ipa.NewClientCustomHttp(config.AppConf.Ipa.Host, "", httpClient)
}

func CheckAuth(token string) error {
	return JwtAlgorithm.Validate(token)
}

func GetUserByToken(token string) (*User, error) {

	cl, err := JwtAlgorithm.Decode(token)
	if err != nil {
		return nil, err
	}

	u := &User{
		Token: token,
	}

	if f, err := cl.Get("uid"); err == nil {
		u.Uid = f.(string)
	}

	if f, err := cl.Get("groups"); err == nil {
		if err := json.Unmarshal([]byte(f.(string)), &u.Groups); err != nil {
			return nil, err
		}
	}

	if f, err := cl.Get("name"); err == nil {
		u.Name = f.(ipa.IpaString)
	}

	return u, nil
}

func AuthUserJwt(login, password string) (*User, error) {

	err := Client.RemoteLogin(login, password)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("ipa auth error: %s\n", err))
	}

	userInfo, err := Client.UserShow(login)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("fetch user from ipa error: %s\n", err))
	}

	token, err := getUserJwtToken(userInfo)
	if err != nil {
		return nil, err
	}

	u := &User{
		Uid:    login,
		Groups: userInfo.Groups,
		Name:   userInfo.DisplayName,
		Token:  token,
	}

	return u, nil
}

func getUserJwtToken(userInfo *ipa.UserRecord) (string, error) {

	gr, err := json.Marshal(userInfo.Groups)
	if err != nil {
		return "", err
	}

	claims := jwt.NewClaim()
	claims.Set("groups", string(gr))
	claims.Set("uid", userInfo.Uid)
	claims.Set("Name", userInfo.DisplayName)
	claims.SetTime("exp", time.Now().Add(time.Hour*time.Duration(730)))

	token, err := JwtAlgorithm.Encode(claims)

	if err != nil {
		return "", err
	}

	return token, nil
}
