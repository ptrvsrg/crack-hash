package model

import "time"

type ErrorOutput struct {
	XMLName struct{} `xml:"ErrorResponse" json:"-"`

	Timestamp time.Time `xml:"Timestamp" json:"timestamp" binding:"required"`
	Message   string    `xml:"Message" json:"message" binding:"required"`
	Status    int       `xml:"Status" json:"status" binding:"required,min=400,max=599"`
	Path      string    `xml:"Path" json:"path" format:"url_path" binding:"required" example:"/api/v0/example"`
}
