package pool

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/panjf2000/ants/v2"
	"github.com/segmentio/ksuid"
)

// Pool 协程池工具类
type Pool struct {
	pool *ants.Pool
	once sync.Once
}

// 全局协程池
var (
	pool            *Pool
	once            sync.Once
	defaultPoolSize = 10
)

// GetPool 获取协程池实例
func GetPool() *Pool {
	once.Do(func() {
		p, err := ants.NewPool(defaultPoolSize)
		if err != nil {
			panic(err)
		}
		pool = &Pool{pool: p}

		fmt.Println("--------------POOL_OPEN---------------")
	})

	return pool
}

// Add 增加任务task并等待完成
func (t *Pool) Add(ctx context.Context, task func(c context.Context)) *Pool {
	err := t.pool.Submit(func() {
		// traceId生成并添加至context
		c := context.WithValue(ctx, consts.ServerTraceIDKey, ksuid.New().String())

		// panic恢复
		defer func() {
			if r := recover(); r != nil {
				hlog.CtxErrorf(c, "Pool task PANIC recovered: panic=%v, stack=%s", r, string(debug.Stack()))
			}
		}()

		task(c)
	})
	if err != nil {
		hlog.CtxErrorf(ctx, "Pool Add error: %v", err)
	}

	return pool
}

// Release 释放协程池资源
func (t *Pool) Release() {
	t.once.Do(func() {
		if t.pool != nil {
			// 无运行中协程 再释放
			for {
				if t.Running() == 0 {
					break
				}
				time.Sleep(time.Second)
			}
			err := t.pool.ReleaseTimeout(1 * time.Second)
			if err != nil {
				panic(err)
				return
			}
			fmt.Println("--------------POOL_CLOSE---------------")
		}
	})
}

// Running 获取当前运行的 goroutine 数量
func (t *Pool) Running() int {
	return t.pool.Running()
}
