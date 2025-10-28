package testutil

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

// GenerateTestID genera un ID único para tests usando UUID
func GenerateTestID() string {
	return uuid.New().String()
}

// GenerateTestSKU genera un SKU único para tests
func GenerateTestSKU() string {
	return fmt.Sprintf("SKU-TEST-%d-%s", time.Now().UnixNano(), uuid.New().String()[:8])
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

// GenerateSKU genera un SKU único
func GenerateSKU() string {
	return fmt.Sprintf("SKU-%s", uuid.New().String()[:13])
}

// GenerateID genera un ID único
func GenerateID() string {
	return uuid.New().String()
}

// GenerateCustomerID genera un ID de cliente
func GenerateCustomerID(id int) string {
	return fmt.Sprintf("customer-%d-%s", id, uuid.New().String()[:8])
}
