package userinput

import "sync"


type MessageQueue struct {
	cond *sync.Cond
	data []*UserInput
}

func NewMessageQueue() *MessageQueue {
	return &MessageQueue{
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

func (q *MessageQueue) Size() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.data)
}

// Push 放入消息进入队尾
func (q *MessageQueue) Push(userInput *UserInput) {
	q.cond.L.Lock()
	q.data = append(q.data, userInput)
	q.cond.L.Unlock()
	q.cond.Signal()
}


// GetInput 获取队首数据
func (q *MessageQueue) GetInput() *UserInput {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	// 阻塞直到 被唤醒 且 队列不为空
	for len(q.data) == 0 {
		q.cond.Wait()
	}

	v := q.data[0]
	q.data = q.data[1:]
	return v
}

// GetInput 获取队首数据并检查类型, 如果 类型不匹配 则什么都不做
func (q *MessageQueue) GetTypedInput(t InputTypeEnum) *UserInput {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if len(q.data) == 0 {
		return nil
	}

	v := q.data[0]
	if v.Type() == t {
		q.data = q.data[1:]
		return v
	}
	return nil
}
