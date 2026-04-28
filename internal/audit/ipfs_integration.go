// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

package audit

import (
	"context"
	"fmt"
	"time"

	"github.com/dotandev/hintents/internal/decenstorage"
	"github.com/dotandev/hintents/internal/session"
)

// AuditLog represents the interface for audit logs that can be published to IPFS
type AuditLog interface {
	// ToJSON converts the audit log to JSON bytes for IPFS storage
	ToJSON() ([]byte, error)
}

// AuditPublisher handles publishing audit reports to IPFS and updating session data
type AuditPublisher struct {
	publisher *decenstorage.Publisher
}

// NewAuditPublisher creates a new AuditPublisher with default IPFS configuration
func NewAuditPublisher() *AuditPublisher {
	cfg := decenstorage.PublishConfig{}
	publisher := decenstorage.New(cfg)
	return &AuditPublisher{
		publisher: publisher,
	}
}

// NewAuditPublisherWithConfig creates a new AuditPublisher with custom configuration
func NewAuditPublisherWithConfig(cfg decenstorage.PublishConfig) *AuditPublisher {
	publisher := decenstorage.New(cfg)
	return &AuditPublisher{
		publisher: publisher,
	}
}

// PublishAuditReport publishes a signed audit log to IPFS and updates the session
func (ap *AuditPublisher) PublishAuditReport(ctx context.Context, auditLog AuditLog, sessionData *session.Data) error {
	// Marshal the audit log to JSON
	auditJSON, err := auditLog.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	// Publish to IPFS
	result, err := ap.publisher.PublishIPFS(ctx, auditJSON)
	if err != nil {
		return fmt.Errorf("failed to publish audit report to IPFS: %w", err)
	}

	// Update session data with IPFS information
	sessionData.AuditCID = result.CID
	sessionData.AuditURL = result.URL
	sessionData.AuditTimestamp = time.Now().UTC().Format(time.RFC3339)

	return nil
}

// PublishAndSave publishes an audit report and saves the updated session
func (ap *AuditPublisher) PublishAndSave(ctx context.Context, auditLog AuditLog, sessionData *session.Data) error {
	// Publish to IPFS
	if err := ap.PublishAuditReport(ctx, auditLog, sessionData); err != nil {
		return err
	}

	// Save the updated session
	store, err := session.NewStore()
	if err != nil {
		return fmt.Errorf("failed to create session store: %w", err)
	}
	defer store.Close()

	if err := store.Save(ctx, sessionData); err != nil {
		return fmt.Errorf("failed to save updated session: %w", err)
	}

	return nil
}

// VerifyAuditTrail retrieves and verifies an audit report from IPFS
func (ap *AuditPublisher) VerifyAuditTrail(ctx context.Context, cid string) (*AuditLog, error) {
	// For now, we'll implement a simple verification that checks if the CID exists
	// In a full implementation, you would retrieve the content from IPFS and verify it
	// This is a placeholder for the verification functionality

	// TODO: Implement IPFS content retrieval and verification
	return nil, fmt.Errorf("verification not yet implemented - CID: %s", cid)
}

// GetAuditURL returns the public gateway URL for an audit report
func (ap *AuditPublisher) GetAuditURL(cid string) string {
	return fmt.Sprintf("https://ipfs.io/ipfs/%s", cid)
}
