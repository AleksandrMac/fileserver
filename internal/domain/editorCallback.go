package domain

type EditorCallback struct {
	Status  float64 `json:"status" validate:"required"`
	Url     string  `json:"url"`
	Key     string  `json:"key" validate:"required"`
	Actions []struct {
		Type   int    `json:"type"`
		UserId string `json:"userid"`
	} `json:"actions"`
	Token string `json:"token"`
}

func (x *EditorCallback) Validate() error {
	return validate.Struct(x)
}
