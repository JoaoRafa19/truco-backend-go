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

func returnError(w http.ResponseWriter, status int) {

	type _Message struct {
		Error string `json:"error"`
	}
	var errorMessage _Message
	w.WriteHeader(status)

	errorMessage = _Message{
		Error: http.StatusText(status),
	}

	data, _ := json.Marshal(errorMessage)
	w.Write(data)
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
		returnError(w, http.StatusInternalServerError)
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
		returnError(w, http.StatusInternalServerError)
		return
	}

	returnData(result, w)
}

func (h apiHandler) handleEnterGame(w http.ResponseWriter, r *http.Request) {

	gameId := chi.URLParam(r, "game_id")

	if gameId == "" {
		returnError(w, http.StatusBadRequest)
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
		returnError(w, http.StatusInternalServerError)
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
		returnError(w, http.StatusInternalServerError)
		return
	}

	type responseBody struct {
		Token string `json:"token"`
	}

	result, err := json.Marshal(responseBody{Token: tokenString})

	if err != nil {
		returnError(w, http.StatusInternalServerError)
		return
	}

	returnData(result, w)
}

func (h apiHandler) handleConnectToRoom(w http.ResponseWriter, r *http.Request) {
	gameId := chi.URLParam(r, "game_id")

	if gameId == "" {
		returnError(w, http.StatusBadRequest)
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
	players, err := h.q.GetRoomPlayers(r.Context(), roomID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusBadRequest)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	
	playerID, roomID, err := h.GetPlayerAndRoom(r, w, roomID)
	if err != nil {
		slog.Error("Erro ao validar informações", "error", err)
		returnError(w, http.StatusBadRequest)
		return
	} 

	playerIsInRoom := func() bool {
		for _, player := range players {

			if player.String() == playerID.String() {
				return true
			}
		}
		return false
	}()

	if !playerIsInRoom {
		w.WriteHeader(http.StatusUnauthorized)
		returnError(w, http.StatusUnauthorized)
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

	go h.readAndNotifyClients(c, r, playerID, roomID)

	<-ctx.Done()
}

func (h apiHandler) readAndNotifyClients(c *websocket.Conn, r *http.Request, playerID uuid.UUID, roomID uuid.UUID) error {
	for {
		msgType, msg, err := c.ReadMessage()

		if err != nil || msgType == -1 {
			h.mu.Lock()
			h.disconectClient(c, r, playerID)
			h.mu.Unlock()
			return err
		}

		if strings.Contains(string(msg), "echo:") {
			slog.Info("ROOMID", "room:", roomID.String())
			c.WriteMessage(websocket.TextMessage, []byte(msg))
		}

	}
}

func (apiHandler) GetPlayerAndRoom(r *http.Request, w http.ResponseWriter, roomID uuid.UUID) (uuid.UUID, uuid.UUID, error) {
	_, data, err := jwtauth.FromContext(r.Context())
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return uuid.Nil, uuid.Nil, err
	}

	fmt.Print(data)
	room := data["room_id"]
	playerID := data["player_id"]
	roomIDString, ok := room.(string)
	if !ok {
		slog.Error("Erro: o valor não é uma string")
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return uuid.Nil, uuid.Nil, err
	}
	playerIDString, ok := playerID.(string)
	if !ok {
		slog.Error("Erro: o valor não é uma string")
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return uuid.Nil, uuid.Nil, err
	}

	requestRoomID, err := uuid.Parse(roomIDString)
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return uuid.Nil, uuid.Nil, err
	}

	if requestRoomID != roomID {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return uuid.Nil, uuid.Nil, fmt.Errorf("player não está nessa sala")
	}
	var player uuid.UUID

	if player, err = uuid.Parse(playerIDString); err != nil {
		http.Error(w, "something went wrong", http.StatusBadRequest)
		return uuid.Nil, uuid.Nil, err
	}
	return player, roomID, nil
}

func (h apiHandler) getAllRooms(w http.ResponseWriter, r *http.Request) {

	rooms, err := h.q.GetAllRooms(r.Context())
	if err != nil {
		returnError(w, http.StatusInternalServerError)
		return
	}

	type roomsResponse struct {
		ID string `json:"id"`
	}

	var response []roomsResponse
	for _, room := range rooms {
		response = append(response, roomsResponse{ID: room.ID.String()})
	}
	result, err := json.Marshal(response)
	if err != nil {
		returnError(w, http.StatusInternalServerError)
		return
	}
	returnData(result, w)
}

func (h apiHandler) disconectClient(c *websocket.Conn, r *http.Request, playerId uuid.UUID) error {

	slog.Info("disconect client")
	room_id, err := h.q.RemovePlayerFromRoom(r.Context(), playerId)

	if err != nil {
		return err
	}

	delete(h.clients[room_id.String()], c)
	room, err := h.q.GetRoomPlayers(r.Context(), room_id)
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
