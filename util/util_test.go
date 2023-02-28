package util

import (
	"fmt"
	"testing"
)

func TestGetMinerProofType(t *testing.T) {
	p, err := GetMinerProofType("f0131822")
	fmt.Println(p, err)

	p, err = GetMinerProofType("f01231")
	fmt.Println(p, err)
}
