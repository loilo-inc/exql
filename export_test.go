package exql

var ErrMapRowSerialDestination = errMapRowSerialDestination
var ErrMapDestination = errMapDestination
var ErrMapManyDestination = errMapManyDestination

type Adb = db

func NewFinder(ex Executor) *finder {
	return newFinder(ex)
}
