package tickevent

type TickEvent[TArgs any] interface {
	Execute(ctx ExecuteContext[TArgs])
}
