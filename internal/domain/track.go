package domain

type TrackRequest struct {
	Status  float64 `json:"status" validate:"required"`
	Url     string  `json:"url"`
	Key     string  `json:"key" validate:"required"`
	Actions []struct {
		Type   int    `json:"type"`
		UserId string `json:"userid"`
	} `json:"actions"`
	Token string `json:"token"`
}

func (x *TrackRequest) Valid() error {
	return validate.Struct(x)
}

type TrackResponse struct {
	Err int `json:"error"`
}
