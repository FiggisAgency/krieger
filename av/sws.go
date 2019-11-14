package av

import (
	"github.com/giorgisio/goav/swscale"
	"sync"
	"time"
)

type poolObject struct {
	m          sync.Mutex
	InUse      bool
	Uses       int
	InsertTime int64
	Context    *swscale.Context
}
type SWSContextPool struct {
	oldContexts map[int]*poolObject
}

func (pool *SWSContextPool) Return(w, h int) bool {
	if v, ok := pool.oldContexts[ w*h ]; ok {
		v.m.Lock()
		defer v.m.Unlock()

		v.InUse = false

		// this context probably isn't used much. Free it up.
		freed := false
		if v.Uses <= 5 && time.Since(time.Unix(0, v.InsertTime)).Nanoseconds() >= (10*time.Minute).Nanoseconds() {
			swscale.SwsFreecontext(v.Context)
			delete(pool.oldContexts, w*h)
			freed = true
		}

		return freed
	} else {
		// not found.
		return false
	}
}

func (pool *SWSContextPool) new(sf swscale.PixelFormat, sw, sh int, df swscale.PixelFormat, dw, dh int, flags int, inUse bool) *swscale.Context {
	ctx := swscale.SwsGetcontext(sw, sh, sf, dw, dh, df, flags, nil, nil, nil)
	po := pool.put(sw, sh, ctx)
	if inUse {
		po.m.Lock()
		defer po.m.Unlock()

		po.InUse = true
		po.Uses++

	}
	return ctx
}

func (pool *SWSContextPool) put(w, h int, ctx *swscale.Context) *poolObject {
	po := &poolObject{
		InsertTime: time.Now().UnixNano(),
		Context:    ctx,
	}
	pool.oldContexts[w*h] = po
	return po
}

func (pool *SWSContextPool) GetContext(sf swscale.PixelFormat, sw, sh int, df swscale.PixelFormat, dw, dh int, flags int) *swscale.Context {
	if v, ok := pool.oldContexts[ sw*sh  ]; ok {

		v.InUse = true
		v.Uses++

		return swscale.SwsGetcachedcontext(v.Context, sw, sh, sf, dw, dh, df, flags, nil, nil, nil)
	}

	return pool.new(sf, sw, sh, df, dw, dh, flags, true)
}
