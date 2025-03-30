package wallet

import "testing"

func TestNewWallet(t *testing.T) {
	w := NewWallet()
	if w.Address == "" {
		t.Error("Not empty wallet address expected")
	}
	if w.PrivateKey == nil || w.PublicKey == nil {
		t.Error("Private and public keys expected")
	}
}

func TestCreateTransaction(t *testing.T) {
	w := NewWallet()
	tx := w.CreateTransaction("someAddress", 10.0, 0)
	if tx.From != w.Address {
		t.Errorf("Expected from %s, got %s", w.Address, tx.From)
	}
	if tx.Amount != 10.0 {
		t.Errorf("Expected amount 10.0, got %f", tx.Amount)
	}
	if tx.Signature == "" {
		t.Error("Expected non-empty signature")
	}
}
