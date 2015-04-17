package counter

type val struct {
	identifier string
	recieve    chan<- int
	reset      bool
}

type Value interface {
	Get(identifier string) int
	Reset(identifier string) int
}

type valCounter chan val

func count(recieve valCounter) {
	var c = make(map[string]int)

	for {
		select {
		case r := <-recieve:
			if r.reset {
				c[r.identifier] = 0
			} else {
				c[r.identifier] += 1
			}
			r.recieve <- c[r.identifier]
		}
	}
}

func New() *valCounter {
	cc := make(valCounter)
	go count(cc)
	return &cc
}

func (cc *valCounter) Get(identifier string) int {
	recieve := make(chan int)
	*cc <- val{
		identifier: identifier,
		recieve:    recieve,
	}
	return <-recieve
}
func (cc *valCounter) Reset(identifier string) int {
	recieve := make(chan int)
	*cc <- val{
		identifier: identifier,
		reset:      true,
		recieve:    recieve,
	}
	return <-recieve
}
