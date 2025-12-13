package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

// ---------------------------------------------------------
// YARDIMCI FONKSİYON: Brevo API İsteği
// ---------------------------------------------------------
func sendEmailWithBrevo(toEmail, resetLink string) error {
	apiKey := os.Getenv("BREVO_API_KEY") // API Key'i buradan alıyoruz
	if apiKey == "" {
		return fmt.Errorf("BREVO_API_KEY ortam değişkeni bulunamadı")
	}

	// HTML Mail İçeriği
	htmlBody := fmt.Sprintf(`
		<h1>Şifre Sıfırlama Talebi</h1>
		<p>Hesabınızın şifresini sıfırlamak için aşağıdaki bağlantıya tıklayın:</p>
		<p><a href="%s">Şifremi Sıfırla</a></p>
		<p>Bu işlemi siz yapmadıysanız, bu maili görmezden gelebilirsiniz.</p>
	`, resetLink)

	// JSON Verisini Hazırla
	emailReq := BrevoEmailRequest{
		Sender:      BrevoUser{Name: "Kampüs Güvenlik", Email: "mobilprogramlama123@gmail.com"}, // Buraya kendi onaylı mailini yazarsan daha iyi olur
		To:          []BrevoUser{{Email: toEmail}},
		Subject:     "Şifre Sıfırlama İşlemi",
		HtmlContent: htmlBody,
	}

	payload, _ := json.Marshal(emailReq)

	// HTTP POST İsteği Oluştur
	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	// Header Ayarları
	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)
	req.Header.Set("content-type", "application/json")

	// İsteği Gönder
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 200 veya 201 dönmezse hata var demektir
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("brevo api hatası: status code %d", resp.StatusCode)
	}

	return nil
}
