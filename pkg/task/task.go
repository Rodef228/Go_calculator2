package task

type Task struct {
	ID            string  `json:"id"`
	Arg1          float64 `json:"arg1"`
	Arg2          float64 `json:"arg2"`
	Operation     string  `json:"operation"`
	OperationTime int     `json:"operation_time"`
	Result        float64 `json:"result,omitempty"`
}

type Future struct {
	ch chan float64
}

func NewFuture() *Future {
	return &Future{ch: make(chan float64, 1)}
}

func (f *Future) SetResult(val float64) {
	f.ch <- val
}

func (f *Future) Get() float64 {
	return <-f.ch
}
