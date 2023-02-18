package util

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890-_."
func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max - min + 1)
}

func RandomString(n int) string {
	var sb strings.Builder

	k := len(alphabet)

	for i := 0; i <n ; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandomUsername() string {
	return RandomString(int(RandomInt(1,25)))
}

func RandomEmail() string {
	email := fmt.Sprintf("%s@groceryplanner.com", RandomUsername())

	return email
}

func RandomRole() string {
	roles := []string{"common", "admin"}
	n := len(roles)

	return(roles[rand.Intn(n)])
}

func RandomUnit() string {
	return RandomString(int(RandomInt(2,5)))
}

func RandomIngredient() string {
	return RandomString(int(RandomInt(5,100)))
}