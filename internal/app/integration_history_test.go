package app

import (
	"testing"
	"time"

	"github.com/manson/port-chaser/internal/models"
	"github.com/manson/port-chaser/internal/storage"
)

// TestHistoryPersistence verifies that history is properly persisted
// when a process is killed and loaded when the app starts.
func TestHistoryPersistence(t *testing.T) {
	// Create temporary storage
	cfg := storage.Config{
		DBPath:     "/tmp/test-history-persistence.db",
		WALEnabled: false,
		Timeout:    50,
	}
	
	sto, err := storage.NewSQLite(cfg)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer func() {
		sto.Close()
		// Clean up
		// (In real test, you'd clean up temp files properly)
	}()
	
	// Create a model with storage
	model := Model{
		Storage: sto,
		History: []models.HistoryEntry{},
	}
	
	// Test 1: Load initial history (should be empty)
	cmd := model.loadHistoryCmd()
	msg := cmd()
	
	historyMsg, ok := msg.(HistoryLoadedMsg)
	if !ok {
		t.Fatalf("Expected HistoryLoadedMsg, got %T", msg)
	}
	
	if historyMsg.Error != nil {
		t.Fatalf("Failed to load history: %v", historyMsg.Error)
	}
	
	if len(historyMsg.History) != 0 {
		t.Fatalf("Expected empty history, got %d entries", len(historyMsg.History))
	}
	
	// Test 2: Simulate a port kill and record to history
	entry := models.HistoryEntry{
		PortNumber:  3000,
		ProcessName: "node",
		PID:         12345,
		Command:     "node server.js",
		KilledAt:    time.Now(),
	}
	
	if err := sto.RecordKill(entry); err != nil {
		t.Fatalf("Failed to record kill: %v", err)
	}
	
	// Test 3: Load history again - should have 1 entry
	cmd = model.loadHistoryCmd()
	msg = cmd()
	
	historyMsg, ok = msg.(HistoryLoadedMsg)
	if !ok {
		t.Fatalf("Expected HistoryLoadedMsg, got %T", msg)
	}
	
	if historyMsg.Error != nil {
		t.Fatalf("Failed to load history: %v", historyMsg.Error)
	}
	
	if len(historyMsg.History) != 1 {
		t.Fatalf("Expected 1 history entry, got %d", len(historyMsg.History))
	}
	
	if historyMsg.History[0].PortNumber != 3000 {
		t.Errorf("Expected port 3000, got %d", historyMsg.History[0].PortNumber)
	}
	
	// Test 4: Verify GetKillCount
	count, err := sto.GetKillCount(3000, 30)
	if err != nil {
		t.Fatalf("Failed to get kill count: %v", err)
	}
	
	if count != 1 {
		t.Errorf("Expected kill count 1, got %d", count)
	}
}

// TestModelWithoutStorage verifies that the model works correctly
// when storage is nil (no persistence).
func TestModelWithoutStorage(t *testing.T) {
	model := Model{
		Storage: nil,
		History: []models.HistoryEntry{},
	}
	
	// Should not panic when storage is nil
	cmd := model.Init()
	if cmd == nil {
		// If no storage, Init should still return a valid batch
		t.Fatal("Init should return a valid command even without storage")
	}
}
