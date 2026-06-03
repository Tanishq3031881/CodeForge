package rooms

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// Word lists kept short and friendly — enough entropy for human-shareable
// URLs without being a dependency. adjective-noun-NN gives ~24*24*100 ≈ 57k
// combinations; the service retries on the rare collision.
var (
	adjectives = []string{
		"purple", "crimson", "golden", "silent", "cosmic", "electric", "frozen",
		"hidden", "lucky", "brave", "swift", "clever", "gentle", "mighty",
		"noble", "rapid", "shiny", "stormy", "sunny", "velvet", "amber",
		"jade", "coral", "ivory",
	}
	nouns = []string{
		"fox", "otter", "falcon", "panda", "tiger", "wolf", "heron", "lynx",
		"comet", "ember", "river", "willow", "cedar", "maple", "pixel", "vector",
		"cipher", "quasar", "nimbus", "harbor", "meadow", "summit", "canyon", "lagoon",
	}
)

// GenerateSlug returns a random slug like "purple-fox-42".
func GenerateSlug() (string, error) {
	adj, err := pick(adjectives)
	if err != nil {
		return "", err
	}
	noun, err := pick(nouns)
	if err != nil {
		return "", err
	}
	n, err := rand.Int(rand.Reader, big.NewInt(100))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%02d", adj, noun, n.Int64()), nil
}

func pick(list []string) (string, error) {
	i, err := rand.Int(rand.Reader, big.NewInt(int64(len(list))))
	if err != nil {
		return "", err
	}
	return list[i.Int64()], nil
}
