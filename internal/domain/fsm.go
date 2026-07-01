package domain

// ValidTransitions maps from-state to allowed to-states.
var ValidTransitions = map[SubscriptionState][]SubscriptionState{
	SubscriptionStateTrialing: {SubscriptionStateActive, SubscriptionStateCanceled},
	SubscriptionStateActive:   {SubscriptionStatePastDue, SubscriptionStateCanceled, SubscriptionStatePaused},
	SubscriptionStatePastDue:  {SubscriptionStateActive, SubscriptionStateCanceled},
	SubscriptionStatePaused:   {SubscriptionStateActive, SubscriptionStateCanceled},
	SubscriptionStateCanceled: {SubscriptionStateExpired},
}

func CanTransition(from, to SubscriptionState) bool {
	allowed, ok := ValidTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func ValidateTransition(from, to SubscriptionState) error {
	if !CanTransition(from, to) {
		return ErrInvalidTransition
	}
	return nil
}
