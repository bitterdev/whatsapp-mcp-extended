package database

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func newTestMessageStore(t *testing.T) *MessageStore {
	t.Helper()

	db, err := sql.Open("sqlite3", "file:"+t.Name()+"?mode=memory&cache=shared&_foreign_keys=on")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := createTables(db); err != nil {
		t.Fatalf("create tables: %v", err)
	}
	if err := runMigrations(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	return &MessageStore{db: db}
}

func TestStoreChatKeepsNewestTimestamp(t *testing.T) {
	store := newTestMessageStore(t)
	older := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)
	newer := older.Add(2 * time.Hour)

	if err := store.StoreChat("123@s.whatsapp.net", "Alice", newer); err != nil {
		t.Fatalf("store newer chat: %v", err)
	}
	if err := store.StoreChat("123@s.whatsapp.net", "Alice Old", older); err != nil {
		t.Fatalf("store older chat: %v", err)
	}

	var name string
	var got time.Time
	if err := store.db.QueryRow("SELECT name, last_message_time FROM chats WHERE jid = ?", "123@s.whatsapp.net").Scan(&name, &got); err != nil {
		t.Fatalf("read chat: %v", err)
	}
	if !got.Equal(newer) {
		t.Fatalf("last_message_time = %v, want %v", got, newer)
	}
	if name != "Alice Old" {
		t.Fatalf("name = %q, want latest non-empty name", name)
	}
}

func TestStoreChatEmptyNameDoesNotEraseExistingName(t *testing.T) {
	store := newTestMessageStore(t)
	now := time.Date(2026, 5, 15, 10, 0, 0, 0, time.UTC)

	if err := store.StoreChat("123@s.whatsapp.net", "Alice", now); err != nil {
		t.Fatalf("store named chat: %v", err)
	}
	if err := store.StoreChat("123@s.whatsapp.net", "", now.Add(time.Hour)); err != nil {
		t.Fatalf("store empty-name chat: %v", err)
	}

	var name string
	if err := store.db.QueryRow("SELECT name FROM chats WHERE jid = ?", "123@s.whatsapp.net").Scan(&name); err != nil {
		t.Fatalf("read chat: %v", err)
	}
	if name != "Alice" {
		t.Fatalf("name = %q, want existing name", name)
	}
}

func TestStoreMessageFilesUnknownChat(t *testing.T) {
	store := newTestMessageStore(t)
	now := time.Date(2026, 8, 13, 20, 0, 0, 0, time.UTC)

	if err := store.StoreMessage("MSG1", "919116092926@s.whatsapp.net", "me", "me", "hello", now, true,
		"", "", "", "", nil, nil, nil, 0); err != nil {
		t.Fatalf("store message for unknown chat: %v", err)
	}

	var content string
	if err := store.db.QueryRow("SELECT content FROM messages WHERE id = ? AND chat_jid = ?",
		"MSG1", "919116092926@s.whatsapp.net").Scan(&content); err != nil {
		t.Fatalf("read message: %v", err)
	}
	if content != "hello" {
		t.Fatalf("content = %q, want %q", content, "hello")
	}
}

func TestStoreMessageKeepsExistingChatMetadata(t *testing.T) {
	store := newTestMessageStore(t)
	filed := time.Date(2026, 8, 13, 20, 0, 0, 0, time.UTC)

	if err := store.StoreChat("26578062962886@lid", "wiz", filed); err != nil {
		t.Fatalf("store chat: %v", err)
	}
	if err := store.StoreMessage("MSG2", "26578062962886@lid", "me", "me", "hello", filed.Add(-time.Hour), true,
		"", "", "", "", nil, nil, nil, 0); err != nil {
		t.Fatalf("store message: %v", err)
	}

	var name string
	var last time.Time
	if err := store.db.QueryRow("SELECT name, last_message_time FROM chats WHERE jid = ?", "26578062962886@lid").Scan(&name, &last); err != nil {
		t.Fatalf("read chat: %v", err)
	}
	if name != "wiz" {
		t.Fatalf("name = %q, want %q", name, "wiz")
	}
	if !last.Equal(filed) {
		t.Fatalf("last_message_time = %v, want %v", last, filed)
	}
}

func TestStoreMessageCarriesChatForward(t *testing.T) {
	store := newTestMessageStore(t)
	filed := time.Date(2026, 8, 13, 20, 0, 0, 0, time.UTC)

	if err := store.StoreChat("26578062962886@lid", "wiz", filed); err != nil {
		t.Fatalf("store chat: %v", err)
	}
	sent := filed.Add(time.Hour)
	if err := store.StoreMessage("MSG3", "26578062962886@lid", "me", "me", "hello", sent, true,
		"", "", "", "", nil, nil, nil, 0); err != nil {
		t.Fatalf("store message: %v", err)
	}

	var name string
	var last time.Time
	if err := store.db.QueryRow("SELECT name, last_message_time FROM chats WHERE jid = ?", "26578062962886@lid").Scan(&name, &last); err != nil {
		t.Fatalf("read chat: %v", err)
	}
	if !last.Equal(sent) {
		t.Fatalf("last_message_time = %v, want %v", last, sent)
	}
	if name != "wiz" {
		t.Fatalf("name = %q, want %q", name, "wiz")
	}
}
