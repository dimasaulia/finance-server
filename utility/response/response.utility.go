package response_utiliy

type StandarGetRequest struct {
	Page    string `json:"page"`
	Record  string `json:"record"`
	Search  string `json:"search"`
	OrderBy string `json:"order_by"`
	Sort    string `json:"sort"`
}
