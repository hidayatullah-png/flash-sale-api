package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongo menyambung ke MongoDB (dipakai modul Review).
// Env: MONGO_URL (wajib), MONGO_DB (opsional, default "flash_sale").
func ConnectMongo() (*mongo.Client, *mongo.Database) {
	uri := os.Getenv("MONGO_URL")
	if uri == "" {
		log.Fatal("MONGO_URL belum disetel di file .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("Gagal menyambung ke MongoDB: %v", err)
	}
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB tidak merespons: %v", err)
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "flash_sale"
	}

	log.Println("Berhasil terhubung ke MongoDB!")
	return client, client.Database(dbName)
}
