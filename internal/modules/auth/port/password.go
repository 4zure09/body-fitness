package port

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(encodedHash, password string) (bool, error)
}
