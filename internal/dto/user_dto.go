package dto

// Requests

type CheckTimezoneRequest struct {
	Timezone string `json:"timezone" binding:"required"`
}

type UpdateTimezoneRequest struct {
	Timezone string `json:"timezone" binding:"required"`
}

// Responses

type CheckTimezoneResponse struct {
	Match    bool   `json:"match"`
	Current  string `json:"current"`
	Detected string `json:"detected"`
}
