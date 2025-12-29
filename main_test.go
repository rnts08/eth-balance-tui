package main

import (
	"math/big"
	"testing"
	"time"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"hello world", 5, "he..."},
		{"short", 10, "short"},
		{"exact", 5, "ex..."},
		{"", 5, ""},
		{"abc", 2, "ab"}, // Test safety fix for small widths
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q; want %q", tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestAddCommas(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123", "123"},
		{"1234", "1,234"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"1234.56", "1,234.56"},
		{"-1234", "-1,234"},
		{"", ""},
	}

	for _, tt := range tests {
		result := addCommas(tt.input)
		if result != tt.expected {
			t.Errorf("addCommas(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		decimals int
		expected string
	}{
		{1234.5678, 2, "1,234.57"},
		{1234.5, 2, "1,234.50"},
		{0, 2, "0.00"},
	}

	for _, tt := range tests {
		result := formatFloat(tt.input, tt.decimals)
		if result != tt.expected {
			t.Errorf("formatFloat(%f, %d) = %q; want %q", tt.input, tt.decimals, result, tt.expected)
		}
	}
}

func TestFormatBigFloat(t *testing.T) {
	tests := []struct {
		input    *big.Float
		decimals int
		expected string
	}{
		{big.NewFloat(1234.5678), 2, "1,234.57"},
		{nil, 2, "0"},
	}

	for _, tt := range tests {
		result := formatBigFloat(tt.input, tt.decimals)
		if result != tt.expected {
			t.Errorf("formatBigFloat(%v, %d) = %q; want %q", tt.input, tt.decimals, result, tt.expected)
		}
	}
}

func TestMasking(t *testing.T) {
	m := model{privacyMode: true}

	if got := m.maskString("100"); got != "****" {
		t.Errorf("maskString() = %q; want ****", got)
	}
	if got := m.maskAddress("0x123456"); got != "0x**...**" {
		t.Errorf("maskAddress() = %q; want 0x**...**", got)
	}

	m.privacyMode = false
	if got := m.maskString("100"); got != "100" {
		t.Errorf("maskString() = %q; want 100", got)
	}
	if got := m.maskAddress("0x123456"); got != "0x123456" {
		t.Errorf("maskAddress() = %q; want 0x123456", got)
	}
}

func TestGetFilteredTransactions(t *testing.T) {
	acc := &accountState{
		address: "0xMyAddress",
		transactions: []txInfo{
			{Hash: "0x1", From: "0xMyAddress", To: "0xOther"}, // Out
			{Hash: "0x2", From: "0xOther", To: "0xMyAddress"}, // In
			{Hash: "0x3", From: "0xOther", To: "0xOther"},     // Irrelevant
		},
	}

	m := model{}

	// Test All
	m.txFilter = "all"
	txs := m.getFilteredTransactions(acc)
	if len(txs) != 3 {
		t.Errorf("Expected 3 transactions for 'all', got %d", len(txs))
	}

	// Test In
	m.txFilter = "in"
	txs = m.getFilteredTransactions(acc)
	// Logic: !isFrom
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions for 'in', got %d", len(txs))
	}

	// Test Out
	m.txFilter = "out"
	txs = m.getFilteredTransactions(acc)
	// Logic: isFrom
	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction for 'out', got %d", len(txs))
	}
	if len(txs) > 0 && txs[0].Hash != "0x1" {
		t.Errorf("Expected hash 0x1, got %s", txs[0].Hash)
	}
}

func TestCalculateAccountTotal(t *testing.T) {
	m := model{
		chains: []ChainConfig{
			{Name: "Ethereum", CoinGeckoID: "ethereum", Tokens: []TokenConfig{{Symbol: "USDC", CoinGeckoID: "usd-coin"}}},
		},
		prices: map[string]float64{
			"ethereum": 2000.0,
			"usd-coin": 1.0,
		},
	}

	acc := &accountState{
		balances: map[string]*big.Float{
			"Ethereum": big.NewFloat(1.5), // 1.5 * 2000 = 3000
		},
		tokenBalances: map[string]map[string]*big.Float{
			"Ethereum": {
				"USDC": big.NewFloat(100), // 100 * 1 = 100
			},
		},
	}

	total := m.calculateAccountTotal(acc)
	fTotal, _ := total.Float64()

	expected := 3100.0
	if fTotal != expected {
		t.Errorf("calculateAccountTotal = %f; want %f", fTotal, expected)
	}
}

func TestGetPrioritizedRPCs(t *testing.T) {
	m := model{
		rpcCooldowns: map[string]time.Time{
			"rpc_cooldown": time.Now().Add(time.Minute),
		},
		rpcLatencies: map[string]time.Duration{
			"rpc_fast":  10 * time.Millisecond,
			"rpc_slow":  100 * time.Millisecond,
			"rpc_error": -1,
			// rpc_unknown is missing from map
		},
	}

	input := []string{"rpc_slow", "rpc_cooldown", "rpc_error", "rpc_fast", "rpc_unknown"}

	// Expected order logic:
	// 1. Healthy (not in cooldown)
	// 2. Valid Latency (lowest first)
	// 3. Unknown Latency
	// 4. Error Latency
	// 5. Cooldown

	got := m.getPrioritizedRPCs(input)

	// Check cooldown is last
	if len(got) > 0 && got[len(got)-1] != "rpc_cooldown" {
		t.Errorf("Expected rpc_cooldown to be last, got %v", got)
	}

	// Check fast before slow
	fastIdx := -1
	slowIdx := -1
	for i, r := range got {
		if r == "rpc_fast" {
			fastIdx = i
		}
		if r == "rpc_slow" {
			slowIdx = i
		}
	}
	if fastIdx == -1 || slowIdx == -1 || fastIdx > slowIdx {
		t.Errorf("Expected rpc_fast before rpc_slow, got %v", got)
	}
}
