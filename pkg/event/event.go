package event

type Event[TArgs any] interface {
	TryMerge(other Event[TArgs]) bool // 尝试将other任务与自身合并，如果成功返回true
	Execute(ctx ExecuteContext[TArgs])
}
