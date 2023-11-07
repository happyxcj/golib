package lib

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// TODO(xcj): 协程池避免无业务时协程冗余？
// 异步工作者，消费信息为方法
type FnWorker struct {
	// 消费协程个数
	consumerCount int
	fnChs          []chan func()
	// 当前使用的消费者索引
	currConsumerIndex uint64
	// 等待所有协程完成退出
	wg sync.WaitGroup
}

func NewFnWorker(consumerCount, chBufferSize int) *FnWorker {
	w := &FnWorker{
		fnChs:          make([]chan func(), consumerCount),
		consumerCount: consumerCount,
	}
	for i := 0; i < consumerCount; i++ {
		fCh := make(chan func(), chBufferSize)
		w.fnChs[i] = fCh
		go w.serveLoop(fCh)
	}
	return w
}

func (w *FnWorker) serveLoop(fnCh chan func()) {
	w.wg.Add(1)
	defer w.wg.Done()
	for {
		fn := <-fnCh
		if fn == nil {
			// 退出信号
			return
		}
		fn()
	}
}

// 获取下一个消费者，简单策略：平均分配
func (w *FnWorker) findNextConsumer() chan func() {
	next := atomic.AddUint64(&w.currConsumerIndex, 1)
	index := next % uint64(w.consumerCount)
	return w.fnChs[index]
}

// 服务指定方法fn直至成功
func (w *FnWorker) Serve(fn func()) error {
	if fn == nil {
		// 防止人为失误导致协程异常退出服务
		return errors.New("fn is nil")
	}
	fCh := w.findNextConsumer()
	fCh <- fn
	return nil
}

// 服务指定方法f，最多等待timeout时长，超过则响应失败
func (w *FnWorker) ServeWithTimeout(fn func(), timeout time.Duration) error {
	if fn == nil {
		// 防止人为失误导致协程异常退出服务
		return errors.New("fn is nil")
	}
	fCh := w.findNextConsumer()
	t := acquireTimer(timeout)
	defer releaseTimer(t)
	select {
	case fCh <- fn:
	case <-t.C:
		return errors.New("failed to serve the fn, caused by timeout")
	}
	return nil
}

// 服务指定方法,失败时(队列满了)直接返回，不等待
func (w *FnWorker) ServeRightAway(fn func()) error {
	if fn == nil {
		// 防止人为失误导致协程异常退出服务
		return errors.New("fn is nil")
	}
	fnCh := w.findNextConsumer()
	select {
	case fnCh <- fn:
		return nil
	default:
		return errors.New("slow consumer detected")
	}
}


// 优雅暂停异步工作者：有待处理消息则等处理完再退出
// 返回：暂停前待处理消息数和处理后最终丢失消息数
func (w *FnWorker) GracefulStop() (int, int) {
	// 待处理个数
	pendingCount := 0
	for _, fnCh := range w.fnChs {
		pendingCount += len(fnCh)
	}
	for _, fnCh := range w.fnChs {
		fnCh <- nil
	}
	w.wg.Wait()
	// 最终丢失个数
	lostCount := 0
	for _, fnCh := range w.fnChs {
		lostCount += len(fnCh)
	}
	return pendingCount, lostCount
}
