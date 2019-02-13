/*
 * Copyright 2018 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"net/url"

	"io/ioutil"

	"github.com/SENERGY-Platform/external-task-worker/util"
)

type JwtImpersonate string

func (this JwtImpersonate) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(this))
	req.Header.Set("Content-Type", contentType)

	resp, err = http.DefaultClient.Do(req)

	if err == nil && resp.StatusCode == 401 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println(buf.String())
		err = errors.New("access denied")
	}
	return
}

func (this JwtImpersonate) PostJSON(url string, body interface{}, result interface{}) (err error) {
	b := new(bytes.Buffer)
	err = json.NewEncoder(b).Encode(body)
	if err != nil {
		return
	}
	resp, err := this.Post(url, "application/json", b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if result != nil {
		err = json.NewDecoder(resp.Body).Decode(result)
	}
	return
}

func (this JwtImpersonate) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", string(this))
	resp, err = http.DefaultClient.Do(req)
	if err == nil && resp.StatusCode == 401 {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		resp.Body.Close()
		log.Println(buf.String())
		err = errors.New("access denied")
	}
	return
}

func (this JwtImpersonate) GetJSON(url string, result interface{}) (err error) {
	resp, err := this.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(result)
}

type OpenidToken struct {
	AccessToken      string    `json:"access_token"`
	ExpiresIn        float64   `json:"expires_in"`
	RefreshExpiresIn float64   `json:"refresh_expires_in"`
	RefreshToken     string    `json:"refresh_token"`
	TokenType        string    `json:"token_type"`
	RequestTime      time.Time `json:"-"`
}

var openid *OpenidToken

func EnsureAccess() (token JwtImpersonate, err error) {
	if openid == nil {
		openid = &OpenidToken{}
	}
	duration := time.Now().Sub(openid.RequestTime).Seconds()

	if openid.AccessToken != "" && openid.ExpiresIn-util.Config.AuthExpirationTimeBuffer > duration {
		token = JwtImpersonate("Bearer " + openid.AccessToken)
		return
	}

	if openid.RefreshToken != "" && openid.RefreshExpiresIn-util.Config.AuthExpirationTimeBuffer < duration {
		log.Println("refresh token", openid.RefreshExpiresIn, duration)
		err = refreshOpenidToken(openid)
		if err != nil {
			log.Println("WARNING: unable to use refreshtoken", err)
		} else {
			token = JwtImpersonate("Bearer " + openid.AccessToken)
			return
		}
	}

	log.Println("get new access token")
	err = getOpenidToken(openid)
	if err != nil {
		log.Println("ERROR: unable to get new access token", err)
		openid = &OpenidToken{}
	}
	token = JwtImpersonate("Bearer " + openid.AccessToken)
	return
}

func getOpenidToken(token *OpenidToken) (err error) {
	requesttime := time.Now()
	resp, err := http.PostForm(util.Config.AuthEndpoint+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":     {util.Config.AuthClientId},
		"client_secret": {util.Config.AuthClientSecret},
		"grant_type":    {"client_credentials"},
	})

	if err != nil {
		log.Println("ERROR: getOpenidToken::PostForm()", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: getOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	token.RequestTime = requesttime
	return
}

func refreshOpenidToken(token *OpenidToken) (err error) {
	requesttime := time.Now()
	resp, err := http.PostForm(util.Config.AuthEndpoint+"/auth/realms/master/protocol/openid-connect/token", url.Values{
		"client_id":     {util.Config.AuthClientId},
		"client_secret": {util.Config.AuthClientSecret},
		"refresh_token": {token.RefreshToken},
		"grant_type":    {"refresh_token"},
	})

	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("ERROR: refreshOpenidToken()", resp.StatusCode, string(body))
		err = errors.New("access denied")
		resp.Body.Close()
		return
	}
	err = json.NewDecoder(resp.Body).Decode(token)
	token.RequestTime = requesttime
	return
}
