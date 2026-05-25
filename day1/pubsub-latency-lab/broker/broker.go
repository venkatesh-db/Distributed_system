package broker

import "pubsublatencylab/shared"

var EmailTopic = make(chan shared.OrderEvent, 100)

var InventoryTopic = make(chan shared.OrderEvent, 100)

var AnalyticsTopic = make(chan shared.OrderEvent, 100)
