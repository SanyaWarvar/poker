package user

import "math/rand"

var sectors = []int{
	10000,
	1000,
	5000,
	1000,
	250,
	1000,
	2000,
	1000,
	750,
	1,
	1500,
	1000,
	500,
	1000,
	3000,
	1000,
	100,
	1000,
}

// DailyReward
// @Schema
type DailyReward struct {
	Amount      int `json:"amount" example:"1000"`
	SectorIndex int `json:"sector" example:"1"`
}

func SpinWheel() DailyReward {
	angle := rand.Intn(360) + 1
	sectorAngle := 360 / len(sectors)
	skipSectors := angle / sectorAngle
	targetSector := (0 + skipSectors) % len(sectors)
	return DailyReward{Amount: sectors[targetSector], SectorIndex: targetSector}
}
