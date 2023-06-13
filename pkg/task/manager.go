package task

import (
	"fmt"
	"sync"
)

type Manager[TCtx any] struct {
	tasks []*Task[TCtx]
	lock  sync.Mutex
	ctx   TCtx
}

func NewManager[TCtx any](ctx TCtx) Manager[TCtx] {
	return Manager[TCtx]{
		ctx: ctx,
	}
}

// StartNew 启动一个新任务
func (m *Manager[TCtx]) StartNew(body TaskBody[TCtx]) *Task[TCtx] {
	m.lock.Lock()
	defer m.lock.Unlock()

	task := &Task[TCtx]{
		body: body,
	}

	m.tasks = append(m.tasks, task)

	m.executeTask(task)

	return task
}

// Start 遍历正在运行的任务列表，如果存在相同的任务，则直接返回这个任务，否则创建一个新任务
func (m *Manager[TCtx]) Start(body TaskBody[TCtx], cmp func(self, other TaskBody[TCtx]) bool) *Task[TCtx] {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, t := range m.tasks {
		if cmp(body, t.body) {
			return t
		}
	}

	task := &Task[TCtx]{
		body: body,
	}

	m.tasks = append(m.tasks, task)

	m.executeTask(task)

	return task
}

func (m *Manager[TCtx]) StartCmp(body ComparableTaskBody[TCtx]) *Task[TCtx] {
	return m.Start(body, func(self, other TaskBody[TCtx]) bool {
		return body.Compare(other)
	})
}

func (m *Manager[TCtx]) Find(predicate func(body TaskBody[TCtx]) bool) *Task[TCtx] {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, t := range m.tasks {
		if predicate(t.body) {
			return t
		}
	}

	return nil
}

func (m *Manager[TCtx]) executeTask(task *Task[TCtx]) {
	go func() {
		task.body.Execute(m.ctx, func(err error, completing func()) {
			// 删除任务
			m.lock.Lock()
			for i, t := range m.tasks {
				if t == task {
					m.tasks[i] = m.tasks[len(m.tasks)-1]
					m.tasks = m.tasks[:len(m.tasks)-1]
					break
				}
			}
			completing()
			m.lock.Unlock()

			task.waiterLock.Lock()
			task.isCompleted = true
			task.err = err
			task.waiterLock.Unlock()

			// 触发回调
			for _, w := range task.waiters {
				close(w)
			}

			for _, c := range task.onCompleted {
				c(task)
			}
		})

		// 如果Task没有调用complete函数就退出了，那么就认为是出错结束
		notCompletedYet := false
		task.waiterLock.Lock()
		if !task.isCompleted {
			task.isCompleted = true
			task.err = fmt.Errorf("task exit without calling complete function")
			notCompletedYet = true
		}
		task.waiterLock.Unlock()

		if notCompletedYet {
			// 触发回调
			for _, w := range task.waiters {
				close(w)
			}

			for _, c := range task.onCompleted {
				c(task)
			}
		}
	}()
}