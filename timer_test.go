package s_prtimer_test

import (
	"github.com/heransoft/s-prtimer"
	"sync/atomic"
	"testing"
	"time"
)

const precision = 100

func TestTimer_AfterFunc(t *testing.T) {
	i := int32(0)
	testCount := int32(2000)
	testTimer(testCount, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		defer func() {
			finish <- 0
		}()
		time1 := time.Now().UnixNano()
		resultChan := make(chan int64, 1)
		timer.AfterFunc(time.Second, func() {
			time2 := time.Now().UnixNano()
			i++
			resultChan <- time2 - time1
		})
		consumeTime := <-resultChan
		if consumeTime > (time.Second + time.Millisecond*precision).Nanoseconds() {
			t.Error("timeout", index, consumeTime)
		}
	})
	if i != testCount {
		t.Error("no sync!")
	}
}

func TestTimer_AfterFuncWithAfterFuncFinishedCallback(t *testing.T) {
	i := int32(0)
	testCount := int32(2000)
	testTimer(testCount, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		defer func() {
			finish <- 0
		}()
		time1 := time.Now().UnixNano()
		resultChan := make(chan int64, 1)
		timer.AfterFuncWithAfterFuncFinishedCallback(time.Second, func() {
			time2 := time.Now().UnixNano()
			i++
			resultChan <- time2 - time1
		}, func(id uint64) {
		})
		consumeTime := <-resultChan
		if consumeTime > (time.Second + time.Millisecond*precision).Nanoseconds() {
			t.Error("timeout", index, consumeTime)
		}
	})
	if i != testCount {
		t.Error("no sync!")
	}
}
func TestTimer_Stop(t *testing.T) {
	testTimer(2000, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		id := timer.AfterFunc(time.Second, func() {
			t.Error(index, "execute error")
		})
		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.Stop(id)
		})
		time.AfterFunc(time.Second*3, func() {
			finish <- 0
		})
	})
}

func TestTimer_StopWithStopFinishedCallback(t *testing.T) {
	testTimer(2000, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		id := timer.AfterFunc(time.Second, func() {
			t.Error(index, "execute error")
		})
		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.StopWithStopFinishedCallback(id, func(success bool) {
				if success == false {
					t.Error(index, "stop fail")
				}
				finish <- 0
			})
		})
	})
}

func TestTimer_Reset(t *testing.T) {
	testTimer(2000, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		i := uint64(0)
		id := timer.AfterFunc(time.Second, func() {
			if atomic.LoadUint64(&i) == 0 {
				t.Error(index, "execute error")
			}
		})
		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			atomic.AddUint64(&i, 1)
			timer.Reset(id, time.Second)
		})
		time.AfterFunc(time.Second*3, func() {
			finish <- 0
		})
	})
}

func TestTimer_ResetWithResetFinishedCallback(t *testing.T) {
	executeErrorCount := uint64(0)
	testTimer(2000, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		i := uint64(0)
		id := timer.AfterFunc(time.Second, func() {
			if atomic.LoadUint64(&i) == 0 {
				t.Error(index, "execute error")
				atomic.AddUint64(&executeErrorCount, 1)
			}
		})
		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.ResetWithResetFinishedCallback(id, time.Second, func(success bool) {
				if success == false {
					t.Error(index, "reset fail")
				}
				finish <- 0
			})
		})
	})
	if executeErrorCount != 0 {
		t.Error("execute error count:", executeErrorCount)
	}
}

func TestTimer_Pause(t *testing.T) {
	executeErrorCount := uint64(0)
	testTimer(2000, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		id := timer.AfterFunc(time.Second, func() {
			t.Error(index, "execute error")
			atomic.AddUint64(&executeErrorCount, 1)
		})

		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.Pause(id)
		})
		time.AfterFunc(time.Second*3, func() {
			finish <- 0
		})
	})
	if executeErrorCount != 0 {
		t.Error("execute error count:", executeErrorCount)
	}
}

func TestTimer_PauseWithPauseFinishedCallback(t *testing.T) {
	executeErrorCount := uint64(0)
	pauseErrorCount := uint64(0)
	testCount := int32(2000)
	testTimer(testCount, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		timer.PauseWithPauseFinishedCallback(uint64(index+testCount), func(success bool) {
			if success {
				t.Error(index, "pause error")
				atomic.AddUint64(&pauseErrorCount, 1)
			}
		})
		id := timer.AfterFunc(time.Second, func() {
			t.Error(index, "execute error")
			atomic.AddUint64(&executeErrorCount, 1)
		})

		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.PauseWithPauseFinishedCallback(id, func(success bool) {
				if !success {
					t.Error(index, "pause error")
				}
				finish <- 0
			})
		})
	})

	if pauseErrorCount != 0 {
		t.Error("execute error count:", pauseErrorCount)
	}
	if executeErrorCount != 0 {
		t.Error("execute error count:", executeErrorCount)
	}
}

func TestTimer_Resume(t *testing.T) {
	executeCount := int32(0)
	testCount := int32(2000)
	testTimer(testCount, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		id := timer.AfterFunc(time.Second, func() {
			finish <- 0
			executeCount++
		})
		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.Pause(id)
		})
		time.AfterFunc(time.Second, func() {
			timer.Resume(id)
		})
	})

	if executeCount != testCount {
		t.Error("resume error miss", testCount-executeCount)
	}
}

func TestTimer_ResumeWithResumeFinishedCallback(t *testing.T) {
	pauseErrorCount := uint64(0)
	resumeErrorCount := uint64(0)
	testCount := int32(2000)
	testTimer(testCount, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		timer.ResumeWithResumeFinishedCallback(uint64(index+testCount), func(success bool) {
			if success {
				t.Error(index, "resume1 error")
				atomic.AddUint64(&resumeErrorCount, 1)
			}
		})
		id := timer.AfterFunc(time.Second, func() {
			finish <- 0
		})

		timer.ResumeWithResumeFinishedCallback(id, func(success bool) {
			if success {
				t.Error(index, "resume2 error")
				atomic.AddUint64(&resumeErrorCount, 1)
			}
		})

		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.PauseWithPauseFinishedCallback(id, func(s bool) {
				if s {
					timer.ResumeWithResumeFinishedCallback(id, func(success bool) {
						if success == false {
							t.Error(index, "resume3 error")
							atomic.AddUint64(&resumeErrorCount, 1)
						}
					})
				} else {
					t.Error(index, "pause error")
					atomic.AddUint64(&pauseErrorCount, 1)

				}
			})
		})
	})

	if resumeErrorCount != 0 {
		t.Error("resume error count:", resumeErrorCount)
	}
	if pauseErrorCount != 0 {
		t.Error("pause error count:", pauseErrorCount)
	}
}

func TestTimer_ResumeWithResumeFinishedCallback2(t *testing.T) {
	resumeErrorCount := uint64(0)
	executeCount := int32(0)
	testCount := int32(2000)
	testTimer(testCount, func(timer *s_prtimer.Timer, index int32, finish chan int64) {
		id := timer.AfterFunc(time.Second, func() {
			finish <- 0
			executeCount++
		})
		time.AfterFunc(time.Second-time.Millisecond*precision, func() {
			timer.Pause(id)
		})
		time.AfterFunc(time.Second, func() {
			timer.ResumeWithResumeFinishedCallback(id, func(success bool) {
				if success == false {
					t.Error(index, "resume error")
					atomic.AddUint64(&resumeErrorCount, 1)
				}
			})
		})
	})

	if resumeErrorCount != 0 {
		t.Error("resume error count:", resumeErrorCount)
	}
	if executeCount != testCount {
		t.Error("resume error miss", testCount-executeCount)
	}
}

func testTimer(caseCount int32, testCase func(*s_prtimer.Timer, int32, chan int64)) {
	timer := s_prtimer.New()
	mainThreadExitChan := make(chan int64, 1)
	mainThreadExitedChan := make(chan int64, 1)
	go func() {
		r := int64(0)
		defer func() {
			mainThreadExitedChan <- r
		}()
		for {
			select {
			case result := <-mainThreadExitChan:
				r = result
				return
			case timerMainChanElement := <-timer.GetMainChan():
				timer.Deal(timerMainChanElement)
			}
		}
	}()
	caseThreadExitedChan := make(chan int64, caseCount)
	for i := int32(0); i < caseCount; i++ {
		index := i
		go func() {
			testCase(timer, index, caseThreadExitedChan)
		}()
	}
	caseThreadExitedCount := int32(0)
	for {
		<-caseThreadExitedChan
		caseThreadExitedCount++
		if caseCount == caseThreadExitedCount {
			mainThreadExitChan <- 0
			<-mainThreadExitedChan
			break
		}
	}
}
