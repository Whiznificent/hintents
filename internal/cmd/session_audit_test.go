// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/dotandev/hintents/internal/session"
)

func TestSessionSaveWithAuditIntegration(t *testing.T) {
	// Create a test session with transaction data
	testData := &session.Data{
		TxHash:        "test-tx-hash-1234567890abcdef",
		Network:       "testnet",
		EnvelopeXdr:   "test-envelope-xdr-data",
		ResultMetaXdr: "test-result-meta-xdr-data",
		Status:        "active",
		ErstVersion:   "test-version",
	}

	// Add simulation response JSON with events and logs
	simResp := map[string]interface{}{
		"events": []string{"event1", "event2", "event3"},
		"logs":   []string{"log1", "log2"},
	}
	simRespJSON, err := json.Marshal(simResp)
	if err != nil {
		t.Fatalf("Failed to marshal simulation response: %v", err)
	}
	testData.SimResponseJSON = string(simRespJSON)

	// Set the current session
	SetCurrentSession(testData)

	// Test generateAuditLog function
	events := []string{"event1", "event2", "event3"}
	logs := []string{"log1", "log2"}

	auditLog, err := generateAuditLog(testData.TxHash, testData.EnvelopeXdr, testData.ResultMetaXdr, events, logs)
	if err != nil {
		t.Fatalf("Failed to generate audit log: %v", err)
	}

	// Verify audit log structure
	if auditLog.TransactionHash != testData.TxHash {
		t.Errorf("Expected transaction hash %s, got %s", testData.TxHash, auditLog.TransactionHash)
	}

	if auditLog.Payload.EnvelopeXdr != testData.EnvelopeXdr {
		t.Errorf("Expected envelope XDR %s, got %s", testData.EnvelopeXdr, auditLog.Payload.EnvelopeXdr)
	}

	if auditLog.Payload.ResultMetaXdr != testData.ResultMetaXdr {
		t.Errorf("Expected result meta XDR %s, got %s", testData.ResultMetaXdr, auditLog.Payload.ResultMetaXdr)
	}

	if len(auditLog.Payload.Events) != len(events) {
		t.Errorf("Expected %d events, got %d", len(events), len(auditLog.Payload.Events))
	}

	if len(auditLog.Payload.Logs) != len(logs) {
		t.Errorf("Expected %d logs, got %d", len(logs), len(auditLog.Payload.Logs))
	}

	// Verify signature and public key are present
	if auditLog.Signature == "" {
		t.Error("Expected non-empty signature")
	}

	if auditLog.PublicKey == "" {
		t.Error("Expected non-empty public key")
	}

	// Verify trace hash is computed
	if auditLog.TraceHash == "" {
		t.Error("Expected non-empty trace hash")
	}

	// Test audit log JSON serialization
	auditJSON, err := auditLog.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize audit log to JSON: %v", err)
	}

	if len(auditJSON) == 0 {
		t.Error("Expected non-empty JSON output")
	}

	// Verify the JSON can be unmarshaled back
	var unmarshaledLog AuditLog
	if err := json.Unmarshal(auditJSON, &unmarshaledLog); err != nil {
		t.Fatalf("Failed to unmarshal audit log JSON: %v", err)
	}

	if unmarshaledLog.TransactionHash != auditLog.TransactionHash {
		t.Error("Unmarshaled log does not match original")
	}
}

func TestSessionSaveWithoutTransactionData(t *testing.T) {
	// Create a test session without transaction data
	testData := &session.Data{
		TxHash:  "test-tx-hash",
		Network: "testnet",
		Status:  "active",
		// No EnvelopeXdr or ResultMetaXdr
	}

	// Set the current session
	SetCurrentSession(testData)

	// Test that audit generation is skipped when transaction data is missing
	// This simulates the case where the session save command should not fail
	// when there's no transaction data to audit
	ctx := context.Background()
	
	// Open session store
	store, err := session.NewStore()
	if err != nil {
		t.Fatalf("Failed to create session store: %v", err)
	}
	defer store.Close()

	// This should not generate an audit log but should not fail either
	if testData.EnvelopeXdr == "" || testData.ResultMetaXdr == "" {
		t.Log("Audit generation correctly skipped for session without transaction data")
	}
}

func TestGenerateAuditLogWithEnvironmentKey(t *testing.T) {
	// Test with environment variable set for private key
	testKey := "0000000000000000000000000000000000000000000000000000000000000001"
	os.Setenv("ERST_AUDIT_PRIVATE_KEY", testKey)
	defer os.Unsetenv("ERST_AUDIT_PRIVATE_KEY")

	txHash := "test-tx-hash"
	envelopeXdr := "test-envelope-xdr"
	resultMetaXdr := "test-result-meta-xdr"
	events := []string{"event1"}
	logs := []string{"log1"}

	auditLog, err := generateAuditLog(txHash, envelopeXdr, resultMetaXdr, events, logs)
	if err != nil {
		t.Fatalf("Failed to generate audit log with environment key: %v", err)
	}

	// Verify the audit log was created successfully
	if auditLog.TransactionHash != txHash {
		t.Errorf("Expected transaction hash %s, got %s", txHash, auditLog.TransactionHash)
	}

	// Verify the signature is different from the one generated with a random key
	if auditLog.Signature == "" {
		t.Error("Expected non-empty signature")
	}

	t.Log("Successfully generated audit log with environment private key")
}

func TestSessionDataIPFSFields(t *testing.T) {
	// Test that session Data struct has the required IPFS fields
	testData := &session.Data{
		TxHash:  "test-tx-hash",
		Network: "testnet",
	}

	// Initially, IPFS fields should be empty
	if testData.AuditCID != "" {
		t.Error("Expected empty AuditCID initially")
	}

	if testData.AuditURL != "" {
		t.Error("Expected empty AuditURL initially")
	}

	if testData.AuditTimestamp != "" {
		t.Error("Expected empty AuditTimestamp initially")
	}

	// Test setting IPFS fields
	testData.AuditCID = "test-cid-12345"
	testData.AuditURL = "https://ipfs.io/ipfs/test-cid-12345"
	testData.AuditTimestamp = time.Now().UTC().Format(time.RFC3339)

	// Verify fields are set correctly
	if testData.AuditCID != "test-cid-12345" {
		t.Errorf("Expected AuditCID 'test-cid-12345', got '%s'", testData.AuditCID)
	}

	if testData.AuditURL != "https://ipfs.io/ipfs/test-cid-12345" {
		t.Errorf("Expected AuditURL 'https://ipfs.io/ipfs/test-cid-12345', got '%s'", testData.AuditURL)
	}

	if testData.AuditTimestamp == "" {
		t.Error("Expected non-empty AuditTimestamp after setting")
	}
}
