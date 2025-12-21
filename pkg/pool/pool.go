package pool

type resetter interface {
	Reset()
}

// Pool - это простой пул объектов типа T, реализующих интерфейс resetter.
type Pool[T resetter] []T

// New создает новый пустой пул объектов типа T.
func New[T resetter]() *Pool[T] {
	return &Pool[T]{}
}

// Get извлекает объект из пула. Если пул пуст, возвращается нулевое значение типа T.
func (p *Pool[T]) Get() T {
	n := len(*p)
	if n == 0 {
		var zero T
		return zero
	}

	item := (*p)[n-1]
	*p = (*p)[:n-1]
	return item
}

// Put возвращает объект в пул, предварительно вызвав у него метод Reset().
func (p *Pool[T]) Put(item T) {
	item.Reset()
	*p = append(*p, item)
}
