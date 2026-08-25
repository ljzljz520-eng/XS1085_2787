package memorial

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"
)

type IDSource struct {
	seed string
}

func NewIDSource(seed string) IDSource {
	if seed == "" {
		seed = "memorial-fixture"
	}
	return IDSource{seed: seed}
}

func (s IDSource) ID(kind string, n int) string {
	h := sha256.Sum256([]byte(s.seed + ":" + kind + ":" + strconv.Itoa(n)))
	return kind + "-" + hex.EncodeToString(h[:6])
}

func fixedNow() time.Time {
	return time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
}
