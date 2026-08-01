package repository

// Compile-time checks to ensure infrastructure implementations satisfy interfaces.
// These are kept commented out because the domain layer must not import
// infrastructure packages. The assertions live next to each implementation and
// are enabled in their respective epics.
//
//	var _ DB = (*database.PostgresDB)(nil)         // Epic 03
//	var _ Cache = (*cache.RedisCache)(nil)         // Epic 24
//	var _ MessageQueue = (*queue.RabbitMQQueue)(nil) // Epic 21
//	var _ MQTTBroker = (*mqtt.EMQXBroker)(nil)     // Epic 22
//	var _ Logger = (*logging.SLogger)(nil)         // Epic 18
//	var _ HTTPRouter = (*router.GinRouter)(nil)    // Epic 13
