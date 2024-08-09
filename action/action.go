package action

type Actionable interface {
	Subscribe()
	Invoke()
}

type Event struct {
	Callbacks [](func())
}

func NewEvent(callbacks ...func()) Event {
	event := Event{}
	for _, callback := range callbacks {
		event.Subscribe(callback)
	}
	return event
}

func (a *Event) Subscribe(callback func()) {
	a.Callbacks = append(a.Callbacks, callback)
}

func (a *Event) Invoke() {
	for _, callback := range a.Callbacks {
		callback()
	}
}

type Action[TParam any] struct {
	Callbacks [](func(TParam))
}

func NewAction[TParam any](callbacks ...func(TParam)) Action[TParam] {
	action := Action[TParam]{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}

func (a *Action[TParam]) Subscribe(callback func(TParam)) {
	a.Callbacks = append(a.Callbacks, callback)
}

func (a *Action[TParam]) Invoke(param TParam) {
	for _, callback := range a.Callbacks {
		callback(param)
	}
}
