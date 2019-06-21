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
	"strings"

	"crypto/x509"
	"encoding/base64"

	"time"

	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/SENERGY-Platform/external-task-worker/util"
)

type RoleMapping struct {
	Name string `json:"name"`
}

func getUserRoles(user string) (roles []string, err error) {
	clientToken, err := EnsureAccess()
	if err != nil {
		log.Println("ERROR: getUserRoles::EnsureAccess()", err)
		return roles, err
	}
	roleMappings := []RoleMapping{}
	err = clientToken.GetJSON(util.Config.AuthEndpoint+"/auth/admin/realms/master/users/"+user+"/role-mappings/realm", &roleMappings)
	if err != nil {
		log.Println("ERROR: getUserRoles::GetJSON()", err, util.Config.AuthEndpoint+"/auth/admin/realms/master/users/"+user+"/role-mappings/realm", string(clientToken))
		return roles, err
	}
	for _, role := range roleMappings {
		roles = append(roles, role.Name)
	}
	return
}

type KeycloakClaims struct {
	RealmAccess RealmAccess `json:"realm_access"`
	jwt.StandardClaims
}

type RealmAccess struct {
	Roles []string `json:"roles"`
}

func GetUserToken(user string) (token JwtImpersonate, err error) {
	roles, err := getUserRoles(user)
	if err != nil {
		log.Println("ERROR: GetUserToken::getUserRoles()", err)
		return token, err
	}

	// Create the Claims
	claims := KeycloakClaims{
		RealmAccess{Roles: roles},
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Duration(util.Config.JwtExpiration)).Unix(),
			Issuer:    util.Config.JwtIssuer,
			Subject:   user,
		},
	}

	jwtoken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	if util.Config.JwtPrivateKey == "" {
		unsignedTokenString, err := jwtoken.SigningString()
		if err != nil {
			log.Println("ERROR: GetUserToken::SigningString()", err)
			return token, err
		}
		tokenString := strings.Join([]string{unsignedTokenString, ""}, ".")
		token = JwtImpersonate("Bearer " + tokenString)
	} else {
		//decode key base64 string to []byte
		b, err := base64.StdEncoding.DecodeString(util.Config.JwtPrivateKey)
		if err != nil {
			log.Println("ERROR: GetUserToken::DecodeString()", err)
			return token, err
		}
		//parse []byte key to go struct key (use most common encoding)
		key, err := x509.ParsePKCS1PrivateKey(b)
		tokenString, err := jwtoken.SignedString(key)
		if err != nil {
			log.Println("ERROR: GetUserToken::SignedString()", err)
			return token, err
		}
		token = JwtImpersonate("Bearer " + tokenString)
	}
	return token, err
}
