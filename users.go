package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetUserProfileHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.URL.Query().Get("user_id")

	if userID == "" {
		http.Error(w, "user_id parametresi zorunludur", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	dsnap, err := FirestoreClient.Collection("users").Doc(userID).Get(ctx)

	if err != nil {

		if status.Code(err) == codes.NotFound {
			http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
		} else {
			http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
		}
		return
	}

	var user User
	if err := dsnap.DataTo(&user); err != nil {
		http.Error(w, "Veri işleme hatası", http.StatusInternalServerError)
		return
	}

	response := ProfileResponse{
		FullName:   user.FullName,
		Email:      user.Email,
		Department: user.Department,
		Role:       user.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateUserProfileHandler(w http.ResponseWriter, r *http.Request) {

	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id parametresi zorunludur", http.StatusBadRequest)
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	updates := []firestore.Update{}

	if req.Department != "" {
		updates = append(updates, firestore.Update{Path: "department", Value: req.Department})
	}
	if len(req.NotificationPrefs) > 0 {
		updates = append(updates, firestore.Update{Path: "notification_prefs", Value: req.NotificationPrefs})
	}

	if len(updates) == 0 {
		http.Error(w, "Güncellenecek veri gönderilmedi", http.StatusBadRequest)
		return
	}

	_, err := FirestoreClient.Collection("users").Doc(userID).Update(ctx, updates)
	if err != nil {

		if status.Code(err) == codes.NotFound {
			http.Error(w, "Kullanıcı bulunamadı", http.StatusNotFound)
		} else {
			log.Printf("Güncelleme hatası: %v", err)
			http.Error(w, "Profil güncellenemedi", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profil başarıyla güncellendi",
	})
}
