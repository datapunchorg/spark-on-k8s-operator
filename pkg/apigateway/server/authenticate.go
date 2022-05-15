/*
Copyright spark-on-k8s-operator contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"net/http"
	"strconv"
	"strings"
)

type AuthenticationHandler interface {
	ValidateUserPassword(user string, password string) error
}

func BasicAuthForRealm(authenticationHandler AuthenticationHandler, realm string) gin.HandlerFunc {
	if realm == "" {
		realm = "Authorization Required"
	}
	realm = "Basic realm=" + strconv.Quote(realm)
	return func(c *gin.Context) {
		headerValue := c.Request.Header.Get("Authorization")
		if headerValue == "" {
			c.Header("WWW-Authenticate", realm)
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("did not find Authorization header in client request"))
			return
		}
		prefixValue := "Basic "
		if !strings.HasPrefix(headerValue, prefixValue) {
			c.Header("WWW-Authenticate", realm)
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid Authorization header value (not start with: %s) in client request", prefixValue))
			return
		}
		credentialsBase64 := headerValue[len(prefixValue):]
		decodedCredentials, err := base64.StdEncoding.DecodeString(credentialsBase64)
		if err != nil {
			msg := "invalid Authorization header value (not encoded in base64 properly) in client request"
			glog.Warningf(msg)
			c.Header("WWW-Authenticate", realm)
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf(msg))
			return
		}
		decodedCredentialsStr := string(decodedCredentials)
		index := strings.Index(decodedCredentialsStr, ":")
		if index <= 0 {
			msg := "invalid Authorization header value (not base64 encoded value like user:password) in client request"
			glog.Warningf(msg)
			c.Header("WWW-Authenticate", realm)
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf(msg))
			return
		}
		user := decodedCredentialsStr[0:index]
		password := decodedCredentialsStr[index+1:]
		err = authenticationHandler.ValidateUserPassword(user, password)
		if err != nil {
			glog.Warningf("failed to validate user password for %s: %s", user, err.Error())
			c.Header("WWW-Authenticate", realm)
			c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("invalid user or password"))
			return
		}

		// The user credentials was found, set user's id to key AuthUserKey in this context, the user's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		c.Set(gin.AuthUserKey, user)
	}
}
