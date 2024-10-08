package api

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/JoaoRafa19/truco-backend-go/internal/deck"
	"github.com/JoaoRafa19/truco-backend-go/internal/store/pgstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth/v5"
	"github.com/gorilla/websocket"
)

type Room struct {
	connections map[*websocket.Conn]context.CancelFunc
	deck        deck.Deck
}

type apiHandler struct {
	q         *pgstore.Queries
	r         *chi.Mux
	tokenAuth *jwtauth.JWTAuth
	upgrader  websocket.Upgrader
	mu        *sync.Mutex
	clients   map[string]Room
}

func NewHandler(q *pgstore.Queries) http.Handler {
	h := apiHandler{
		q:         q,
		tokenAuth: jwtauth.New("HS256", []byte("go-truco"), nil),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},

		mu:      &sync.Mutex{},
		clients: make(map[string]Room),
	}

	r := chi.NewRouter()

	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/echo/{message}/teste", h.handleEcho)

	r.Route("/game", func(r chi.Router) {
		r.Post("/", h.handleCreateGame)
		r.Get("/", h.getAllRooms)
		r.Patch("/{game_id}/enter", h.handleEnterGame)
		r.Route("/{game_id}/", func(r chi.Router) {
			r.Use(jwtauth.Verifier(h.tokenAuth))
			r.Use(jwtauth.Authenticator(h.tokenAuth))
			r.Get("/connect", h.handleConnectToRoom) //ws
			r.Get("/", h.getGameState)
			r.Patch("/start", h.handleStartGame)
		})
	})

	h.r = r

	return h
}

func (h apiHandler) notifyClients(event []byte, roomId string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, ok := h.clients[roomId]
	if !ok || len(room.connections) == 0 {
		return
	}

	for conn, cancel := range room.connections {
		if err := conn.WriteMessage(websocket.BinaryMessage, event); err != nil {
			slog.Error("failed to send message to client", "error", err)
			cancel()
		}
	}
}

func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}
