package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

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
