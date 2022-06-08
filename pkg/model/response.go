package model

const (
	StaleValue = 0
	FreshValue = 1
)

type Response struct {
	Header     map[string][]string
	Body       []byte
	Status     int
	StaleValue uint8 `json:"-"`
}

func (r Response) IsStale() bool {
	return r.StaleValue == StaleValue
}
