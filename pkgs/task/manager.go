package task

import (
	"fmt"
	"sync"
	"time"

	mylo "gitlink.org.cn/cloudream/common/utils/lo"
)

type Manager[TCtx any] struct {
	taskNextID uint64
	tasks      []*Task[TCtx]
	lock       sync.Mutex
	ctx        TCtx
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
		id:   fmt.Sprintf("%d", m.taskNextID),
		body: body,
	}
	m.taskNextID++

	m.tasks = append(m.tasks, task)

	m.executeTask(task)

	return task
}

// Start 遍历正在运行的任务列表，如果存在相同的任务，则直接返回这个任务，否则创建一个新任务
func (m *Manager[TCtx]) Start(body TaskBody[TCtx], cmp func(self TaskBody[TCtx], other *Task[TCtx]) bool) *Task[TCtx] {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, t := range m.tasks {
		if cmp(body, t) {
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

func (m *Manager[TCtx]) StartComparable(body ComparableTaskBody[TCtx]) *Task[TCtx] {
	return m.Start(body, func(self TaskBody[TCtx], other *Task[TCtx]) bool {
		return body.Compare(other)
	})
}

func (m *Manager[TCtx]) Find(predicate func(task *Task[TCtx]) bool) *Task[TCtx] {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, t := range m.tasks {
		if predicate(t) {
			return t
		}
	}

	return nil
}

func (m *Manager[TCtx]) FindByID(id string) *Task[TCtx] {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, t := range m.tasks {
		if t.id == id {
			return t
		}
	}

	return nil
}

func (m *Manager[TCtx]) executeTask(task *Task[TCtx]) {
	go func() {
		task.body.Execute(task, m.ctx, func(err error, opts ...CompleteOption) {
			opt := CompleteOption{}
			if len(opts) > 0 {
				opt = opts[0]
			}

			m.lock.Lock()
			if opt.Completing != nil {
				opt.Completing()
			}

			// 立刻删除任务，或者延迟一段时间再删除
			if opt.RemovingDelay == 0 {
				m.tasks = mylo.Remove(m.tasks, task)
			} else {
				go func() {
					<-time.After(opt.RemovingDelay)
					m.lock.Lock()
					m.tasks = mylo.Remove(m.tasks, task)
					m.lock.Unlock()
				}()
			}
			m.lock.Unlock()

			task.waiterLock.Lock()
			task.err = err
			task.isCompleted.Store(true)
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
		uncompleted := false
		task.waiterLock.Lock()
		if !task.isCompleted.Load() {
			task.err = fmt.Errorf("task exit without calling complete function")
			task.isCompleted.Store(true)
			uncompleted = true
		}
		task.waiterLock.Unlock()

		if uncompleted {
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
