package sigapi

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	h "net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// AuthHd header key of JWT value
	AuthHd = "authHd"
	// MalformedHd is the error message sent when the
	// header of a request is empty
	MalformedHd = "Malformed header"
	// NotJWTUser is the error message sent when the
	// JWTUser type assertion fails. This is a fatal
	// security breach since it can only occurr when
	// the private key is compromised.
	NotJWTUser = `False JWTUser type assertion. 
	Security breach. Private key compromised`
)

type Credentials struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

func credentials(r *h.Request) (c *Credentials, e error) {
	var bs []byte
	bs, e = ioutil.ReadAll(r.Body)
	if e == nil {
		r.Body.Close()
		c = new(Credentials)
		e = json.Unmarshal(bs, c)
	}
	return
}

func writeErr(w h.ResponseWriter, e error) {
	if e != nil {
		// The order of the following commands matter since
		// httptest.ResponseRecorder ignores parameter sent
		// to WriteHeader if Write was called first
		w.WriteHeader(h.StatusBadRequest)
		w.Write([]byte(e.Error()))
	}
}

// Decode decodes an io.Reader with a JSON formatted object
func Decode(r io.Reader, v interface{}) (e error) {
	var bs []byte
	bs, e = ioutil.ReadAll(r)
	if e == nil {
		e = json.Unmarshal(bs, v)
	}
	return
}

// Encode encodes an object in JSON notation into w
func Encode(w io.Writer, v interface{}) (e error) {
	cd := json.NewEncoder(w)
	cd.SetIndent("", "	")
	e = cd.Encode(v)
	return
}

// JWTCrypt is the object for encrypting JWT
type JWTCrypt struct {
	pKey *rsa.PrivateKey
}

// JWTUser adds jwt.StandardClaims to an User
type JWTUser struct {
	User string `json:"user"`
	jwt.StandardClaims
}

// NewJWTCrypt creates a new JWTCrypt
func NewJWTCrypt() (j *JWTCrypt) {
	j = new(JWTCrypt)
	var e error
	j.pKey, e = rsa.GenerateKey(rand.Reader, 512)
	if e != nil {
		panic(e.Error())
	}
	return
}

func (j *JWTCrypt) encrypt(c *Credentials) (s string, e error) {
	uc := &JWTUser{User: c.User}
	uc.ExpiresAt = time.Now().Add(time.Hour).Unix()
	t := jwt.NewWithClaims(jwt.GetSigningMethod("RS256"), uc)
	s, e = t.SignedString(j.pKey)
	return
}

func (j *JWTCrypt) decrypt(r *h.Request) (usr string, e error) {
	s := r.Header.Get(AuthHd)
	if s == "" {
		e = HeaderErr()
	}
	t, e := jwt.ParseWithClaims(s, &JWTUser{},
		func(x *jwt.Token) (a interface{}, d error) {
			a, d = &j.pKey.PublicKey, nil
			return
		})
	var clm *JWTUser
	if e == nil {
		var ok bool
		clm, ok = t.Claims.(*JWTUser)
		if !ok || clm.User == "" {
			panic(NotJWTUser)
			// { the private key was used to sign something
			//   different from a *JWTUser, which is not
			//   done in this program, therefore it has
			//   been compromised }
		}
	}
	if e == nil {
		usr = clm.User
	}
	return
}

func HeaderErr() (e error) {
	e = fmt.Errorf(MalformedHd)
	return
}
