package telemetry

import (
	"context"
	"maps"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/exaring/otelpgx"
)

const (
	nearCallers = 16

	farCallers = 64
)

type queryNames struct {
	module, own string

	learning sync.Mutex

	byPC atomic.Pointer[map[uintptr]string]
}

func newQueryNames() *queryNames {
	names := &queryNames{own: reflect.TypeFor[queryNames]().PkgPath() + "."}
	names.byPC.Store(&map[uintptr]string{})

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Path != "" {
		names.module = info.Main.Path + "/"
	}

	return names
}

func (q *queryNames) name(ctx context.Context, sql string) string {
	if q.module == "" {
		return otelpgx.SQLOperationName(ctx, sql)
	}

	var near [nearCallers]uintptr
	if name := q.caller(near[:runtime.Callers(2, near[:])]); name != "" {
		return name
	}

	var far [farCallers]uintptr
	if name := q.caller(far[:runtime.Callers(2, far[:])]); name != "" {
		return name
	}

	return otelpgx.SQLOperationName(ctx, sql)
}

func (q *queryNames) caller(pcs []uintptr) string {
	known := *q.byPC.Load()

	for _, pc := range pcs {
		name, found := known[pc]
		if found == false {
			name = q.learn(pc)
		}

		if name != "" {
			return name
		}
	}

	return ""
}

func (q *queryNames) learn(pc uintptr) string {
	name := q.resolve(pc)

	q.learning.Lock()
	defer q.learning.Unlock()

	learned := maps.Clone(*q.byPC.Load())
	learned[pc] = name
	q.byPC.Store(&learned)

	return name
}

func (q *queryNames) resolve(pc uintptr) string {
	frames := runtime.CallersFrames([]uintptr{pc})

	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.Function, q.own) == false {
			if function, found := strings.CutPrefix(frame.Function, q.module); found {
				return function
			}
		}

		if more == false {
			return ""
		}
	}
}
