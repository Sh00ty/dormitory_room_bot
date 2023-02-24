package recaller

type Opt[t any] func(r *recaller[t])

func WithRedis[t any](addr, password, setName string) Opt[t] {
	return func(r *recaller[t]) {
		r.cache = NewRedisRepo[t](addr, password, setName)
	}
}

func WithDeadChannel[t any](buffer uint) Opt[t] {
	return func(r *recaller[t]) {
		r.deadChan = make(chan t, buffer)
		r.closeChan = make(chan struct{})
		r.withDeadChan = true
	}
}

func WithExpRegrassion[t any]() Opt[t] {
	return func(r *recaller[t]) {
		r.srtategy = exponential
	}
}

func WithLinear[t any]() Opt[t] {
	return func(r *recaller[t]) {
		r.srtategy = linear
	}
}
