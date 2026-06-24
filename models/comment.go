package models

type Comment struct {
	ID         string `json:"id"`
	NewsID     string `json:"news_id"`
	ParentID   string `json:"parent_id,omitempty"`
	Text       string `json:"text"`
	Author     string `json:"author"`
	CreatedAt  string `json:"created_at"`
	IsApproved bool   `json:"is_approved"`
}
