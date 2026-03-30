package ws

import (
	"encoding/json"
	"log"
	"sync"
)

type Message struct {
	Type    string      `json:"type"`
	OrderID string      `json:"order_id,omitempty"`
	Status  string      `json:"status,omitempty"`
	Reason  string      `json:"reason,omitempty"`
	Order   interface{} `json:"order,omitempty"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]bool // key -> set of clients
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Subscribe(key string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[key] == nil {
		h.clients[key] = make(map[*Client]bool)
	}
	h.clients[key][client] = true
}

func (h *Hub) Unsubscribe(key string, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[key]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, key)
		}
	}
}

func (h *Hub) UnsubscribeAll(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for key, clients := range h.clients {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, key)
		}
	}
}

func (h *Hub) Broadcast(key string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("ws: failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	clients := h.clients[key]
	h.mu.RUnlock()

	for client := range clients {
		select {
		case client.send <- data:
		default:
			// Client buffer full, skip
		}
	}
}

func (h *Hub) NotifyUser(userID string, msg Message) {
	h.Broadcast("user:"+userID, msg)
}

func (h *Hub) NotifyStore(storeID string, msg Message) {
	h.Broadcast("store:"+storeID, msg)
}

func (h *Hub) NotifyBranch(branchID string, msg Message) {
	h.Broadcast("branch:"+branchID, msg)
}

func (h *Hub) NotifyStoreAndBranch(storeID, branchID string, msg Message) {
	h.NotifyStore(storeID, msg)
	if branchID != "" {
		h.NotifyBranch(branchID, msg)
	}
}
