package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/JoaoRafa19/truco-backend-go/internal/store/pgstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx"
)

func (h apiHandler) handleEcho(w http.ResponseWriter, r *http.Request) {
	message := chi.URLParam(r, "message")
	fmt.Println(r.URL)
	fmt.Println("Hello", message)
	w.Write([]byte("echo " + message))
}

func returnError(handler string, err error, w http.ResponseWriter, status int) {
	slog.Info(handler, "error", err)
	switch status {
	case http.StatusInternalServerError:
		http.Error(w, "something went wrong", status)
	case http.StatusBadRequest:
		http.Error(w, "bad request", status)
	}
}

func returnData(result []byte, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write(result)
	if err != nil {
		slog.Error("failed to return response room", "error", err)
	}
}

func (h apiHandler) handleCreateGame(w http.ResponseWriter, r *http.Request) {

	game, err := h.q.CreateNewGame(r.Context())
	if err != nil {
		slog.Error("CreateGame", "error", err)
		returnError("handleCreateGame", err, w, http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
		Result    []byte `json:"result"`
		State     string `json:"state"`
		Round     int32  `json:"round"`
	}

	result, err := json.Marshal(
		responseBody{
			ID:        game.ID.String(),
			CreatedAt: game.CreatedAt.Time.String(),
			Result:    game.Result,
			State:     string(game.State),
			Round:     game.Round,
		})
	if err != nil {
		returnError("handleCreateGame", err, w, http.StatusInternalServerError)
		return
	}

	returnData(result, w)
}

func (h apiHandler) handleEnterGame(w http.ResponseWriter, r *http.Request) {

	gameId := chi.URLParam(r, "game_id")

	if gameId == "" {
		returnError("handleEnterGame", fmt.Errorf("missing game id"), w, http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(gameId)
	if err != nil {
		slog.Info("invalid room id")
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	type requestBody struct {
		Name string `json:"name"`
	}

	var body requestBody

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		returnError("handleEnterGame", err, w, http.StatusInternalServerError)
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	//TODO:validate auth if necessary

	playerID, err := h.q.CreatePlayer(r.Context(), pgstore.CreatePlayerParams{
		Name:   body.Name,
		RoomID: roomID,
	})

	if err != nil {
		slog.Info("unable to create player")
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	var responsePayload = map[string]interface{}{
		"player_id": playerID.String(),
		"room_id":   roomID.String(),
	}

	_, tokenString, err := h.tokenAuth.Encode(responsePayload)
	if err != nil {
		returnError("handleEnterGame", err, w, http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		Token string `json:"token"`
	}

	result, err := json.Marshal(responseBody{Token: tokenString})

	if err != nil {
		returnError("handleEnterGame", err, w, http.StatusInternalServerError)
		return
	}

	returnData(result, w)
}

func (h apiHandler) handleConnectToRoom(w http.ResponseWriter, r *http.Request) {
	gameId := chi.URLParam(r, "game_id")

	if gameId == "" {
		returnError("handleEnterGame", fmt.Errorf("missing game id"), w, http.StatusBadRequest)
		return
	}

	roomID, err := uuid.Parse(gameId)
	if err != nil {
		slog.Info("invalid room id")
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	gameRoom, err := h.q.GetRoom(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	fmt.Print(gameRoom)

	// verify if the user is on this room
	players, err := h.q.GetGamePlayers(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	_, data, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	fmt.Print(data)
	room := data["room_id"]
	playerID := data["player_id"]
	roomIDString, ok := room.(string)
	if !ok {
		slog.Error("Erro: o valor não é uma string")
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	playerIDString, ok := playerID.(string)
	if !ok {
		slog.Error("Erro: o valor não é uma string")
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	requestRoomID, err := uuid.Parse(roomIDString)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	if requestRoomID != roomID {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return
	}
	var player uuid.UUID

	if player, err = uuid.Parse(playerIDString); err != nil {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return
	}

	playerIsInRoom := func() bool {
		for _, player := range players {
			if player.String() == playerIDString {
				return true
			}
		}
		return false
	}()

	if !playerIsInRoom {
		http.Error(w, "something went wrong", http.StatusUnauthorized)
		return
	}

	c, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Warn("failed to upgrade connection", "error", err)
		http.Error(w, "failed to upgrade to ws connection", http.StatusBadRequest)
		return
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(r.Context())

	// Trava o mutex para fazer alteração no map de conexões
	h.mu.Lock()

	if _, ok := h.clients[roomID.String()]; !ok {
		h.clients[roomID.String()] = make(map[*websocket.Conn]context.CancelFunc)
	}

	slog.Info("new client", "room", roomID.String())
	h.clients[roomID.String()][c] = cancel

	h.mu.Unlock()

	for {
		msgType, msg, err := c.ReadMessage()

		if msgType == websocket.CloseAbnormalClosure {
			break
		}
		if err != nil {
			if errors.Is(err, &websocket.CloseError{}) {
				h.mu.Lock()
				h.disconectClient(c, r, player)
				h.mu.Unlock()
				break
			}
		}

		if msgType == -1 {
			h.mu.Lock()
			h.disconectClient(c, r, player)
			h.mu.Unlock()
			return
		}

		if err != nil {
			slog.Error("erro", "error", err)
			continue
		}

		if strings.Contains(string(msg), "echo:") {
			slog.Info(string(msg))
			c.WriteMessage(websocket.TextMessage, []byte("hi 2"))
		}

	}
	<-ctx.Done()
}

func (h apiHandler) disconectClient(c *websocket.Conn, r *http.Request, playerId uuid.UUID) error {

	slog.Info("disconect client")
	room_id, err := h.q.RemovePlayerFromRoomReturningRoom(r.Context(), playerId)

	if err != nil {
		return err
	}

	delete(h.clients[room_id.String()], c)
	

	room, err := h.q.GetGamePlayers(r.Context(), room_id)

	if err != nil {
		slog.Error("erro ao terminar jogo", "error", err)
		return err
	}

	if len(room) == 0 {
		delete(h.clients, room_id.String())
		id, err := h.q.DeleteGameRoom(r.Context(), room_id)
		if err != nil {
			slog.Error("erro ao terminar jogo", "error", err, "id", id)
			return err
		}
	}
	return c.Close()
}
