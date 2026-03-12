package domain

type Preset struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	ServoPositions map[int]int `json:"servoPositions"` // JSON парсер Go сам переведет строковые ключи "1" в int
}

type HistoryItem struct {
	Date           int64       `json:"date"`
	Name           string      `json:"name"`
	ServoPositions map[int]int `json:"servoPositions"`
}