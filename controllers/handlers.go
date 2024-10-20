package controllers

import (
	"encoding/json"
	"io"
	"key_store/store"
	"net/http"
	"time"
)

type SetBody struct {
	Key string `json:"key"`
}

type RequestData struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	TTL   string `json:"ttl"`
}

func Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Please provide a key", http.StatusNotFound)
		return
	}
	val := store.KvStore.Get(key)
	if val == "" {
		http.Error(w, "Key does not exist", http.StatusNotFound)
		return
	}
	res := map[string]string{
		"data": val,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func Set(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	var data RequestData
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error decoding request", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	if data.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}
	var ttl time.Duration
	ttl, err = time.ParseDuration(data.TTL)
	if err != nil {
		ttl = time.Hour * 24
	}
	store.KvStore.Set(
		data.Key,
		store.KvMapValue{Value: data.Value, ExpireAt: time.Now().Add(ttl)},
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]string{"status": "success", "message": "Key-value pair set successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	var data RequestData
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, "Error decoding response", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	if data.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}
	isUpdated := store.KvStore.Update(data.Key, data.Value)
	w.Header().Set("Content-Type", "application/json")
	var response map[string]string
	if isUpdated {
		response = map[string]string{"status": "success", "message": "Key updated successfully"}
		w.WriteHeader(http.StatusOK)
	} else {
		response = map[string]string{"status": "failure", "message": "Couldn't update successfully"}
		w.WriteHeader(http.StatusNotFound)
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Please provide key", http.StatusBadRequest)
		return
	}
	store.KvStore.Delete(key)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{"status": "success", "message": "Key-Value deleted successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}
