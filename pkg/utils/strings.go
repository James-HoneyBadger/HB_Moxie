package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/crypto/pbkdf2"

	"github.com/James-HoneyBadger/honey-ledger/pkg/errs"
)

const (
	availableCharacters                      = "!#$&()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[]^_abcdefghijklmnopqrstuvwxyz{|}~"
	availableNumberAndLetters                = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	availableNumberAndLowercaseLetters       = "0123456789abcdefghijklmnopqrstuvwxyz"
	availableCharactersLength                = len(availableCharacters)
	availableNumberAndLettersLength          = len(availableNumberAndLetters)
	availableNumberAndLowercaseLettersLength = len(availableNumberAndLowercaseLetters)

	passwordHashAlgorithm     = "pbkdf2-sha256"
	passwordHashSeparator     = "$"
	passwordHashIterations    = 600000
	passwordHashKeyLength     = 48
	legacyPasswordIterations  = 10000
	maxPasswordHashIterations = 10000000
)

// SubString returns part of the source string according to start index and length
func SubString(str string, start int, length int) string {
	chars := []rune(str)
	realLength := len(chars)
	end := 0

	if start < 0 {
		start = realLength + start
	}

	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}

	if start > realLength {
		start = realLength
	}

	if end < 0 {
		end = 0
	}

	if end > realLength {
		end = realLength
	}

	return string(chars[start:end])
}

// ContainsAnyString returns whether the specified string contains any string of sub string slice
func ContainsAnyString(s string, substrs []string) bool {
	for i := 0; i < len(substrs); i++ {
		if strings.Index(s, substrs[i]) >= 0 {
			return true
		}
	}

	return false
}

// GetFirstLowerCharString returns the source string parameter, but makes the first character lower case
func GetFirstLowerCharString(s string) string {
	if s == "" {
		return s
	}

	chars := []rune(s)

	if unicode.IsLower(chars[0]) {
		return s
	}

	chars[0] = unicode.ToLower(chars[0])
	return string(chars)
}

// ContainsOnlyOneRune returns the source string only contains one character
func ContainsOnlyOneRune(s string, r rune) bool {
	if len(s) < 1 {
		return false
	}

	for i := 0; i < len(s); i++ {
		if rune(s[i]) != r {
			return false
		}
	}

	return true
}

// GetRandomString returns a random string of which length is n
func GetRandomString(n int) (string, error) {
	var result = make([]byte, n)

	for i := 0; i < n; i++ {
		index, err := GetRandomInteger(availableCharactersLength)

		if err != nil {
			return "", err
		}

		result[i] = availableCharacters[index]
	}

	return string(result), nil
}

// GetRandomNumberOrLetter returns a random string which only contains number or letter characters
func GetRandomNumberOrLetter(n int) (string, error) {
	var result = make([]byte, n)

	for i := 0; i < n; i++ {
		index, err := GetRandomInteger(availableNumberAndLettersLength)

		if err != nil {
			return "", err
		}

		result[i] = availableNumberAndLetters[index]
	}

	return string(result), nil
}

// GetRandomNumberOrLowercaseLetter returns a random string which only contains number or letter characters
func GetRandomNumberOrLowercaseLetter(n int) (string, error) {
	var result = make([]byte, n)

	for i := 0; i < n; i++ {
		index, err := GetRandomInteger(availableNumberAndLowercaseLettersLength)

		if err != nil {
			return "", err
		}

		result[i] = availableNumberAndLowercaseLetters[index]
	}

	return string(result), nil
}

// MD5Encode returns a hashed string by md5
func MD5Encode(data []byte) []byte {
	m := md5.New()
	m.Write(data)
	return m.Sum(nil)
}

// MD5EncodeToString returns a hashed string by md5
func MD5EncodeToString(data []byte) string {
	hash := MD5Encode(data)
	return hex.EncodeToString(hash)
}

// SHA256Encode returns a hashed byte slice by sha256
func SHA256Encode(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// SHA256EncodeToString returns a hashed hex string by sha256
func SHA256EncodeToString(data []byte) string {
	hash := SHA256Encode(data)
	return hex.EncodeToString(hash)
}

// AESGCMEncrypt returns a encrypted string by aes-gcm
func AESGCMEncrypt(key []byte, plainText []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcm.NonceSize())

	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, plainText, nil)
	result := append(nonce, ciphertext...)

	return result, nil
}

// AESGCMDecrypt returns a decrypted string by aes-gcm
func AESGCMDecrypt(key []byte, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)

	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(block)

	if err != nil {
		return nil, err
	}

	nonceSize := aesgcm.NonceSize()

	if len(ciphertext)-nonceSize <= 0 {
		return nil, errs.ErrCiphertextInvalid
	}

	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]

	plainText, err := aesgcm.Open(nil, nonce, ciphertext, nil)

	if err != nil {
		return nil, err
	}

	return plainText, nil
}

// EncodePassword returns a versioned PBKDF2-HMAC-SHA256 hash of the password
func EncodePassword(password string, salt string) string {
	return encodePasswordWithIterations(password, salt, passwordHashIterations)
}

// EncodeRecoveryCode returns a legacy-compatible hash for two-factor recovery codes
func EncodeRecoveryCode(recoveryCode string, salt string) string {
	return encodeLegacyPassword(recoveryCode, salt)
}

// VerifyPassword returns whether the password matches and whether the stored hash
// uses an older iteration count that should be upgraded after a successful login.
func VerifyPassword(password string, salt string, encodedPassword string) (bool, bool) {
	parts := strings.Split(encodedPassword, passwordHashSeparator)

	if len(parts) != 3 || parts[0] != passwordHashAlgorithm {
		expected := encodeLegacyPassword(password, salt)
		return SecureCompare(expected, encodedPassword), true
	}

	iterations, err := strconv.Atoi(parts[1])

	if err != nil || iterations < 1 || iterations > maxPasswordHashIterations {
		return false, false
	}

	expected := encodePasswordWithIterations(password, salt, iterations)
	matched := SecureCompare(expected, encodedPassword)

	return matched, matched && iterations < passwordHashIterations
}

func encodePasswordWithIterations(password string, salt string, iterations int) string {
	encodedPassword := pbkdf2.Key([]byte(password), []byte(salt), iterations, passwordHashKeyLength, sha256.New)
	return fmt.Sprintf("%s%s%d%s%s", passwordHashAlgorithm, passwordHashSeparator, iterations, passwordHashSeparator, base64.RawStdEncoding.EncodeToString(encodedPassword))
}

func encodeLegacyPassword(password string, salt string) string {
	encodedPassword := pbkdf2.Key([]byte(password), []byte(salt), legacyPasswordIterations, passwordHashKeyLength, sha256.New)
	return strings.TrimRight(base64.StdEncoding.EncodeToString(encodedPassword), "=")
}

// SecureCompare performs a constant-time string comparison to prevent timing attacks
func SecureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// EncryptSecret returns a encrypted secret
func EncryptSecret(secret string, key string) (string, error) {
	encryptedSecret, err := AESGCMEncrypt(MD5Encode([]byte(key)), []byte(secret)) // md5encode make the aes key's length to 16

	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedSecret), nil
}

// DecryptSecret returns a decrypted secret
func DecryptSecret(encyptedSecret string, key string) (string, error) {
	encyptedData, err := base64.StdEncoding.DecodeString(encyptedSecret)

	if err != nil {
		return "", err
	}

	secret, err := AESGCMDecrypt(MD5Encode([]byte(key)), []byte(encyptedData))

	if err != nil {
		return "", err
	}

	return string(secret), nil
}
