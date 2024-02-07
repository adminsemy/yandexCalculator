package queue

import (
	"sync/atomic"
	"unsafe"
)

type Expression interface {
	GetExpression() string
	SetID(uint64)
	Result() []string
}
type Node struct {
	value Expression
	next  *Node
}

type LockFreeQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

func NewLockFreeQueue() *LockFreeQueue {
	return &LockFreeQueue{}
}

func (l *LockFreeQueue) Enqueue(value Expression) {
	node := &Node{value: value}
	for {
		if atomic.LoadPointer(&l.tail) == nil {
			if atomic.CompareAndSwapPointer(&l.tail, l.tail, unsafe.Pointer(node)) {
				break
			}
		}
		if atomic.LoadPointer(&l.head) == nil {
			first := (*Node)(atomic.LoadPointer(&l.tail))
			first.next = node
			if atomic.CompareAndSwapPointer(&l.head, l.head, unsafe.Pointer(node)) {
				break
			}
		}
		oldNode := (*Node)(atomic.LoadPointer(&l.head))
		oldNode.next = node
		ptr := unsafe.Pointer(node)
		ptrOldNode := unsafe.Pointer(oldNode)
		if atomic.CompareAndSwapPointer(&l.head, ptrOldNode, ptr) {
			break
		}
	}
}

func (l *LockFreeQueue) Dequeue() (Expression, bool) {
	var result *Node
	for {
		result = (*Node)(atomic.LoadPointer(&l.tail))
		if result == nil {
			return nil, false
		}
		newTail := result.next
		newTailUnsafe := unsafe.Pointer(newTail)
		if atomic.CompareAndSwapPointer(&l.tail, l.tail, newTailUnsafe) {
			break
		}
	}
	return result.value, true
}
