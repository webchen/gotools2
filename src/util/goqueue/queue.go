package goqueue

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
)

// Queue 队列
type Queue struct {
	list  *list.List
	mutex sync.Mutex
}

// NewQueue 新建队列对象
func NewQueue() *Queue {
	return &Queue{
		list: list.New(),
	}
}

// Push 入队列
func (queue *Queue) Push(data interface{}) {
	if data == nil {
		return
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	queue.list.PushBack(data)
}

// Pop 出队列
func (queue *Queue) Pop() (interface{}, error) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if element := queue.list.Front(); element != nil {
		queue.list.Remove(element)
		return element.Value, nil
	}
	if queue.list.Len() == 0 {
		return nil, errors.New("队列为空，没有相关对象或数据")
	}
	return nil, errors.New("pop failed")
}

// Clear 清空队列
func (queue *Queue) Clear() {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	for element := queue.list.Front(); element != nil; {
		elementNext := element.Next()
		queue.list.Remove(element)
		element = elementNext
	}
}

// Clear2List 一次性返回所有的对象到列表
func (queue *Queue) Clear2List() (list []interface{}) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	for element := queue.list.Front(); element != nil; {
		list = append(list, element)
		elementNext := element.Next()
		queue.list.Remove(element)
		element = elementNext
	}
	return list
}

// Len 队列长度
func (queue *Queue) Len() int {
	return queue.list.Len()
}

// Show 遍历打印数据
func (queue *Queue) Show() {
	for item := queue.list.Front(); item != nil; item = item.Next() {
		fmt.Printf("%+v", item.Value)
	}
}
