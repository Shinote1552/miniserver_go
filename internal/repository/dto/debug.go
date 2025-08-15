package dto

import "time"

/*

DEBUG

*/

type FullURLInfo struct {
	URLID         int64     `json:"url_id"`
	ShortKey      string    `json:"short_key"`
	OriginalURL   string    `json:"original_url"`
	URLCreatedAt  time.Time `json:"url_created_at"`
	UserID        int64     `json:"user_id"`
	UserCreatedAt time.Time `json:"user_created_at"`
}
