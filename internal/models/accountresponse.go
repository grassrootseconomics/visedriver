package models

type AccountResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"` // Include the description field
	Result      struct {
		PublicKey  string `json:"publicKey"`
		TrackingId string `json:"trackingId"`
	} `json:"result"`
}
