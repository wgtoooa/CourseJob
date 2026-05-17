package dto

type PlanUpsertRequest struct {
	Course       int    `json:"course"`
	Subject      string `json:"subject"`
	PlannedPairs int    `json:"planned_pairs"`
}

type PlanItemResponse struct {
	Course       int    `json:"course"`
	Subject      string `json:"subject"`
	PlannedPairs int    `json:"planned_pairs"`
}
