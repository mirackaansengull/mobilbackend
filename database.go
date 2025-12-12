package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	ctx := context.Background()

	sa := option.WithCredentialsFile("/etc/secrets/serviceAccountKey.json")

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("HATA: Firebase başlatılamadı: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("HATA: Firestore veritabanına bağlanılamadı: %v", err)
	}

	defer client.Close()

	log.Println("BAŞARILI: Firestore veritabanına bağlantı sağlandı!")
}
