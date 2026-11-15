package config

import (
	"context"
	"testing"
)

func TestNewRedisClient(t *testing.T) {
	ctx := context.Background()

	client, err := NewRedisClient(ctx)
	if err != nil {
		t.Fatalf("Gagal menghubungkan Redis: %v", err)
	}
	defer client.Close()

	t.Log("Berhasil terhubung ke Redis!")
}
