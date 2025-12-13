package main

import "time"

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	Department string `json:"department"` // Birim bilgisi
}

// Veritabanına kaydedilecek veri yapısı
type User struct {
	ID         string    `firestore:"id"`
	Email      string    `firestore:"email"`
	FullName   string    `firestore:"full_name"`
	Department string    `firestore:"department"`
	Role       string    `firestore:"role"` // Varsayılan: user
	CreatedAt  time.Time `firestore:"created_at"`
}
