package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL = "http://localhost:8080/messages"
	// Number of concurrent requests to make for both POST and GET
	parallelRequests = 10
)

// Struct for the POST request body
type PostMessageRequest struct {
	Message string `json:"message"`
}

// Struct for the GET request response item
type GetMessageResponse struct {
	MessageID string `json:"messageId"`
	Message   string `json:"message"`
}

// Helper function to generate a random string of a given length
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 "
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// postMessage sends a new message to the server.
func postMessage(wg *sync.WaitGroup) {
	defer wg.Done()

	// 1. Generate a random message of length between 150 and 500 characters
	messageLength := rand.Intn(351) + 150 // Random length between 150 and 500
	message := generateRandomString(messageLength)

	reqBody := PostMessageRequest{Message: message}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("Error marshalling JSON for POST request: %v", err)
		return
	}

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error making POST request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("POST request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
		return
	}

	log.Printf("Successfully posted a message. Status: %s", resp.Status)
}

// getMessages fetches messages from the server and checks for duplicates.
func getMessages(wg *sync.WaitGroup, receivedIDs *sync.Map, duplicateCount *int64) {
	defer wg.Done()

	// 3. Prepare GET request with random parameters
	clientID := uuid.New().String()
	count := rand.Intn(5) + 1                   // Count between 1 and 5
	leaseSeconds := rand.Intn(21) + 10          // Lease time between 10 and 30 seconds

	// Construct the URL with query parameters
	url := fmt.Sprintf("%s?clientId=%s&count=%d&leaseExpiredAtInSeconds=%d", baseURL, clientID, count, leaseSeconds)

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error making GET request: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("GET request failed with status: %s, body: %s", resp.Status, string(bodyBytes))
		return
	}

	var messages []GetMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		log.Printf("Error decoding GET response: %v", err)
		return
	}

	log.Printf("GET request successful for clientId %s. Received %d messages.", clientID, len(messages))

	// 5. Check for duplicate message IDs
	for _, msg := range messages {
		// The LoadOrStore method is perfect for this: it loads a value if it exists,
		// or stores the given value if it doesn't. It returns the loaded value and
		// a boolean 'loaded' which is true if the value was loaded (i.e., it was a duplicate).
		_, loaded := receivedIDs.LoadOrStore(msg.MessageID, true)
		if loaded {
			// This messageId has been seen before.
			log.Printf("Error: Duplicate messageId received: %s", msg.MessageID)
			atomic.AddInt64(duplicateCount, 1)
		}
	}
}

func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	var wg sync.WaitGroup

	// --- Phase 1: Make 10 POST requests in parallel ---
	log.Println("--- Starting Phase 1: Posting Messages ---")
	wg.Add(parallelRequests)
	for i := 0; i < parallelRequests; i++ {
		go postMessage(&wg)
	}
	wg.Wait()
	log.Println("--- Finished Phase 1 ---")
	log.Println("") // Add a blank line for readability

	// --- Phase 2: Make 10 GET requests in parallel ---
	log.Println("--- Starting Phase 2: Getting Messages ---")
	// Use sync.Map for concurrent-safe map operations.
	// It's specifically optimized for the write-once, read-many-times pattern.
	var receivedMessageIDs sync.Map
	var duplicateCount int64 = 0

	wg.Add(parallelRequests)
	for i := 0; i < parallelRequests; i++ {
		// We add a small delay to make it more likely that different GET requests
		// might pull the same messages if the server logic allows it.
		time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)
		go getMessages(&wg, &receivedMessageIDs, &duplicateCount)
	}
	wg.Wait()
	log.Println("--- Finished Phase 2 ---")
	log.Println("")

	// --- Final Report ---
	fmt.Println("========================================")
	fmt.Printf("           Client Run Report          \n")
	fmt.Println("========================================")
	fmt.Printf("Total duplicate message IDs received: %d\n", duplicateCount)
	fmt.Println("========================================")
}

// --- MOCK SERVER FOR TESTING ---
// To test the client, you can run this mock server in a separate terminal.
// Save it as server.go and run `go run server.go`.

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        string `json:"messageId"`
	Content   string `json:"message"`
	LeasedBy  string
	LeaseExp  time.Time
}

var (
	messageStore = make(map[string]*Message)
	mu           sync.Mutex
)

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	switch r.Method {
	case http.MethodPost:
		var reqBody struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		msg := &Message{
			ID:      uuid.New().String(),
			Content: reqBody.Message,
		}
		messageStore[msg.ID] = msg
		log.Printf("Stored message: %s", msg.ID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(msg)

	case http.MethodGet:
		countStr := r.URL.Query().Get("count")
		leaseStr := r.URL.Query().Get("leaseExpiredAtInSeconds")
		clientID := r.URL.Query().Get("clientId")

		count, _ := strconv.Atoi(countStr)
		lease, _ := strconv.Atoi(leaseStr)

		var availableMessages []*Message
		for _, msg := range messageStore {
			if msg.LeasedBy == "" || time.Now().After(msg.LeaseExp) {
				availableMessages = append(availableMessages, msg)
			}
		}

		var responseMessages []*Message
		for i := 0; i < count && i < len(availableMessages); i++ {
			msg := availableMessages[i]
			msg.LeasedBy = clientID
			msg.LeaseExp = time.Now().Add(time.Duration(lease) * time.Second)
			responseMessages = append(responseMessages, msg)
		}
		log.Printf("Leased %d messages to client %s", len(responseMessages), clientID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responseMessages)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/messages", messagesHandler)
	log.Println("Mock server starting on :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
