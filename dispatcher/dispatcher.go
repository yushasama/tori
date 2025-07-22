package dispatcher

import (
	"context"
	"fmt"
	"math/bits"
	"time"

	"github.com/yushasama/tori/data_structures"
	"github.com/yushasama/tori/notifier"
	"github.com/yushasama/tori/types"
)

type Dispatcher struct {
	ring      *data_structures.Ring[time.Time]  // rate limit tracker
	pending   *data_structures.Ring[*types.Job] // job queue
	notifier  notifier.Notifier
	rateLimit int
	interval  time.Duration
}

func NewDispatcher(rateLimit int, interval time.Duration, notifier notifier.Notifier) *Dispatcher {
	bitSize := bits.Len(uint(rateLimit - 1))
	size := 1 << bitSize // next power of 2

	return &Dispatcher{
		ring:      data_structures.New[time.Time](size),
		pending:   data_structures.New[*types.Job](size),
		notifier:  notifier,
		rateLimit: rateLimit,
		interval:  interval,
	}
}

func debugJob(job *types.Job) {
	fmt.Printf("job: %+v\n", job)
}

func (d *Dispatcher) Submit(job *types.Job) {
	if d.pending.IsFull() {
		d.pending.PopOldest() // drop LRU
	}

	d.pending.Push(job)
}

func (d *Dispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			d.ring.PruneBefore(now.Add(-1*d.interval), func(a, b time.Time) bool {
				return a.Before(b)
			})

			if d.pending.IsEmpty() || d.ring.Len() >= d.rateLimit {
				continue
			}

			job, ok := d.pending.PopOldest() // FIFO

			if !ok {
				continue
			}

			d.notifier.Notify(job)
			d.ring.Push(now)
		}
	}
}
