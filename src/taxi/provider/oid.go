package provider

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func getIntOID() string {
	rand.Seed(time.Now().UnixNano())
	oid := strconv.Itoa(rand.Int())
	return oid
}

func getIntOIDInt32() string {
	rand.Seed(time.Now().UnixNano())
	oid := fmt.Sprint(rand.Int31())
	return oid
}

func getPrestizhOID(reqID string) string {
	oid := strings.Replace(reqID, "-", "", -1)
	if len(oid) > 30 {
		oid = oid[:30]
	}
	return oid
}
