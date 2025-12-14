package main

import (
	"context"
	"encoding/json"
	"net/http"

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
