package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"firebase.google.com/go/v4/auth"
)

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.FullName == "" {
		http.Error(w, "Eksik bilgi", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

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
		AuthClient.DeleteUser(ctx, u.UID)
		http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Kayıt başarılı",
		"user_id": u.UID,
	})
}

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {

	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email adresi zorunludur", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	resetLink, err := AuthClient.PasswordResetLink(ctx, req.Email)
	if err != nil {

		log.Printf("Link üretme hatası: %v", err)
		http.Error(w, "İşlem başarısız: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = sendEmailWithBrevo(req.Email, resetLink)
	if err != nil {
		log.Printf("Brevo hatası: %v", err)
		http.Error(w, "Mail gönderilemedi", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Şifre sıfırlama bağlantısı e-posta adresine gönderildi.",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}

	webAPIKey := os.Getenv("FIREBASE_WEB_API_KEY")
	if webAPIKey == "" {

		http.Error(w, "API Key eksik", http.StatusInternalServerError)
		return
	}

	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + webAPIKey

	requestBody, _ := json.Marshal(map[string]interface{}{
		"email":             req.Email,
		"password":          req.Password,
		"returnSecureToken": true,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, "Google API hatası", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {

		http.Error(w, "Giriş başarısız! E-posta veya şifre hatalı.", http.StatusUnauthorized)
		return
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		http.Error(w, "Cevap okunamadı", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Giriş başarılı",
		"uid":     loginResp.LocalId,
		"token":   loginResp.IdToken,
	})
}
