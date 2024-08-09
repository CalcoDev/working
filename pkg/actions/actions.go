package actions

type Actionable interface {
	Subscribe()
	Invoke()
}

type Action struct {
	Callbacks [](func())
}
type Action1[TParam any] struct {
	Callbacks [](func(TParam))
}
type Action2[T1 any, T2 any] struct {
	Callbacks [](func(T1, T2))
}
type Action3[T1 any, T2 any, T3 any] struct {
	Callbacks []func(T1, T2, T3)
}
type Action4[T1 any, T2 any, T3 any, T4 any] struct {
	Callbacks []func(T1, T2, T3, T4)
}
type Action5[T1 any, T2 any, T3 any, T4 any, T5 any] struct {
	Callbacks []func(T1, T2, T3, T4, T5)
}

// FUNCTIONS 0
func NewAction(callbacks ...func()) Action {
	action := Action{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}
func (a *Action) Subscribe(callback func()) {
	a.Callbacks = append(a.Callbacks, callback)
}
func (a *Action) Invoke() {
	for _, callback := range a.Callbacks {
		callback()
	}
}

// FUNCTIONS 1
func NewAction1[TParam any](callbacks ...func(TParam)) Action1[TParam] {
	action := Action1[TParam]{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}
func (a *Action1[TParam]) Subscribe(callback func(TParam)) {
	a.Callbacks = append(a.Callbacks, callback)
}
func (a *Action1[TParam]) Invoke(param TParam) {
	for _, callback := range a.Callbacks {
		callback(param)
	}
}

// FUNCTIONS 2
func NewAction2[T1 any, T2 any](callbacks ...func(T1, T2)) Action2[T1, T2] {
	action := Action2[T1, T2]{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}
func (a *Action2[T1, T2]) Subscribe(callback func(T1, T2)) {
	a.Callbacks = append(a.Callbacks, callback)
}
func (a *Action2[T1, T2]) Invoke(param1 T1, param2 T2) {
	for _, callback := range a.Callbacks {
		callback(param1, param2)
	}
}

// FUNCTIONS 3
func NewAction3[T1 any, T2 any, T3 any](callbacks ...func(T1, T2, T3)) Action3[T1, T2, T3] {
	action := Action3[T1, T2, T3]{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}
func (a *Action3[T1, T2, T3]) Subscribe(callback func(T1, T2, T3)) {
	a.Callbacks = append(a.Callbacks, callback)
}
func (a *Action3[T1, T2, T3]) Invoke(param1 T1, param2 T2, param3 T3) {
	for _, callback := range a.Callbacks {
		callback(param1, param2, param3)
	}
}

// FUNCTIONS 4
func NewAction4[T1 any, T2 any, T3 any, T4 any](callbacks ...func(T1, T2, T3, T4)) Action4[T1, T2, T3, T4] {
	action := Action4[T1, T2, T3, T4]{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}
func (a *Action4[T1, T2, T3, T4]) Subscribe(callback func(T1, T2, T3, T4)) {
	a.Callbacks = append(a.Callbacks, callback)
}
func (a *Action4[T1, T2, T3, T4]) Invoke(param1 T1, param2 T2, param3 T3, param4 T4) {
	for _, callback := range a.Callbacks {
		callback(param1, param2, param3, param4)
	}
}

// FUNCTIONS 5
func NewAction5[T1 any, T2 any, T3 any, T4 any, T5 any](callbacks ...func(T1, T2, T3, T4, T5)) Action5[T1, T2, T3, T4, T5] {
	action := Action5[T1, T2, T3, T4, T5]{}
	for _, callback := range callbacks {
		action.Subscribe(callback)
	}
	return action
}
func (a *Action5[T1, T2, T3, T4, T5]) Subscribe(callback func(T1, T2, T3, T4, T5)) {
	a.Callbacks = append(a.Callbacks, callback)
}
func (a *Action5[T1, T2, T3, T4, T5]) Invoke(param1 T1, param2 T2, param3 T3, param4 T4, param5 T5) {
	for _, callback := range a.Callbacks {
		callback(param1, param2, param3, param4, param5)
	}
}
