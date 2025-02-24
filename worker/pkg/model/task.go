package model

type HashCrackTaskInput struct {
	RequestID  string `xml:"RequestId" binding:"required"`
	PartNumber int    `xml:"PartNumber"`
	PartCount  int    `xml:"PartCount"`
	Hash       string `xml:"Hash" binding:"required"`
	MaxLength  int    `xml:"MaxLength" binding:"min=0,max=6"`
	Alphabet   struct {
		Symbols []string `xml:"Symbols"`
	} `xml:"Alphabet" binding:"required"`
}
