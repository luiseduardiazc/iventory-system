package testutil

import (
	"fmt"
	"math/rand"
	"time"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateTestID genera un ID único para tests
func GenerateTestID() string {
	return fmt.Sprintf("test-%d-%d", time.Now().UnixNano(), seededRand.Intn(10000))
}

// GenerateTestSKU genera un SKU único para tests
func GenerateTestSKU() string {
	return fmt.Sprintf("SKU-TEST-%d", time.Now().UnixNano())
}

// GenerateTestStoreID genera un ID de tienda para tests
func GenerateTestStoreID() string {
	stores := []string{"MAD-001", "BCN-001", "VAL-001", "SEV-001", "BIL-001"}
	return stores[seededRand.Intn(len(stores))]
}

// PtrString devuelve un puntero a string
func PtrString(s string) *string {
	return &s
}

// PtrInt devuelve un puntero a int
func PtrInt(i int) *int {
	return &i
}

// PtrBool devuelve un puntero a bool
func PtrBool(b bool) *bool {
	return &b
}

// PtrTime devuelve un puntero a time.Time
func PtrTime(t time.Time) *time.Time {
	return &t
}
