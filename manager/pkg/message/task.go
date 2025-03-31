package message

type HashCrackTaskStarted struct {
	RequestID  string   `json:"requestID" xml:"RequestId" validate:"required"`
	PartNumber int      `json:"partNumber" xml:"PartNumber"`
	PartCount  int      `json:"partCount" xml:"PartCount"`
	Hash       string   `json:"hash" xml:"Hash" validate:"required"`
	MaxLength  int      `json:"maxLength" xml:"MaxLength" validate:"min=0,max=6"`
	Alphabet   Alphabet `json:"alphabet" xml:"Alphabet" validate:"required"`
}

type Alphabet struct {
	Symbols []string `json:"symbols" xml:"Symbols" validate:"required,min=1,dive,required"`
}

type HashCrackTaskResult struct {
	RequestID  string  `json:"requestID" xml:"RequestId" validate:"required"`
	PartNumber int     `json:"partNumber" xml:"PartNumber"`
	Status     string  `json:"status" xml:"Status" validate:"required,oneof=IN_PROGRESS SUCCESS ERROR"`
	Answer     *Answer `json:"answer" xml:"Answer"`
	Error      *string `json:"error" xml:"Error"`
}

type Answer struct {
	Words   []string `json:"words" xml:"Words" validate:"required,min=1,dive,required"`
	Percent float64  `json:"percent" xml:"Percent" validate:"required,min=0,max=100"`
}
