package task

type Status int

const (
	ToDo Status = iota
	InProgress
	Done
)

const (
	firstStatus = ToDo
	lastStatus  = Done
)

const (
	NumberOfStatus = 3
)

var StatusNames = map[Status]string{
	ToDo:       "To Do",
	InProgress: "In Progress",
	Done:       "Done",
}

func Next(status Status) Status {
	if status == lastStatus {
		return firstStatus
	}
	return status + 1
}

func Prev(status Status) Status {
	if status == firstStatus {
		return lastStatus
	}
	return status - 1
}
