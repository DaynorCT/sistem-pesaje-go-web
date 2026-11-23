package clave

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const (
	Prefijo     = "pbkdf2"
	PrefijoSHA1 = "sha1"
	Iteraciones = 120_000
	LongitudSalt  = 16
	LongitudHash  = 32
)

func Hash(contrasena string) string {
	salt := make([]byte, LongitudSalt)
	_, _ = rand.Read(salt)
	derivado := pbkdf2.Key([]byte(contrasena), salt, Iteraciones, LongitudHash, sha256.New)
	return fmt.Sprintf("%s:%d:%s:%s",
		Prefijo,
		Iteraciones,
		base64.StdEncoding.EncodeToString(salt),
		base64.StdEncoding.EncodeToString(derivado),
	)
}

func Verificar(contrasena, almacenado string) bool {
	if contrasena == "" || almacenado == "" {
		return false
	}
	if strings.HasPrefix(almacenado, Prefijo+":") {
		return verificarPBKDF2(contrasena, almacenado)
	}
	if strings.HasPrefix(almacenado, PrefijoSHA1+":") {
		return verificarSHA1Legado(contrasena, almacenado)
	}
	return false
}

func verificarPBKDF2(contrasena, almacenado string) bool {
	partes := strings.Split(almacenado, ":")
	if len(partes) != 4 || partes[0] != Prefijo {
		return false
	}
	iteraciones, err := strconv.Atoi(partes[1])
	if err != nil || iteraciones <= 0 {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(partes[2])
	if err != nil {
		return false
	}
	esperado, err := base64.StdEncoding.DecodeString(partes[3])
	if err != nil {
		return false
	}
	derivado := pbkdf2.Key([]byte(contrasena), salt, iteraciones, LongitudHash, sha256.New)
	return subtle.ConstantTimeCompare(derivado, esperado) == 1
}

func verificarSHA1Legado(contrasena, almacenado string) bool {
	cuerpo := almacenado[len(PrefijoSHA1)+1:]
	partes := strings.SplitN(cuerpo, ":", 2)
	if len(partes) != 2 {
		return false
	}
	hexBytes, err := base64.StdEncoding.DecodeString(partes[0])
	if err != nil {
		return false
	}
	saltBytes, err := base64.StdEncoding.DecodeString(partes[1])
	if err != nil {
		return false
	}
	hexEsperado := string(hexBytes)
	salt := string(saltBytes)
	h := sha1.Sum([]byte(salt + contrasena))
	var b strings.Builder
	for _, by := range h {
		b.WriteString(fmt.Sprintf("%02x", by))
	}
	return subtle.ConstantTimeCompare([]byte(hexEsperado), []byte(b.String())) == 1
}
