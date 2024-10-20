package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	KvStore *kvStore
	once    sync.Once
)

func InitKVStore(filename string) {
	once.Do(func() {
		KvStore = newKVStore(filename)
		if err := KvStore.LoadFromFile(); err != nil {
			fmt.Println("Error loading data from file: ", err)
		}
		go KvStore.startPeriodicSave()
		go KvStore.cleanUp()
	})
}

type kvStore struct {
	Store    map[string]KvMapValue
	Mu       sync.RWMutex
	Filename string
	OpCount  int
}

type KvMapValue struct {
	Value    string    `json:"value"`
	ExpireAt time.Time `json:"expires_at"`
}

func newKVStore(filename string) *kvStore {
	return &kvStore{
		Store:    make(map[string]KvMapValue),
		Mu:       sync.RWMutex{},
		Filename: filename,
		OpCount:  0,
	}
}

func (kv *kvStore) Set(key string, val KvMapValue) {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	kv.Store[key] = val
	kv.OpCount++
	kv.AddLog("SET", "Success", key)
}

func (kv *kvStore) Get(key string) string {
	kv.Mu.RLock()
	defer kv.Mu.RUnlock()
	val, ok := kv.Store[key]
	if ok {
		kv.AddLog("READ", "Success", key)
		return val.Value
	}
	kv.AddLog("READ", "Failure", key)
	return ""
}

func (kv *kvStore) Delete(key string) {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	delete(kv.Store, key)
	kv.OpCount++
	kv.AddLog("DELETE", "Success", key)
}

func (kv *kvStore) Update(key, value string) bool {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()
	if val, exists := kv.Store[key]; exists {
		kv.Store[key] = KvMapValue{Value: value, ExpireAt: val.ExpireAt}
		kv.OpCount++
		kv.AddLog("UPDATE", "Success", key)
		return true
	}
	kv.AddLog("UPDATE", "Failure", key)
	return false

}

func (kv *kvStore) AddLog(operation, status, key string) {
	log := fmt.Sprintf("%s - %s operation - Key: %s - Status: %s\n",
		time.Now().Format(time.RFC3339),
		operation,
		key,
		status)

	f, err := os.OpenFile("logs.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	_, err = f.WriteString(log)
	if err != nil {
		fmt.Println("Error writing to log file:", err)
	}
}

func (kv *kvStore) LoadFromFile() error {
	data, err := os.ReadFile(kv.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File does not exist, starting with an empty Store.")
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &kv.Store)
}

func (kv *kvStore) SaveToFile() error {
	jsonData, err := json.Marshal(kv.Store)
	if err != nil {
		return err
	}
	err = os.WriteFile(kv.Filename, []byte(jsonData), 0644)
	if err != nil {
		return err
	}
	kv.OpCount = 0
	return nil
}

func (kv *kvStore) startPeriodicSave() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			kv.checkAndSave("Periodic")
		default:
			time.Sleep(1000 * time.Millisecond)
			kv.checkAndSave("OpCount")
		}
	}
}

func (kv *kvStore) checkAndSave(triggerType string) {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	if triggerType == "Periodic" || kv.OpCount >= 5 {
		err := kv.SaveToFile()
		if err != nil {
			fmt.Printf("Error saving to file: %v\n", err)
		}
	}
}

func (kv *kvStore) cleanUp() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for range ticker.C {
		for key, val := range kv.Store {
			if time.Now().After(val.ExpireAt) {
				kv.Delete(key)
				kv.AddLog("CLEANUP", "Expired", key)
			}
		}
	}
}
