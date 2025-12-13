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

func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	// 1. E-postayı al
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

	// 2. Firebase'den Şifre Sıfırlama Linki Üret
	// Bu fonksiyon, kullanıcıya tıklatacağımız o uzun güvenli linki oluşturur.
	resetLink, err := AuthClient.PasswordResetLink(ctx, req.Email)
	if err != nil {
		// Kullanıcı bulunamazsa da güvenlik gereği "Email gönderildi" denir
		// ama loglara hatayı basabilirsin.
		log.Printf("Link üretme hatası: %v", err)
		http.Error(w, "İşlem başarısız: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Brevo ile Email Gönder
	err = sendEmailWithBrevo(req.Email, resetLink)
	if err != nil {
		log.Printf("Brevo hatası: %v", err)
		http.Error(w, "Mail gönderilemedi", http.StatusInternalServerError)
		return
	}

	// 4. Başarılı
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Şifre sıfırlama bağlantısı e-posta adresine gönderildi.",
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	// 1. İsteği Oku
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz JSON", http.StatusBadRequest)
		return
	}

	// 2. Firebase REST API'ye İstek At (Simülasyon)
	// NOT: Bu "API_KEY"i environment variable'dan almak daha güvenlidir.
	// Şimdilik test için buraya string olarak yapıştırabilirsin ya da os.Getenv kullanabilirsin.
	webAPIKey := os.Getenv("FIREBASE_WEB_API_KEY")
	if webAPIKey == "" {
		// Hata almamak için şimdilik hardcoded yazabilirsin ama production'da yapma:
		// webAPIKey = "BURAYA_FIREBASE_CONSOLE_DAN_ALDIGIN_KEYI_YAZ"
		http.Error(w, "API Key eksik", http.StatusInternalServerError)
		return
	}

	url := "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=" + webAPIKey

	// Google'a gidecek veri
	requestBody, _ := json.Marshal(map[string]interface{}{
		"email":             req.Email,
		"password":          req.Password,
		"returnSecureToken": true,
	})

	// İsteği Gönder
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		http.Error(w, "Google API hatası", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 3. Cevabı Kontrol Et
	if resp.StatusCode != http.StatusOK {
		// Şifre yanlışsa buraya düşer
		http.Error(w, "Giriş başarısız! E-posta veya şifre hatalı.", http.StatusUnauthorized)
		return
	}

	// 4. Başarılıysa Token Bilgilerini Döndür
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		http.Error(w, "Cevap okunamadı", http.StatusInternalServerError)
		return
	}

	// İPUCU: Burada ileride kullanıcının ROLÜNÜ de (admin/user) veritabanından çekip dönebiliriz.
	// Proje isterlerinde "Başarılı giriş sonrası rol otomatik atanır" diyor[cite: 27].

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Giriş başarılı",
		"uid":     loginResp.LocalId,
		"token":   loginResp.IdToken, // Bu token ile ileride güvenli işlem yapılacak
	})
}
