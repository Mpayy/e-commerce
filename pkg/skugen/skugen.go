package skugen

import (
    "crypto/rand"
    "fmt"
    "math/big"
    "strings"
    "time"
)

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Generate menghasilkan SKU format: PRD-{6 karakter random}-{4 digit akhir unix timestamp}
// Contoh output: PRD-K7X2QA-7291
func Generate() string {
    randomPart := make([]byte, 6)
    for i := range randomPart {
        n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        randomPart[i] = charset[n.Int64()]
    }

    // 4 digit terakhir unix timestamp untuk tambahan uniqueness
    timestamp := fmt.Sprintf("%d", time.Now().Unix())
    timePart := timestamp[len(timestamp)-4:]

    return fmt.Sprintf("PRD-%s-%s", string(randomPart), timePart)
}

// Sanitize membersihkan SKU yang diinput manual:
// uppercase semua, trim spasi, ganti spasi tengah dengan dash
func Sanitize(sku string) string {
    return strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(sku, " ", "-")))
}