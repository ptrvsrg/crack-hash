package infrastructure

type HashBruteForce interface {
	BruteForceMD5(hash string, alphabet []string, maxLength, partNumber int) ([]string, error)
}

type Services struct {
	HashBruteForce HashBruteForce
}
