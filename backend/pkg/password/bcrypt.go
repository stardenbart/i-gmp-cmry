package password

import "golang.org/x/crypto/bcrypt"

const defaultCost = bcrypt.DefaultCost

// Hash generates a bcrypt hash from a plain-text password.
func Hash(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), defaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Compare checks whether a plain-text password matches a bcrypt hash.
// Returns true if they match.
func Compare(plain, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain))
	return err == nil
}
