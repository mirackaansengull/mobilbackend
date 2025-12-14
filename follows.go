package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	//"google.golang.org/grpc/codes"
	//"google.golang.org/grpc/status"
)

func FollowReportHandler(w http.ResponseWriter, r *http.Request) {

	var req FollowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz veri", http.StatusBadRequest)
		return
	}

	if req.UserID == "" || req.ReportID == "" {
		http.Error(w, "user_id ve report_id zorunludur", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	userRef := FirestoreClient.Collection("users").Doc(req.UserID)

	if r.Method == http.MethodPost {
		_, err := userRef.Update(ctx, []firestore.Update{
			{
				Path:  "followed_reports",
				Value: firestore.ArrayUnion(req.ReportID),
			},
		})
		if err != nil {
			log.Printf("Takip ekleme hatası: %v", err)
			http.Error(w, "İşlem başarısız", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Bildirim takip listesine eklendi"})

	} else if r.Method == http.MethodDelete {
		_, err := userRef.Update(ctx, []firestore.Update{
			{
				Path:  "followed_reports",
				Value: firestore.ArrayRemove(req.ReportID),
			},
		})
		if err != nil {
			log.Printf("Takipten çıkma hatası: %v", err)
			http.Error(w, "İşlem başarısız", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "Bildirim takipten çıkarıldı"})

	} else {
		http.Error(w, "Sadece POST ve DELETE metodları desteklenir", http.StatusMethodNotAllowed)
	}
}
