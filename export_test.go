package exql

var ErrMapRowSerialDestination = errMapRowSerialDestination
var ErrMapDestination = errMapDestination
var ErrMapManyDestination = errMapManyDestination

type Adb = db

func NewFinder(ex Executor, m Mapper) *finder {
	return newFinder(ex, m)
}
