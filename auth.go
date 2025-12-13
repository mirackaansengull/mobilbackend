package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"firebase.google.com/go/v4/auth"
)

// Handler Fonksiyonu
func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	// JSON çözümleme
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}

	// Basit doğrulama
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		http.Error(w, "Eksik bilgi", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// main.go'da tanımladığımız AuthClient'ı kullanıyoruz
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		DisplayName(req.FullName)

	u, err := AuthClient.CreateUser(ctx, params)
	if err != nil {
		log.Printf("Kullanıcı oluşturma hatası: %v", err)
		http.Error(w, "Kullanıcı oluşturulamadı: "+err.Error(), http.StatusBadRequest)
		return
	}

	// main.go'da tanımladığımız FirestoreClient'ı kullanıyoruz
	newUser := User{
		ID:         u.UID,
		Email:      req.Email,
		FullName:   req.FullName,
		Department: req.Department,
		Role:       "user",
		CreatedAt:  time.Now(),
	}

	_, err = FirestoreClient.Collection("users").Doc(u.UID).Set(ctx, newUser)
	if err != nil {
		log.Printf("Firestore kayıt hatası: %v", err)
		AuthClient.DeleteUser(ctx, u.UID) // Temizlik
		http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
		return
	}

	// Başarılı yanıt
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Kayıt başarılı",
		"user_id": u.UID,
	})
}
