package logicCache

type item struct {
	value interface{}
	done  chan struct{}
}

func newItem(value interface{}) item {
	return item{
		value: value,
		done:  make(chan struct{}),
	}
}

func (i item) get() interface{} {
	return i.value
}

func (i *item) set(value interface{}) {
	i.value = value
}

func (i item) delete() {
	close(i.done)
}
