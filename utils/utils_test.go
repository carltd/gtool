package utils

import "testing"

func TestStrFirstToUpper(t *testing.T) {
	str := "ab_cd"
	result := StrFirstToUpper(str)
	if result != "AbCd" {
		t.Errorf("%s != AbCd", result)
	}
}
