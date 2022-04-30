package models

import "time"

type Post struct {
	Id          int64     `json:"id"`
	PostContent string    `json:"post_content"`
	CreateAt    time.Time `json:"create_at"`
	UserId      int64     `json:"user_id"`
}
