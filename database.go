package main

import (
	"context"
	"log"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

func main() {
	// 1. Context oluştur
	ctx := context.Background()

	sa := option.WithCredentialsFile("/etc/secrets/serviceAccountKey.json")

	// 3. Firebase uygulamasını başlat
	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("HATA: Firebase başlatılamadı: %v", err)
	}

	// 4. Firestore servisine bağlanmayı dene
	client, err := app.Firestore(ctx)
	if err != nil {
		log.Fatalf("HATA: Firestore veritabanına bağlanılamadı: %v", err)
	}

	// Bağlantı nesnesini işimiz bitince kapatmak için defer kullanıyoruz
	defer client.Close()

	// 5. Eğer buraya kadar hata almadıysak bağlantı nesnesi başarıyla oluşmuştur
	log.Println("BAŞARILI: Firestore veritabanına bağlantı sağlandı!")
}
