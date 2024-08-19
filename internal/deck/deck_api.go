package deck

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Card struct {
	Code     string `json:"code"`
	ImageURL string `json:"image"`
	Value    string `json:"value"`
	Suit     string `json:"suit"`
}

type Deck struct {
	DeckID    string `json:"deck_id"`
	Shuffled  bool   `json:"shuffled"`
	Remaining int    `json:"remaining"`
}

type CardsApiRespnse struct {
	Success   bool   `json:"success"`
	DeckID    string `json:"deck_id"`
	Remaining int16  `json:"remaining"`
	Cards     []Card `json:"cards"`
}

func (d *Deck) DrawCards(numberOfCards int) ([]Card, error) {
	response, err := http.Get(fmt.Sprintf("https://www.deckofcardsapi.com/api/deck/%s/draw/?count=%d", d.DeckID, numberOfCards))

	if err != nil {
		return nil, fmt.Errorf("unable to get cards")
	}

	

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get cards")
	}

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("unable to get cards")
	}

	var responseBody CardsApiRespnse
	if err := json.Unmarshal(responseData, &responseBody); err != nil {
		return nil, fmt.Errorf("unable to get cards")
	}

	return responseBody.Cards, nil
}

func (d *Deck) GetDeckState(deckID string) (Deck, error) {
	response, err := http.Get(fmt.Sprintf("https://www.deckofcardsapi.com/api/deck/%s", deckID ))

	if err != nil {
		return Deck{}, fmt.Errorf("unable to get a new game")
	}

	type cardsApiRespnse struct {
		Success   bool   `json:"success"`
		DeckID    string `json:"deck_id"`
		Remaining int16  `json:"remaining"`
		Shuffled  bool   `json:"shuffled"`
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return Deck{}, fmt.Errorf("unable to get a new game")
	}

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		return Deck{}, fmt.Errorf("unable to get a new game")
	}

	var responseBody cardsApiRespnse
	if err := json.Unmarshal(responseData, &responseBody); err != nil {
		return Deck{}, fmt.Errorf("unable to get a new game")
	}

	return Deck{
		DeckID: responseBody.DeckID,
		Shuffled: responseBody.Shuffled,
		Remaining: int(responseBody.Remaining),
	}, nil
}

func CreateDeck() (*Deck, error) {
	response, err := http.Get("https://www.deckofcardsapi.com/api/deck/new/shuffle/?cards=AS,KS,QS,JS,7S,6S,5S,4S,3S,2S,AD,KD,QD,JD,7D,6D,5D,4D,3D,2D,AC,KC,QC,JC,7C,6C,5C,4C,3C,2C,AH,KH,QH,JH,7H,6H,5H,4H,3H,2H")

	if err != nil {
		return nil, fmt.Errorf("unable to get a new game")
	}

	type cardsApiRespnse struct {
		Success   bool   `json:"success"`
		DeckID    string `json:"deck_id"`
		Remaining int16  `json:"remaining"`
		Shuffled  bool   `json:"shuffled"`
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get a new game")
	}

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("unable to get a new game")
	}

	var responseBody cardsApiRespnse
	if err := json.Unmarshal(responseData, &responseBody); err != nil {
		return nil, fmt.Errorf("unable to get a new game")
	}

	return &Deck{
		DeckID: responseBody.DeckID,
		Shuffled: responseBody.Shuffled,
		Remaining: int(responseBody.Remaining),
	}, nil
}
