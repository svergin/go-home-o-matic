package tado

import (
	"fmt"
	"testing"
)

func TestXxx(t *testing.T) {
	getUserInfo()
}
func TestAuthorize(t *testing.T) {
	res, err := authorize()
	if err != nil {
		t.Fatal("Test failed due to: &v", err)
	}
	fmt.Println(res.Accesstoken)
}
