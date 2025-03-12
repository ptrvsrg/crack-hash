package model

type HashCrackTaskInput struct {
	RequestID  string   `xml:"RequestID" validate:"required"`
	PartNumber int      `xml:"PartNumber"`
	PartCount  int      `xml:"PartCount"`
	Hash       string   `xml:"Hash" validate:"required"`
	MaxLength  int      `xml:"MaxLength" validate:"min=0,max=6"`
	Alphabet   Alphabet `xml:"Alphabet" validate:"required"`
}

type Alphabet struct {
	Symbols []string `xml:"Symbols"`
}
