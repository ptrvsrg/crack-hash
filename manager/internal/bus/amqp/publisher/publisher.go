package publisher

import (
	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/publisher"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
)

type Publishers struct {
	TaskStarted publisher.Publisher[message.HashCrackTaskStarted]
}
