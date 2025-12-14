package main

import "time"

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	Department string `json:"department"`
}

type User struct {
	ID                string    `firestore:"id"`
	Email             string    `firestore:"email"`
	FullName          string    `firestore:"full_name"`
	Department        string    `firestore:"department"`
	Role              string    `firestore:"role"`
	NotificationPrefs []string  `firestore:"notification_prefs"`
	CreatedAt         time.Time `firestore:"created_at"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type BrevoEmailRequest struct {
	Sender      BrevoUser   `json:"sender"`
	To          []BrevoUser `json:"to"`
	Subject     string      `json:"subject"`
	HtmlContent string      `json:"htmlContent"`
}

type BrevoUser struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Google'dan Dönen Cevap Yapısı
type LoginResponse struct {
	IdToken      string `json:"idToken"` // Mobil uygulamada kullanacağımız token
	Email        string `json:"email"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
	LocalId      string `json:"localId"` // User ID (UID)
}

type ProfileResponse struct {
	FullName   string `json:"full_name"`
	Email      string `json:"email"`
	Department string `json:"department"`
	Role       string `json:"role"`
}

type UpdateProfileRequest struct {
	Department        string   `json:"department"`
	NotificationPrefs []string `json:"notification_prefs"` // Örn: ["Sağlık", "Güvenlik"]
}
