package store

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"refweir.dev/pkg/idgen"
)

func BenchmarkStore_OpenDisk(b *testing.B) {
	dir := b.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dbPath := filepath.Join(dir, fmt.Sprintf("bench_%d.db", i))
		db, err := Open(dbPath)
		if err != nil {
			b.Fatalf("Open failed: %v", err)
		}
		db.Close()
	}
}

func BenchmarkStore_ReadQuery(b *testing.B) {
	ctx := context.Background()
	db, err := Open(":memory:")
	if err != nil {
		b.Fatalf("Open(:memory:): %v", err)
	}
	defer db.Close()

	userID := idgen.NewUser()
	if _, err := db.UpsertUserByEmail(ctx, "bench@refweir.dev", func() string { return userID }); err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		u, err := db.GetUserByEmail(ctx, "bench@refweir.dev")
		if err != nil || u.ID == "" {
			b.Fatalf("GetUserByEmail failed: %v", err)
		}
	}
}

func BenchmarkStore_WriteTransaction(b *testing.B) {
	ctx := context.Background()
	dir := b.TempDir()
	dbPath := filepath.Join(dir, "bench_write.db")
	db, err := Open(dbPath)
	if err != nil {
		b.Fatalf("Open failed: %v", err)
	}
	defer db.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		email := fmt.Sprintf("user_%d@refweir.dev", i)
		userID := idgen.NewUser()
		if _, err := db.UpsertUserByEmail(ctx, email, func() string { return userID }); err != nil {
			b.Fatalf("UpsertUserByEmail failed: %v", err)
		}
	}
}
