package chhash

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"errors"
	"github.com/JohnHuahuaZhan/gocommon/chkey"
	"hash"
)

const (
	defaultSaltLen    = 16 //不能大于256
	defaultIterations = 32
	defaultKeyLen     = 32
	sumLen            = defaultSaltLen + 6 + defaultKeyLen
)

var encString = []byte("chkey")
var defaultHashFunction = sha512.New

var WrongFormatPasswd = errors.New("wrong password format")
var WrongPasswd = errors.New("wrong password")

type Options struct {
	SaltLen      int
	Iterations   int
	KeyLen       int
	HashFunction func() hash.Hash
}

func generateSalt(length int) []byte {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	salt := make([]byte, length)
	rand.Read(salt)
	for key, val := range salt {
		salt[key] = alphanum[val%byte(len(alphanum))]
	}
	return salt
}

// Encode takes two arguments, a raw password, and a pointer to an Options struct.
// In order to use default options, pass `nil` as the second argument.
// It returns the generated salt and encoded supper_key for the user.
func Encode(rawPwd []byte, options *Options) ([]byte, []byte) {
	if options == nil {
		salt := generateSalt(defaultSaltLen)
		encodedPwd := chkey.Key(rawPwd, salt, defaultIterations, defaultKeyLen, defaultHashFunction)
		return salt, encodedPwd
	}
	salt := generateSalt(options.SaltLen)
	encodedPwd := chkey.Key(rawPwd, salt, options.Iterations, options.KeyLen, options.HashFunction)
	return salt, encodedPwd
}

// Verify takes four arguments, the raw password, its generated salt, the encoded password,
// and a pointer to the Options struct, and returns a boolean value determining whether the password is the correct one or not.
// Passing `nil` as the last argument resorts to default options.
func Verify(rawPwd, salt, encodedPwd []byte, options *Options) bool {
	if options == nil {
		return bytes.Equal(encodedPwd, chkey.Key(rawPwd, salt, defaultIterations, defaultKeyLen, defaultHashFunction))
	}
	return bytes.Equal(encodedPwd, chkey.Key(rawPwd, salt, options.Iterations, options.KeyLen, options.HashFunction))
}

// Passwd sha512 default options. salt append const string chkey append encoded passwd
func Passwd(rawPwd string) string {
	b := &bytes.Buffer{}
	salt, encoded := Encode([]byte(rawPwd), nil)
	b.Write(salt)
	b.Write(encString)
	b.Write(encoded)
	return hex.EncodeToString(b.Bytes())
}
func VerifyPasswd(rawPwd string, encoded string) error {
	enc, err := hex.DecodeString(encoded)
	if err != nil || len(enc) != sumLen || !bytes.Equal(enc[defaultSaltLen:defaultSaltLen+6], encString) {
		return WrongFormatPasswd
	}
	if !Verify([]byte(rawPwd), enc[:defaultSaltLen], enc[defaultSaltLen+6:], nil) {
		return WrongPasswd
	}
	return nil
}
