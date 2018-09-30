/*
 defines the auth module
*/
package model

import (
	"errors"
	"fmt"
	"time"

	"github.com/dailing/levlog"
	"github.com/dgrijalva/jwt-go"
)

type UserAuth struct {
	UserID      int64 `xorm:"pk,autoincr"`
	Salt        string
	PswHash     string
	AccessLevel int64
}

type Token struct {
	UserID      int64
	AccessLevel int64
	ExpireAt    time.Time
	jwt.StandardClaims
}

func (t *Token) getToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, t)
	ss, err := token.SignedString([]byte("Secret"))
	levlog.E(err)
	return ss
}

func ParseToken(tokenString string) (*Token, error) {
	pt := &Token{}
	token, err := jwt.ParseWithClaims(tokenString, pt, func(token *jwt.Token) (interface{}, error) {
		return []byte("Secret"), nil
	})
	if _, ok := token.Claims.(*Token); ok && token.Valid {
		if pt.ExpireAt.Before(time.Now()) {
			return nil, errors.New("Data expired")
		}
	} else {
		levlog.E(err)
		return nil, err
	}
	return pt, nil
}

func (u *UserAuth) GetToken() string {
	claims := &Token{
		UserID:      u.UserID,
		AccessLevel: u.AccessLevel,

		ExpireAt: time.Now(),
	}
	fmt.Print(claims)
	return ""
}
