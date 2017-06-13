package s_prtimer

import (
	"sync/atomic"
	"time"
)

const (
	mainChanLength = 1024
)

const (
	afterFunc = iota
	afterFuncWithAfterFuncFinishedCallback
	execute
	stop
	stopWithStopFinishedCallback
	reset
	resetWithResetFinishedCallback
	remove
	removeWithRemoveFinishedCallback
	pause
	pauseWithPauseFinishedCallback
	resume
	resumeWithResumeFinishedCallback
)

type event struct {
	_type int
	arg   interface{}
}

type afterFuncEvent struct {
	id uint64
	d  time.Duration
	f  func()
}

type afterFuncWithAfterFuncFinishedCallbackEvent struct {
	id uint64
	d  time.Duration
	f  func()
	cb func(uint64)
}

type executeEvent struct {
	id uint64
	f  func()
}

type stopEvent struct {
	id uint64
}

type stopWithStopFinishedCallbackEvent struct {
	id uint64
	f  func(bool)
}

type resetEvent struct {
	id uint64
	d  time.Duration
}

type resetWithResetFinishedCallbackEvent struct {
	id uint64
	d  time.Duration
	f  func(bool)
}

type removeEvent struct {
	id uint64
}

type removeWithRemoveFinishedCallbackEvent struct {
	id uint64
	f  func(bool)
}

type pauseEvent struct {
	id uint64
}

type pauseWithPauseFinishedCallbackEvent struct {
	id uint64
	f  func(bool)
}

type resumeEvent struct {
	id uint64
}

type resumeWithResumeFinishedCallbackEvent struct {
	id uint64
	f  func(bool)
}

type Timer struct {
	allocID        uint64
	mainChan       chan *event
	timers         map[uint64]*time.Timer
	endTimestamp   map[uint64]int64
	pauseTimestamp map[uint64]int64
}

func New() *Timer {
	t := new(Timer)
	t.allocID = 0
	t.mainChan = make(chan *event, mainChanLength)
	t.timers = make(map[uint64]*time.Timer)
	t.endTimestamp = make(map[uint64]int64)
	t.pauseTimestamp = make(map[uint64]int64)
	return t
}

func (t *Timer) GetMainChan() <-chan *event {
	return t.mainChan
}

func (t *Timer) Deal(e *event) {
	switch e._type {
	case afterFunc:
		arg := e.arg.(*afterFuncEvent)
		t.timers[arg.id] = time.AfterFunc(arg.d, func() {
			t.mainChan <- &event{
				_type: execute,
				arg: &executeEvent{
					id: arg.id,
					f:  arg.f,
				},
			}
		})
		t.endTimestamp[arg.id] = arg.d.Nanoseconds() + time.Now().UnixNano()
	case afterFuncWithAfterFuncFinishedCallback:
		arg := e.arg.(*afterFuncWithAfterFuncFinishedCallbackEvent)
		t.timers[arg.id] = time.AfterFunc(arg.d, func() {
			t.mainChan <- &event{
				_type: execute,
				arg: &executeEvent{
					id: arg.id,
					f:  arg.f,
				},
			}
		})
		t.endTimestamp[arg.id] = arg.d.Nanoseconds() + time.Now().UnixNano()
		arg.cb(arg.id)
	case execute:
		arg := e.arg.(*executeEvent)
		arg.f()
		delete(t.timers, arg.id)
		delete(t.endTimestamp, arg.id)
		delete(t.pauseTimestamp, arg.id)
	case stop:
		arg := e.arg.(*stopEvent)
		timer, exist := t.timers[arg.id]
		if exist {
			timer.Stop()
		}
	case stopWithStopFinishedCallback:
		arg := e.arg.(*stopWithStopFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			result = timer.Stop()
		}
		arg.f(result)
	case reset:
		arg := e.arg.(*resetEvent)
		timer, exist := t.timers[arg.id]
		if exist {
			timer.Reset(arg.d)
		}
	case resetWithResetFinishedCallback:
		arg := e.arg.(*resetWithResetFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			result = timer.Reset(arg.d)
		}
		arg.f(result)
	case remove:
		arg := e.arg.(*removeEvent)
		timer, exist := t.timers[arg.id]
		if exist {
			timer.Stop()
			delete(t.timers, arg.id)
			delete(t.endTimestamp, arg.id)
			delete(t.pauseTimestamp, arg.id)
		}
	case removeWithRemoveFinishedCallback:
		arg := e.arg.(*removeWithRemoveFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			timer.Stop()
			delete(t.timers, arg.id)
			delete(t.endTimestamp, arg.id)
			delete(t.pauseTimestamp, arg.id)
			result = true
		}
		arg.f(result)
	case pause:
		arg := e.arg.(*pauseEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			result = timer.Stop()
		}
		if result {
			t.pauseTimestamp[arg.id] = time.Now().UnixNano()
		}
	case pauseWithPauseFinishedCallback:
		arg := e.arg.(*pauseWithPauseFinishedCallbackEvent)
		timer, exist := t.timers[arg.id]
		result := false
		if exist {
			result = timer.Stop()
		}
		if result {
			t.pauseTimestamp[arg.id] = time.Now().UnixNano()
		}
		arg.f(result)
	case resume:
		arg := e.arg.(*resumeEvent)
		timer, timerExist := t.timers[arg.id]
		pauseTimestamp, pauseTimestampExist := t.pauseTimestamp[arg.id]
		if timerExist && pauseTimestampExist {
			d := time.Duration(t.endTimestamp[arg.id] - pauseTimestamp)
			timer.Reset(d)
			delete(t.pauseTimestamp, arg.id)
		}
	case resumeWithResumeFinishedCallback:
		arg := e.arg.(*resumeWithResumeFinishedCallbackEvent)
		timer, timerExist := t.timers[arg.id]
		pauseTimestamp, pauseTimestampExist := t.pauseTimestamp[arg.id]
		result := false
		if timerExist && pauseTimestampExist {
			d := time.Duration(t.endTimestamp[arg.id] - pauseTimestamp)
			timer.Reset(d)
			result = true
			delete(t.pauseTimestamp, arg.id)
		}
		arg.f(result)
	}
}

func (t *Timer) AfterFunc(d time.Duration, f func()) uint64 {
	id := atomic.AddUint64(&t.allocID, 1)
	go func() {
		t.mainChan <- &event{
			_type: afterFunc,
			arg: &afterFuncEvent{
				id: id,
				d:  d,
				f:  f,
			},
		}
	}()
	return id
}

func (t *Timer) AfterFuncWithAfterFuncFinishedCallback(d time.Duration, f func(), cb func(uint64)) uint64 {
	id := atomic.AddUint64(&t.allocID, 1)
	go func() {
		t.mainChan <- &event{
			_type: afterFuncWithAfterFuncFinishedCallback,
			arg: &afterFuncWithAfterFuncFinishedCallbackEvent{
				id: id,
				d:  d,
				f:  f,
				cb: cb,
			},
		}
	}()
	return id
}

func (t *Timer) Stop(id uint64) {
	go func() {
		t.mainChan <- &event{
			_type: stop,
			arg: &stopEvent{
				id: id,
			},
		}
	}()
}

func (t *Timer) StopWithStopFinishedCallback(id uint64, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: stopWithStopFinishedCallback,
			arg: &stopWithStopFinishedCallbackEvent{
				id: id,
				f:  f,
			},
		}
	}()
}

func (t *Timer) Reset(id uint64, d time.Duration) {
	go func() {
		t.mainChan <- &event{
			_type: reset,
			arg: &resetEvent{
				id: id,
				d:  d,
			},
		}
	}()
}

func (t *Timer) ResetWithResetFinishedCallback(id uint64, d time.Duration, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: resetWithResetFinishedCallback,
			arg: &resetWithResetFinishedCallbackEvent{
				id: id,
				d:  d,
				f:  f,
			},
		}
	}()
}

func (t *Timer) Remove(id uint64) {
	go func() {
		t.mainChan <- &event{
			_type: remove,
			arg: &removeEvent{
				id: id,
			},
		}
	}()
}

func (t *Timer) RemoveWithRemoveFinishedCallback(id uint64, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: removeWithRemoveFinishedCallback,
			arg: &removeWithRemoveFinishedCallbackEvent{
				id: id,
				f:  f,
			},
		}
	}()
}

func (t *Timer) Pause(id uint64) {
	go func() {
		t.mainChan <- &event{
			_type: pause,
			arg: &pauseEvent{
				id: id,
			},
		}
	}()
}

func (t *Timer) PauseWithPauseFinishedCallback(id uint64, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: pauseWithPauseFinishedCallback,
			arg: &pauseWithPauseFinishedCallbackEvent{
				id: id,
				f:  f,
			},
		}
	}()
}

func (t *Timer) Resume(id uint64) {
	go func() {
		t.mainChan <- &event{
			_type: resume,
			arg: &resumeEvent{
				id: id,
			},
		}
	}()
}

func (t *Timer) ResumeWithResumeFinishedCallback(id uint64, f func(bool)) {
	go func() {
		t.mainChan <- &event{
			_type: resumeWithResumeFinishedCallback,
			arg: &resumeWithResumeFinishedCallbackEvent{
				id: id,
				f:  f,
			},
		}
	}()
}
