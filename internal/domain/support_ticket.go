package domain

import "time"

type SupportTicket struct {
	ID        string
	UserUUID  string
	Email     string
	Header    string
	Text      string
	ImagePath string
	CreatedAt time.Time
}
