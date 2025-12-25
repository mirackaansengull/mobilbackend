package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func CreateReportHandler(w http.ResponseWriter, r *http.Request) {

	var req CreateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	if req.Title == "" || req.Description == "" || req.Type == "" {
		http.Error(w, "Başlık, açıklama ve tür alanları zorunludur", http.StatusBadRequest)
		return
	}
	if req.UserID == "" {
		http.Error(w, "Kullanıcı ID (user_id) eksik", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	ref := FirestoreClient.Collection("reports").NewDoc()

	newReport := Report{
		ID:          ref.ID,
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		Status:      "Açık",
		Location: GeoPoint{
			Lat: req.Latitude,
			Lng: req.Longitude,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := ref.Set(ctx, newReport)
	if err != nil {
		log.Printf("Rapor kaydetme hatası: %v", err)
		http.Error(w, "Bildirim oluşturulamadı", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":   "Bildirim başarıyla oluşturuldu",
		"report_id": ref.ID,
	})
}

func GetReportsHandler(w http.ResponseWriter, r *http.Request) {

	statusParam := r.URL.Query().Get("status")
	typeParam := r.URL.Query().Get("type")

	ctx := context.Background()

	query := FirestoreClient.Collection("reports").OrderBy("created_at", firestore.Desc)

	if statusParam != "" {
		query = query.Where("status", "==", statusParam)
	}

	if typeParam != "" {
		query = query.Where("type", "==", typeParam)
	}

	iter := query.Documents(ctx)
	var reports []Report

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("Veri çekme hatası: %v", err)
			http.Error(w, "Veriler alınamadı", http.StatusInternalServerError)
			return
		}

		var report Report
		if err := doc.DataTo(&report); err != nil {
			continue
		}
		reports = append(reports, report)
	}

	w.Header().Set("Content-Type", "application/json")
	if reports == nil {
		reports = []Report{}
	}
	json.NewEncoder(w).Encode(reports)
}

func UpdateReportStatusHandler(w http.ResponseWriter, r *http.Request) {
	var req UpdateReportStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Geçersiz veri formatı", http.StatusBadRequest)
		return
	}

	if req.ReportID == "" || req.Status == "" {
		http.Error(w, "report_id ve status alanları zorunludur", http.StatusBadRequest)
		return
	}

	// Geçerli durum kontrolü
	validStatuses := map[string]bool{
		"Açık":       true,
		"İnceleniyor": true,
		"Çözüldü":    true,
	}

	if !validStatuses[req.Status] {
		http.Error(w, "Geçersiz durum. Geçerli durumlar: Açık, İnceleniyor, Çözüldü", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	reportRef := FirestoreClient.Collection("reports").Doc(req.ReportID)

	// Bildirimin var olup olmadığını kontrol et
	_, err := reportRef.Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			http.Error(w, "Bildirim bulunamadı", http.StatusNotFound)
		} else {
			log.Printf("Bildirim kontrol hatası: %v", err)
			http.Error(w, "Veritabanı hatası", http.StatusInternalServerError)
		}
		return
	}

	// Durumu güncelle
	_, err = reportRef.Update(ctx, []firestore.Update{
		{
			Path:  "status",
			Value: req.Status,
		},
		{
			Path:  "updated_at",
			Value: time.Now(),
		},
	})

	if err != nil {
		log.Printf("Durum güncelleme hatası: %v", err)
		http.Error(w, "Durum güncellenemedi", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Bildirim durumu başarıyla güncellendi",
		"status":  req.Status,
	})
}