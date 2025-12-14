package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func sendEmailWithBrevo(toEmail, resetLink string) error {
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("BREVO_API_KEY ortam değişkeni bulunamadı")
	}

	htmlBody := fmt.Sprintf(`
		<h1>Şifre Sıfırlama Talebi</h1>
		<p>Hesabınızın şifresini sıfırlamak için aşağıdaki bağlantıya tıklayın:</p>
		<p><a href="%s">Şifremi Sıfırla</a></p>
		<p>Bu işlemi siz yapmadıysanız, bu maili görmezden gelebilirsiniz.</p>
	`, resetLink)

	emailReq := BrevoEmailRequest{
		Sender:      BrevoUser{Name: "Kampüs Güvenlik", Email: "mobilprogramlama123@gmail.com"},
		To:          []BrevoUser{{Email: toEmail}},
		Subject:     "Şifre Sıfırlama İşlemi",
		HtmlContent: htmlBody,
	}

	payload, _ := json.Marshal(emailReq)

	req, err := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("api-key", apiKey)
	req.Header.Set("content-type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("brevo api hatası: status code %d", resp.StatusCode)
	}

	return nil
}
