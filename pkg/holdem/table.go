package holdem

import (
	"errors"
	"fmt"
	"math/rand"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMaxPlayers       = errors.New("count of players reached max value")
	ErrGameStarted      = errors.New("this game already started")
	ErrGameNotStarted   = errors.New("this game not started")
	ErrNotEnoughCards   = errors.New("not enough card in deck")
	ErrNotYourTurn      = errors.New("not your turn t")
	ErrPlayerIsFold     = errors.New("player already fold his cards")
	ErrCantCheck        = errors.New("you cant check")
	ErrCantRaise        = errors.New("raise must exceed the current bet by at least two times")
	ErrNotEnoughMoney   = errors.New("not enough money for this  action")
	ErrUnexpectedAction = errors.New("unexpected action")
	ErrPlayerNotFound   = errors.New("player not found")
)

type IPokerTable interface {
	StartGame() error
	AddObserver(o IObserver)
	AddPlayer(player IPlayer) error
	RemovePlayer(playerId string) error
	MakeMove(playerId, action string, amount int) error
	GetConfig() *TableConfig
	CheckPlayer(playerId string) bool
}

// TableConfig
// @Schema
type TableConfig struct {
	TableId           uuid.UUID     `json:"lobby_id"`
	BlindIncreaseTime time.Duration `json:"blind_increase_time"`
	LastBlindIncrease time.Time     `json:"last_blind_increase_time"`
	MaxPlayers        int           `json:"max_players"`
	MinPlayers        int           `json:"min_players_to_start"`
	CurrentPlayers    int           `json:"current_players_count"`
	EnterAfterStart   bool          `json:"cache_game"` //true = cache game. false = sit n go
	SmallBlind        int           `json:"small_blind"`
	Ante              int           `json:"ante"`
	BankAmount        int           `json:"bank_amount"`
	Seed              int64         `json:"-"`
}

// TODO add timeout for 1 move and time bank
type TableMeta struct {
	DealerIndex    int
	PlayerTurnInd  int
	CurrentBet     int
	CommunityCards []Card
	PlayersOrder   []string
	Players        map[string]IPlayer
	Query          map[string]IPlayer
	Pots           []Pot
	Deck           []Card
	CurrentRound   int
	GameStarted    bool
}

type PokerTable struct {
	observers []IObserver
	mu        sync.Mutex
	Config    *TableConfig
	Meta      *TableMeta
}

func NewTableConfig(
	BlindIncreaseTime time.Duration,
	maxPlayers, minPlayers, smallBlind, ante, bankAmount int,
	enterAfteStart bool,
	seed int64) *TableConfig {
	return &TableConfig{
		BlindIncreaseTime: BlindIncreaseTime,
		LastBlindIncrease: time.Now(),
		MaxPlayers:        maxPlayers,
		MinPlayers:        minPlayers,
		CurrentPlayers:    0,
		EnterAfterStart:   enterAfteStart,
		SmallBlind:        smallBlind,
		Ante:              ante,
		Seed:              seed,
		BankAmount:        bankAmount,
		TableId:           uuid.New(),
	}
}

func NewTableMeta() *TableMeta {
	return &TableMeta{
		DealerIndex:    0,
		PlayerTurnInd:  0,
		CurrentBet:     0,
		CommunityCards: []Card{},
		PlayersOrder:   make([]string, 0, 10),
		Players:        make(map[string]IPlayer),
		Query:          make(map[string]IPlayer),
		Pots:           []Pot{},
		Deck:           []Card{},
		CurrentRound:   -1,
		GameStarted:    false,
	}
}

func NewPokerTable(config *TableConfig) *PokerTable {
	return &PokerTable{
		observers: []IObserver{},
		mu:        sync.Mutex{},
		Config:    config,
		Meta:      NewTableMeta(),
	}
}

func (t *PokerTable) GetConfig() *TableConfig {
	return t.Config
}

func (t *PokerTable) CheckPlayer(playerId string) bool {
	for k := range t.Meta.Players {
		if k == playerId {
			return true
		}
	}
	for k := range t.Meta.Query {
		if k == playerId {
			return true
		}
	}
	return false
}

func (m *TableMeta) refreshDeck(seed int64) {
	m.Deck = GetStandardDeck()
	var r *rand.Rand
	if seed != 0 {
		r = rand.New(rand.NewSource(seed))
	}
	r.Shuffle(len(m.Deck), func(i, j int) {
		m.Deck[i], m.Deck[j] = m.Deck[j], m.Deck[i]
	})
}

func (m *TableMeta) addPlayerInGame(p IPlayer, bankAmount int) {
	m.Players[p.GetId()] = p
	if bankAmount != 0 {
		p.SetBalance(bankAmount)
	}
}

//TODO remove player

func (m *TableMeta) addPlayerInQuery(p IPlayer) {
	m.Query[p.GetId()] = p
}

func (t *PokerTable) AddObserver(obs IObserver) {
	t.observers = append(t.observers, obs)
}

func (t *PokerTable) NotifyObservers(event string) {
	for _, obs := range t.observers {
		obs.Update(event)
	}
}

func (t *PokerTable) AddPlayer(p IPlayer) error {
	if t.Meta.GameStarted && !t.Config.EnterAfterStart {
		return ErrGameStarted
	}

	if t.Config.MaxPlayers <= len(t.Meta.Players)+len(t.Meta.Query)+1 {
		return ErrMaxPlayers
	}

	if t.Meta.GameStarted {
		t.Meta.addPlayerInQuery(p)
	} else {
		t.Meta.addPlayerInGame(p, t.Config.BankAmount)
		t.Meta.PlayersOrder = append(t.Meta.PlayersOrder, p.GetId())
	}
	t.Config.CurrentPlayers += 1
	t.NotifyObservers(fmt.Sprintf("Player %s enter the game", p.GetId()))
	return nil
}

func (t *PokerTable) enterPlayersFromQuery() {
	for k, v := range t.Meta.Query {
		t.Meta.Players[k] = v
		t.Meta.PlayersOrder = append(t.Meta.PlayersOrder, k)
	}
}

func (t *PokerTable) StartGame() error {
	if t.Meta.GameStarted {
		return ErrGameStarted
	}
	t.Meta.GameStarted = true
	t.Meta.CurrentRound = -1
	t.Meta.refreshDeck(t.Config.Seed)
	t.NotifyObservers("Game started")
	t.NewRound()
	return nil
}

func (t *PokerTable) NewRound() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	t.createPots()
	t.Meta.CurrentRound += 1
	t.Meta.CurrentBet = 0
	t.NotifyObservers(fmt.Sprintf("New round started. Current round: %d", t.Meta.CurrentRound))

	refreshPlayers(t.Meta.Players, t.Meta.CurrentRound == 4)
	switch t.Meta.CurrentRound {
	case 0: //pre flop
		t.enterPlayersFromQuery()
		t.betAnte()
		for _, k := range t.Meta.PlayersOrder {
			cards, _ := t.drawCard(2)
			t.Meta.Players[k].SetHand(Hand{[2]Card{cards[0], cards[1]}})
			t.NotifyObservers(fmt.Sprintf("Player %s get cards: %v", t.Meta.Players[k].GetId(), cards))
		}
		t.choiceDealer()
		t.betBlinds()
	case 1: // flop
		t.Meta.CommunityCards, _ = t.drawCard(3)
		t.NotifyObservers(fmt.Sprintf("Community cards: %v", t.Meta.CommunityCards))
		t.Meta.PlayerTurnInd = (t.Meta.DealerIndex + 1) % len(t.Meta.PlayersOrder)

	case 2: // turn
		cards, _ := t.drawCard(1)
		t.Meta.CommunityCards = append(t.Meta.CommunityCards, cards...)
		t.NotifyObservers(fmt.Sprintf("Community cards: %v", t.Meta.CommunityCards))

	case 3: // river
		cards, _ := t.drawCard(1)
		t.Meta.CommunityCards = append(t.Meta.CommunityCards, cards...)
		t.NotifyObservers(fmt.Sprintf("Community cards: %v", t.Meta.CommunityCards))

	case 4: // determinate winner
		t.PayMoney()
		t.Config.updateSeed()
		t.Meta.GameStarted = false
		t.Meta.CurrentRound = -1
		t.Meta.Pots = t.Meta.Pots[:0]
	}
	t.choiceFirstMovePlayer()

	return nil
}

func (cfg *TableConfig) updateSeed() {
	if cfg.Seed != 0 {
		r := rand.New(rand.NewSource(cfg.Seed))
		for {
			newSeed := r.Int63()
			if newSeed == 0 {
				continue
			}
			cfg.Seed = newSeed
			break
		}
	}
}

func (t *PokerTable) PayMoney() {
	for ind, pot := range t.Meta.Pots {
		applicants := make(map[string]IPlayer)
		for _, k := range pot.Applicants {
			p := t.Meta.Players[k]
			if p.GetFold() { // если игрок сбросил то он не претендует на банк
				continue
			}
			applicants[k] = p
		}
		winners, _ := DeterminateWinner(t.Meta.CommunityCards, applicants)
		winAmount := pot.Amount / len(winners)
		for _, winner := range winners {
			t.Meta.Players[winner].ChangeBalance(winAmount)
		}
		t.NotifyObservers(fmt.Sprintf("Winners of pot %.2d with %d amount: %v", ind+1, winAmount, winners))
		if winAmount*len(winners) == pot.Amount {
			continue
		}
		counter := pot.Amount - winAmount*len(winners)
		for i := 1; counter > 0; i++ {
			targetPlayer := t.Meta.PlayersOrder[(t.Meta.DealerIndex+i)%len(t.Meta.Players)]
			if t.Meta.Players[targetPlayer].GetFold() || !slices.Contains(winners, t.Meta.Players[targetPlayer].GetId()) {
				continue
			}
			t.Meta.Players[targetPlayer].ChangeBalance(1)
			counter--
		}
	}
}

func (t *PokerTable) createPots() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}

	pots := CreatePots(t.Meta.Players)
	t.Meta.Pots = append(t.Meta.Pots, pots...)

	return nil
}

func (t *PokerTable) RemovePlayer(playerId string) error {
	_, ok1 := t.Meta.Players[playerId]
	_, ok2 := t.Meta.Query[playerId]
	if !(ok1 || ok2) {
		return ErrPlayerNotFound
	}
	if ok2 {
		delete(t.Meta.Query, playerId)
		return nil
	}
	delete(t.Meta.Players, playerId)
	t.Config.CurrentPlayers -= 1
	ind := slices.Index(t.Meta.PlayersOrder, playerId)
	t.Meta.PlayersOrder = append(t.Meta.PlayersOrder[:ind], t.Meta.PlayersOrder[ind+1:]...)
	return nil
}

func (t *PokerTable) betAnte() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	//TODO check if not 0 round
	toRemove := []string{}
	for k, v := range t.Meta.Players {
		if v.GetBalance() < t.Config.Ante {
			v.GetFold()
			t.NotifyObservers(fmt.Sprintf("Player %s cant bet ante", k))
			toRemove = append(toRemove, k)
		}
	}
	for _, id := range toRemove {
		t.RemovePlayer(id)
	}

	t.Meta.Pots = append(t.Meta.Pots, Pot{Amount: t.Config.Ante * len(t.Meta.Players), Applicants: t.Meta.PlayersOrder})
	t.NotifyObservers(fmt.Sprintf("Get ante: %d", t.Config.Ante*len(t.Meta.Players)))
	return nil
}

func (t *PokerTable) betBlinds() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	var smallBlindPlayer, bigBlindPlayer string
	if len(t.Meta.PlayersOrder) > 2 {
		smallBlindPlayer = t.Meta.PlayersOrder[(t.Meta.DealerIndex+1)%len(t.Meta.PlayersOrder)]
		bigBlindPlayer = t.Meta.PlayersOrder[(t.Meta.DealerIndex+2)%len(t.Meta.PlayersOrder)]
	} else {
		smallBlindPlayer = t.Meta.PlayersOrder[t.Meta.DealerIndex] // дилер ставит малый блайнд в хендз апе
		bigBlindPlayer = t.Meta.PlayersOrder[(t.Meta.DealerIndex+1)%len(t.Meta.PlayersOrder)]
	}

	smallBlindPlayerBet := min(t.Config.SmallBlind, t.Meta.Players[smallBlindPlayer].GetBalance())
	t.Meta.Players[smallBlindPlayer].ChangeBalance(-smallBlindPlayerBet)
	t.Meta.Players[smallBlindPlayer].SetLastBet(smallBlindPlayerBet)
	t.NotifyObservers(fmt.Sprintf("Player %s bet %d as small blind", smallBlindPlayer, smallBlindPlayerBet))

	bigBlindPlayerBet := min(t.Config.SmallBlind*2, t.Meta.Players[bigBlindPlayer].GetBalance())
	t.Meta.Players[bigBlindPlayer].ChangeBalance(-bigBlindPlayerBet)
	t.Meta.Players[bigBlindPlayer].SetLastBet(bigBlindPlayerBet)
	t.NotifyObservers(fmt.Sprintf("Player %s bet %d as big blind", bigBlindPlayer, bigBlindPlayerBet))
	t.Meta.CurrentBet = max(bigBlindPlayerBet, smallBlindPlayerBet)
	return nil
}

func (t *PokerTable) getNextPlayer() {
	for i := 1; i < len(t.Meta.PlayersOrder); i++ {
		nextIndex := (t.Meta.PlayerTurnInd + i) % len(t.Meta.PlayersOrder)
		nextPlayer := t.Meta.PlayersOrder[nextIndex]
		if !t.Meta.Players[nextPlayer].GetFold() && !t.Meta.Players[nextPlayer].GetReadyStatus() {
			t.Meta.PlayerTurnInd = nextIndex
			t.NotifyObservers(fmt.Sprintf("Next move expect from %s player", nextPlayer))
			return
		}
	}
}

func (t *PokerTable) choiceDealer() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	t.Meta.DealerIndex = (t.Meta.DealerIndex + 1) % len(t.Meta.PlayersOrder)
	t.NotifyObservers(fmt.Sprintf("dealer is %s", t.Meta.PlayersOrder[t.Meta.DealerIndex]))
	return nil
}

func (t *PokerTable) drawCard(n int) ([]Card, error) {
	output := make([]Card, 0, n)
	if len(t.Meta.Deck) < n {
		return output, ErrNotEnoughCards
	}
	output = append(output, t.Meta.Deck[:n]...)
	t.Meta.Deck = t.Meta.Deck[n:]
	return output, nil
}

func (t *PokerTable) choiceFirstMovePlayer() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	if t.Meta.CurrentRound == 0 { //utg
		t.Meta.PlayerTurnInd = (t.Meta.DealerIndex + 3) % len(t.Meta.PlayersOrder)
	} else {
		t.Meta.PlayerTurnInd = (t.Meta.DealerIndex + 1) % len(t.Meta.PlayersOrder)
	}
	return nil
}

func (t *PokerTable) MakeMove(playerId, action string, amount int) error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}

	if t.Meta.PlayersOrder[t.Meta.PlayerTurnInd] != playerId {
		return ErrNotYourTurn
	}

	if t.Meta.Players[playerId].GetFold() {
		return ErrPlayerIsFold
	}

	switch action {
	case "check":
		t.handleCheck(playerId)
	case "raise":
		t.handleRaise(playerId, amount)
	case "call":
		t.handleCall(playerId)
	case "fold":
		t.handleFold(playerId)
	default:
		return ErrUnexpectedAction
	}
	t.Meta.Players[playerId].SetStatus(true)
	t.getNextPlayer()
	if t.checkReady() {
		t.NewRound()
	} else {
		t.notifyNext()
	}
	return nil
}

func (t *PokerTable) notifyNext() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	pId := t.Meta.PlayersOrder[t.Meta.PlayerTurnInd]
	if t.Meta.CurrentBet != 0 {
		t.NotifyObservers(fmt.Sprintf("player %s can do call with %d", pId, t.Meta.CurrentBet))
	} else {
		t.NotifyObservers(fmt.Sprintf("player %s can do check", pId))
	}
	return nil
}

func (t *PokerTable) checkReady() bool {
	if !t.Meta.GameStarted {
		return false
	}
	for _, v := range t.Meta.Players {
		if (!v.GetFold() && !v.GetReadyStatus()) || v.GetBalance() == 0 {
			return false
		}
	}
	return true
}

func (t *PokerTable) handleCheck(playerId string) error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	if t.Meta.Players[playerId].GetFold() {
		return ErrPlayerIsFold
	}

	if t.Meta.CurrentBet != 0 {
		return ErrCantCheck
	}
	t.Meta.Players[playerId].SetStatus(true)
	t.NotifyObservers(fmt.Sprintf("Player %s do check", playerId))
	return nil
}

func (t *PokerTable) handleFold(playerId string) error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}

	t.Meta.Players[playerId].SetStatus(true)
	t.Meta.Players[playerId].SetFold(true)
	t.NotifyObservers(fmt.Sprintf("Player %s do fold", playerId))
	return nil
}

func (t *PokerTable) handleRaise(playerId string, amount int) error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}
	if t.Meta.Players[playerId].GetFold() {
		return ErrPlayerIsFold
	}
	if !(amount > t.Meta.CurrentBet*2 && amount > t.Meta.Players[playerId].GetLastBet() && amount > 0) {
		return ErrCantRaise
	}
	delta := amount - t.Meta.Players[playerId].GetLastBet()
	if delta > t.Meta.Players[playerId].GetBalance() {
		return ErrNotEnoughMoney
	}
	t.resetPlayersStatus()
	t.Meta.Players[playerId].SetLastBet(amount)
	t.Meta.Players[playerId].ChangeBalance(-delta)
	t.Meta.Players[playerId].SetStatus(true)
	t.Meta.CurrentBet = amount

	t.NotifyObservers(fmt.Sprintf("Player %s do raise with %d amount", playerId, amount))
	return nil
}

func (t *PokerTable) resetPlayersStatus() error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}

	for k, v := range t.Meta.Players {
		if !v.GetFold() {
			t.Meta.Players[k].SetStatus(false)
		}
	}
	return nil
}

func refreshPlayers(players map[string]IPlayer, fold bool) error {
	for _, v := range players {
		if fold {
			v.SetFold(false)
		}
		v.SetLastBet(0)
		v.SetStatus(false)
	}
	return nil
}

func (t *PokerTable) handleCall(playerId string) error {
	if !t.Meta.GameStarted {
		return ErrGameNotStarted
	}

	if t.Meta.CurrentBet == 0 {
		return t.handleCheck(playerId)
	}
	if t.Meta.Players[playerId].GetFold() {
		return ErrPlayerIsFold
	}
	needToBet := t.Meta.CurrentBet - t.Meta.Players[playerId].GetLastBet()
	possibleBet := min(needToBet, t.Meta.Players[playerId].GetBalance())

	t.Meta.Players[playerId].ChangeBalance(-possibleBet)
	t.Meta.Players[playerId].SetStatus(true)
	if t.Meta.Players[playerId].GetBalance() > 0 {
		t.Meta.Players[playerId].SetLastBet(t.Meta.CurrentBet)
	} else {
		t.Meta.Players[playerId].SetLastBet(possibleBet)
	}

	t.NotifyObservers(fmt.Sprintf("Player %s do call with %d amount", playerId, t.Meta.CurrentBet))
	return nil
}
