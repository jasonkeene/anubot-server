package store

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const tmpl = "$scrypt$ln=%d,r=%d,p=%d$%s$%s\n"

// Hash computes a scrypt PHC string that can be used to verify the password
// at a later time. For more information on PHC string format see:
// https://github.com/P-H-C/phc-string-format/blob/master/phc-sf-spec.md
func Hash(password string) (phc string, err error) {
	salt, err := createSalt()
	if err != nil {
		return "", err
	}

	const (
		ln  = 16      // ln is the exponent for the cost parameter
		n   = 1 << ln // n is the CPU/memory cost parameter
		r   = 8       // r is the memory cost parameter
		p   = 1       // p is the parallelization cost parameter
		len = 32      // len is the desired key length
	)
	hash, err := scrypt.Key([]byte(password), salt, n, r, p, len)
	if err != nil {
		return "", err
	}

	b64hash := base64.RawStdEncoding.EncodeToString(hash)
	b64salt := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf(tmpl, ln, r, p, b64salt, b64hash), nil
}

func createSalt() ([]byte, error) {
	buf := make([]byte, 32)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// Verify takes a password and a PHC string and verifies if the PHC string was
// originally computed with the password or not.
func Verify(password, phc string) (valid bool, err error) {
	parts, err := parsePHC(phc)
	if err != nil {
		return false, err
	}

	hash, err := scrypt.Key(
		[]byte(password),
		parts.salt,
		1<<uint(parts.ln),
		parts.r,
		parts.p,
		len(parts.hash),
	)
	if err != nil {
		return false, err
	}

	return bytes.Equal(hash, parts.hash), nil
}

type phcParts struct {
	ln   int
	r    int
	p    int
	salt []byte
	hash []byte
}

func parsePHC(phc string) (phcParts, error) {
	parts := strings.Split(phc, "$")
	if len(parts) != 5 {
		return phcParts{}, errors.New("invalid hash length")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return phcParts{}, err
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return phcParts{}, err
	}

	params, err := parseParams(parts[2])
	if err != nil {
		return phcParts{}, err
	}

	return phcParts{
		ln:   params.ln,
		r:    params.r,
		p:    params.p,
		salt: salt,
		hash: hash,
	}, nil
}

type phcParams struct {
	ln int
	r  int
	p  int
}

func parseParams(params string) (phcParams, error) {
	var (
		p           phcParams
		lnw, rw, pw bool
	)

	paramPairs := strings.Split(params, ",")

	if len(paramPairs) != 3 {
		return phcParams{}, errors.New("invalid params")
	}
	for _, epair := range paramPairs {
		pair := strings.Split(epair, "=")
		if len(pair) != 2 {
			return phcParams{}, errors.New("invalid params")
		}
		var err error
		switch pair[0] {
		case "ln":
			if lnw {
				return phcParams{}, errors.New("invalid params")
			}
			lnw = true
			p.ln, err = strconv.Atoi(pair[1])
			if err != nil {
				return phcParams{}, errors.New("invalid params")
			}
		case "r":
			if rw {
				return phcParams{}, errors.New("invalid params")
			}
			rw = true
			p.r, err = strconv.Atoi(pair[1])
			if err != nil {
				return phcParams{}, errors.New("invalid params")
			}
		case "p":
			if pw {
				return phcParams{}, errors.New("invalid params")
			}
			pw = true
			p.p, err = strconv.Atoi(pair[1])
			if err != nil {
				return phcParams{}, errors.New("invalid params")
			}
		default:
			return phcParams{}, errors.New("invalid params")
		}
	}
	return p, nil
}
