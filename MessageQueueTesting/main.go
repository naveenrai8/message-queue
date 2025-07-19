package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	baseURL           = "http://localhost:8080/messages"
	totalPostRequests = 1_000_000
	// Number of concurrent workers for POST requests.
	// Running 1 billion goroutines at once is not feasible.
	// A worker pool is a more robust approach.
	postWorkerCount = 100
	// Number of concurrent GET requests to make in each batch.
	getConcurrency = 10
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

// Generates a meaningful statement by combining phrases, with a max length of 450 bytes.
func generateRandomString() string {
	phrases := []string{
		"The quick brown fox jumps over the lazy dog.",
		"Microservices architecture enables scalable and resilient systems.",
		"Concurrent programming in Go is managed effectively with goroutines and channels.",
		"Continuous integration and continuous deployment pipelines automate software delivery.",
		"Observability is key to understanding complex distributed systems.",
		"REST APIs provide a flexible, lightweight way to integrate applications.",
		"Database indexing is crucial for improving query performance.",
		"The CAP theorem presents a fundamental choice between consistency, availability, and partition tolerance.",
		"Containerization with Docker and orchestration with Kubernetes have revolutionized application deployment.",
		"Writing clean, maintainable, and testable code is a hallmark of a professional software engineer.",
		"Agile methodologies help teams deliver value to their customers faster and with fewer headaches.",
		"A distributed ledger is a database that is consensually shared and synchronized across multiple sites, institutions, or geographies.",
		"The primary goal of a load balancer is to distribute traffic across multiple servers to ensure high availability and reliability.",
	}
	// Create a new random source for each goroutine to avoid contention on the global rand lock.
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	var builder strings.Builder
	maxLength := 450
	// Aim for a random length between 150 and 350 to keep some variability.
	targetMinLength := 150 + seededRand.Intn(201)

	for builder.Len() < targetMinLength {
		phrase := phrases[seededRand.Intn(len(phrases))]
		// Check if adding the new phrase (plus a space) would exceed the max length.
		if builder.Len()+len(phrase)+1 > maxLength {
			// If we haven't added anything yet and the first phrase is too long, just break.
			if builder.Len() == 0 {
				break
			}
			// Otherwise, stop before this phrase gets added.
			continue
		}
		builder.WriteString(phrase)
		builder.WriteString(" ")
	}

	// If the loop produced nothing (e.g., all phrases were too long), add the shortest one.
	if builder.Len() == 0 {
		builder.WriteString(phrases[0])
	}

	return strings.TrimSpace(builder.String())
}

// postWorker is a worker goroutine that processes POST requests from a channel.
func postWorker(workerID int, wg *sync.WaitGroup, jobs <-chan int) {
	defer wg.Done()
	for jobNum := range jobs {
		// 1. Generate a meaningful message of max length 450 bytes
		message := generateRandomString()

		reqBody := PostMessageRequest{Message: message}
		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			log.Printf("[Worker %d, Job %d] Error marshalling JSON: %v", workerID, jobNum, err)
			continue
		}

		resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("[Worker %d, Job %d] Error making POST request: %v", workerID, jobNum, err)
			continue
		}

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("[Worker %d, Job %d] POST request failed with status: %s, body: %s", workerID, jobNum, resp.Status, string(bodyBytes))
		} else {
			// Reduce log verbosity for the massive number of requests
			if jobNum%1000 == 0 {
				log.Printf("[Worker %d] Successfully posted message for job %d. Status: %s", workerID, jobNum, resp.Status)
			}
		}
		resp.Body.Close()
	}
}

// getMessages fetches messages from the server and checks for duplicates.
// It signals whether any messages were found in this request.
func getMessages(wg *sync.WaitGroup, receivedIDs *sync.Map, duplicateCount *int64, messagesFound *atomic.Bool) {
	defer wg.Done()

	// 3. Prepare GET request with random parameters
	clientID := uuid.New().String()
	count := rand.Intn(5) + 1          // Count between 1 and 5
	leaseSeconds := rand.Intn(21) + 10 // Lease time between 10 and 30 seconds

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

	if len(messages) > 0 {
		messagesFound.Store(true) // Signal that this batch found messages
		log.Printf("GET request successful for clientId %s. Received %d messages.", clientID, len(messages))
	}

	// 5. Check for duplicate message IDs
	for _, msg := range messages {
		_, loaded := receivedIDs.LoadOrStore(msg.MessageID, true)
		if loaded {
			log.Printf("Error: Duplicate messageId received: %s", msg.MessageID)
			atomic.AddInt64(duplicateCount, 1)
		}
	}
}

func main() {
	// Capture the start time to measure total execution time
	startTime := time.Now()

	// Seed the global random number generator once.
	rand.Seed(time.Now().UnixNano())

	// --- Phase 1: Make 1 Billion POST requests using a worker pool ---
	log.Println("--- Starting Phase 1: Posting 1 Billion Messages ---")
	var postWg sync.WaitGroup
	jobs := make(chan int, postWorkerCount)

	// Start workers
	for w := 1; w <= postWorkerCount; w++ {
		postWg.Add(1)
		go postWorker(w, &postWg, jobs)
	}

	// Send jobs to the workers
	for j := 1; j <= totalPostRequests; j++ {
		jobs <- j
	}
	close(jobs) // Close channel to signal workers to stop

	postWg.Wait() // Wait for all workers to finish
	log.Println("--- Finished Phase 1 ---")
	log.Println("")

	// --- Delay before starting GET requests ---
	log.Println("Waiting for 1 second before starting GET requests...")
	time.Sleep(1 * time.Second)
	log.Println("")

	// --- Phase 2: Make GET requests until no messages are returned ---
	log.Println("--- Starting Phase 2: Getting All Messages ---")
	var receivedMessageIDs sync.Map
	var duplicateCount int64 = 0

	for {
		var getWg sync.WaitGroup
		var messagesFoundInBatch atomic.Bool

		getWg.Add(getConcurrency)
		for i := 0; i < getConcurrency; i++ {
			go getMessages(&getWg, &receivedMessageIDs, &duplicateCount, &messagesFoundInBatch)
		}
		getWg.Wait()

		// If a full batch of concurrent GET requests found no messages, we can assume we're done.
		if !messagesFoundInBatch.Load() {
			log.Println("No messages found in the last batch. Ending GET phase.")
			break
		}
		// Optional: Add a small delay between batches to avoid hammering the server.
		time.Sleep(100 * time.Millisecond)
	}

	log.Println("--- Finished Phase 2 ---")
	log.Println("")

	// Calculate the total duration
	duration := time.Since(startTime)

	// --- Final Report ---
	fmt.Println("========================================")
	fmt.Printf("           Client Run Report          \n")
	fmt.Println("========================================")
	fmt.Printf("Total duplicate message IDs received: %d\n", duplicateCount)
	fmt.Printf("Total execution time: %s\n", duration)
	fmt.Println("========================================")
}

/*
// --- MOCK SERVER FOR TESTING ---
// To test the client, you can run this mock server in a separate terminal.
// Save it as server.go and run `go run server.go`.

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	// Using a slice and a map for more efficient iteration over available messages.
	messageStore []*Message
	messageMap   = make(map[string]*Message)
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
		messageStore = append(messageStore, msg)
		messageMap[msg.ID] = msg
		// Reduce log verbosity on the server side as well
		if len(messageStore)%1000 == 0 {
			log.Printf("Stored message count: %d", len(messageStore))
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(msg)

	case http.MethodGet:
		countStr := r.URL.Query().Get("count")
		leaseStr := r.URL.Query().Get("leaseExpiredAtInSeconds")
		clientID := r.URL.Query().Get("clientId")

		count, _ := strconv.Atoi(countStr)
		lease, _ := strconv.Atoi(leaseStr)

		var responseMessages []*Message
		// Iterate through messages to find available ones
		for _, msg := range messageStore {
			if len(responseMessages) >= count {
				break // We have collected enough messages
			}
			// Check if the message is not leased or the lease has expired
			if msg.LeasedBy == "" || time.Now().After(msg.LeaseExp) {
				msg.LeasedBy = clientID
				msg.LeaseExp = time.Now().Add(time.Duration(lease) * time.Second)
				responseMessages = append(responseMessages, msg)
			}
		}

		if len(responseMessages) > 0 {
			log.Printf("Leased %d messages to client %s", len(responseMessages), clientID)
		}
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
*/
