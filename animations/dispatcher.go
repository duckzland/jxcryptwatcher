package animations

import (
	"time"

	JC "jxwatcher/core"
)

var animationDispatcher JC.Dispatcher = nil

func RegisterAnimationDispatcher() JC.Dispatcher {
	if animationDispatcher == nil {
		animationDispatcher = JC.NewDispatcher(200, 2, 200*time.Millisecond)
	}
	return animationDispatcher
}

func UseAnimationDispatcher() JC.Dispatcher {
	return animationDispatcher
}
