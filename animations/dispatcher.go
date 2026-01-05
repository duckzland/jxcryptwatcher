package animations

import (
	"time"

	JC "jxwatcher/core"
)

var animationDispatcher JC.Dispatcher = nil

func RegisterAnimationDispatcher() JC.Dispatcher {
	if animationDispatcher == nil {
		animationDispatcher = JC.NewDispatcher(8, 2, 200*time.Millisecond)
		animationDispatcher.SetKey("Animations")
	}
	return animationDispatcher
}

func UseAnimationDispatcher() JC.Dispatcher {
	return animationDispatcher
}
