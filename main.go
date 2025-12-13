package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// Global Değişkenler: auth.go dosyası da bunları görebilir
var FirestoreClient *firestore.Client
var AuthClient *auth.Client

func main() {
	ctx := context.Background()

	// 1. Firebase Bağlantısı
	credentialsPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credentialsPath == "" {
		credentialsPath = "serviceAccountKey.json"
	}
	sa := option.WithCredentialsFile(credentialsPath)

	app, err := firebase.NewApp(ctx, nil, sa)
	if err != nil {
		log.Fatalf("HATA: Firebase başlatılamadı: %v", err)
	}

	// 2. Client'ları Başlat ve Global Değişkenlere Ata
	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalf("Firestore hatası: %v", err)
	}
	defer FirestoreClient.Close()

	AuthClient, err = app.Auth(ctx)
	if err != nil {
		log.Fatalf("Auth hatası: %v", err)
	}

	log.Println("BAŞARILI: Veritabanı bağlantısı hazır.")

	//Endpointler
	http.HandleFunc("POST /register", RegisterHandler)
	http.HandleFunc("POST /forgot-password", ForgotPasswordHandler)

	// Sunucuyu Başlat
	log.Println("Sunucu 8080 portunda çalışıyor...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
