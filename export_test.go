package exql

var ErrMapRowSerialDestination = errMapRowSerialDestination

type Adb = db

func NewFinder(ex Executor) *finder {
	return newFinder(ex)
}
