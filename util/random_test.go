package util

import (
	"strings"
	"testing"
)

func TestRandomInt(t *testing.T) {
	// Test the RandomInt function
	n := RandomInt(0, 1000)
	if n <= 0 || n > 1000 {
		t.Errorf("RandomInt() failed")
	}
}

func TestRandomString(t *testing.T) {
	// Test the RandomString function
	for i := 6; i < 12; i++ {
		s := RandomString(i)
		if len(s) != i {
			t.Errorf("RandomString() failed")
		}
	}
}

func TestRandomOwner(t *testing.T) {
	// Test the RandomOwner function
	s := RandomOwner()
	if len(s) != 6 {
		t.Errorf("RandomOwner() failed")
	}
}

func TestRandomMoney(t *testing.T) {
	// Test the RandomMoney function
	n := RandomMoney()
	if n <= 0 || n > 1000 {
		t.Errorf("RandomMoney() failed")
	}
}
func TestRandomCurrency(t *testing.T) {
	// Test the RandomCurrency function
	currency := RandomCurrency()
	if currency != USD && currency != EUR && currency != BRL {
		t.Errorf("RandomCurrency() failed, got: %s", currency)
	}
}

func TestRandomEmail(t *testing.T) {
	// Test the RandomEmail function
	email := RandomEmail()
	if !strings.Contains(email, "@email.com") || len(email) <= 10 {
		t.Errorf("RandomEmail() failed, got: %s", email)
	}
}
