package core

// TODO Exception handling will be added

/*
	Slice based List implementation
 */
type List struct {
	container []interface{}
}

func (list *List) Init() *List {
	list.container = make([]interface{},0)
	return list
}

func (list *List) Add(element interface{}) {
	list.container = append(list.container, element)
}

func (list *List) Remove(element interface{}) {
	indexOfElem := list.IndexOf(element)

	list.container = append(list.container[:indexOfElem],list.container[indexOfElem+1:]...)
}

func (list *List) Get(index int) interface{} {
	return list.container[index]
}

func (list *List) Values() []interface{} {
	return list.container
}

func (list *List) Len() int {
	return len(list.container)
}

func (list *List) IndexOf(element interface{}) int {
	index := 0
	for _, containerElem := range list.container {
		if containerElem == element {
			return index
		}
		index++
	}

	return -1
}


