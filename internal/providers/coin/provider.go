package coin_provider

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/aspiration-labs/pyggpot/internal/models"
	coin_service "github.com/aspiration-labs/pyggpot/rpc/go/coin"
	"github.com/twitchtv/twirp"
)

type ID int32

func init() {
	rand.Seed(time.Now().UnixNano())
}
type coinServer struct{
	DB *sql.DB
}

func New(db *sql.DB) *coinServer {
	return &coinServer{
		DB: db,
	}
}

func (s *coinServer) AddCoins(ctx context.Context, request *coin_service.AddCoinsRequest) (*coin_service.CoinsListResponse, error) {
	if err := request.Validate(); err != nil {
		return nil, twirp.InvalidArgumentError(err.Error(), "")
	}

	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}
	for _, coin := range request.Coins {
		fmt.Println(coin)
		newCoin := models.Coin{
			PotID: request.PotId,
			Denomination: int32(coin.Kind),
			CoinCount: coin.Count,
		}
		err = newCoin.Save(tx)
		if err != nil {
			return nil, twirp.InvalidArgumentError(err.Error(), "")
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	return &coin_service.CoinsListResponse{
		Coins: request.Coins,
	}, nil
}

func (s *coinServer) RemoveCoins(ctx context.Context, request *coin_service.RemoveCoinsRequest) (*coin_service.CoinsListResponse, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, twirp.InternalError(err.Error())
	}

	coins, err := models.CoinsInPotsByPot_id(tx, int(request.PotId))
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	coinsSlippedOut := s.getCoinsFromPot(coins, request.Count)

	if len(coinsSlippedOut) == 0 {
		return nil, twirp.InvalidArgumentError("not enough coins in pot", "")
	}

	var coinsTakenOut []*coin_service.Coins

	coinsTakenOutByDenomination := make(map[int32]*coin_service.Coins)

	for k, v := range coinsSlippedOut {
		coinToUpdate, err := models.CoinByID(tx, int32(k))
		if err != nil {
			return nil, twirp.NotFoundError(err.Error())
		}

		coinToUpdate.CoinCount -= v
		err = coinToUpdate.Update(tx)
		if err != nil {
			return nil, twirp.NotFoundError(err.Error())
		}

		coinTakenOut, ok := coinsTakenOutByDenomination[coinToUpdate.Denomination]
		if !ok {
			coinsTakenOutByDenomination[coinToUpdate.Denomination] = &coin_service.Coins{
				Kind: coin_service.Coins_Kind(coinToUpdate.Denomination),
				Count: v,
			}
			continue
		}
		coinTakenOut.Count += v
	}
	err = tx.Commit()
	if err != nil {
		return nil, twirp.NotFoundError(err.Error())
	}

	for _, c := range coinsTakenOutByDenomination {
		coinsTakenOut = append(coinsTakenOut, c)
	}

	return &coin_service.CoinsListResponse{
		Coins: coinsTakenOut,
	}, nil
}

// generates returns an index where tha denomination is stored
func (s *coinServer) generateRandomIndex(totalCoinsByDenomination int) int {
	return rand.Intn(totalCoinsByDenomination)
}

func remove(slice []*models.CoinsInPot, s int) []*models.CoinsInPot {
	return append(slice[:s], slice[s+1:]...)
}

// Simulate shaking a piggy bank upside down, so gets coins that are in the pot
// the type (denomination) of the coin is random, based on the proportion of that coin type in the pot
// returns a map where the key is the id and the value represents how many coins of that denomination are taken
func (s *coinServer) getCoinsFromPot(coinsInPot []*models.CoinsInPot, count int32) map[ID]int32 {
	coinsInPotFiltered := make([]*models.CoinsInPot, 0, len(coinsInPot))
	var totalCoins int32
	for _, c := range coinsInPot {
		if c.CoinCount > 0 {
			totalCoins += c.CoinCount
			coinsInPotFiltered = append(coinsInPotFiltered, c)
		}
	}
	var counter int32
	coinsToRemove := make(map[ID]int32, 0)
	for counter < count && totalCoins >= count {
		counter++
		index := s.generateRandomIndex(len(coinsInPotFiltered))
		coin := coinsInPotFiltered[index]
		coin.CoinCount -= 1
		if coin.CoinCount == 0 {
			coinsInPotFiltered = remove(coinsInPotFiltered, index)
		}
		coinToRemove, ok := coinsToRemove[ID(coin.ID)]
		if ok {
			coinsToRemove[ID(coin.ID)] = coinToRemove + 1
			continue
		}
		coinsToRemove[ID(coin.ID)] = 1
	}
	return coinsToRemove
}

