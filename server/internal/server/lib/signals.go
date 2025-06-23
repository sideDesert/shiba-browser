package lib

const (
	AcceptCall = iota
	DeclineCall
)

type SignalRequestCall struct {
	Id   string
	From string
	To   string
}

type SignalRequestCallAnswer struct {
	UserId string
	Answer int
}

type SignalEndCall struct {
	CallId string
}

type SignalJoinCall struct {
	CallId string
}

type SignalLeaveCall struct {
	CallId string
	UserId string
}

type SignalStartStream struct {
	UserId     string
	ChatroomId string
}

type SignalJoinStream struct {
	UserId     string
	ChatroomId string
}

type SignalLeaveStream struct {
	UserId     string
	ChatroomId string
}

type SignalEndStream struct {
	UserId     string
	ChatroomId string
}
