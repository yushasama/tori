package notifier

// import "github.com/yushasama/tori/types"
// import "fmt"

import (
	"github.com/yushasama/tori/types"
)

// Notify sends a job as read-only payload. Do not modify.

type Notifier interface {
	Notify(job *types.Job)
}
