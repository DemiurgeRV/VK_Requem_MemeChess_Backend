package ws

type BroadcastMessage struct {
	GameID  string
	Payload []byte
}

type subscription struct {
	client *Client
	gameID string
}

type Hub struct {
	clients      map[*Client]struct{}
	gameRooms    map[string]map[*Client]struct{}
	register     chan *Client
	unregister   chan *Client
	joinRoom     chan subscription
	leaveAllRoom chan *Client
	broadcast    chan BroadcastMessage
}

func NewHub() *Hub {
	return &Hub{
		clients:      make(map[*Client]struct{}),
		gameRooms:    make(map[string]map[*Client]struct{}),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		joinRoom:     make(chan subscription),
		leaveAllRoom: make(chan *Client),
		broadcast:    make(chan BroadcastMessage, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				h.removeClientFromAllRooms(client)
				close(client.send)
			}

		case sub := <-h.joinRoom:
			if _, ok := h.gameRooms[sub.gameID]; !ok {
				h.gameRooms[sub.gameID] = make(map[*Client]struct{})
			}
			h.gameRooms[sub.gameID][sub.client] = struct{}{}
			sub.client.gameIDs[sub.gameID] = struct{}{}

		case client := <-h.leaveAllRoom:
			h.removeClientFromAllRooms(client)

		case msg := <-h.broadcast:
			clients := h.gameRooms[msg.GameID]
			for client := range clients {
				select {
				case client.send <- msg.Payload:
				default:
					delete(h.clients, client)
					h.removeClientFromAllRooms(client)
					close(client.send)
				}
			}
		}
	}
}

func (h *Hub) removeClientFromAllRooms(client *Client) {
	for gameID := range client.gameIDs {
		if room, ok := h.gameRooms[gameID]; ok {
			delete(room, client)
			if len(room) == 0 {
				delete(h.gameRooms, gameID)
			}
		}
	}
	client.gameIDs = make(map[string]struct{})
}
