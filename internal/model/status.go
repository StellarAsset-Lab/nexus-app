package model

// OrderStatus mirrors the Order contract's on-chain status exactly. Do not
// invent an additional state such as "Completed" — the contract's terminal
// state is "Settled".
type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "Created"
	OrderStatusSettled   OrderStatus = "Settled"
	OrderStatusCancelled OrderStatus = "Cancelled"
	OrderStatusExpired   OrderStatus = "Expired"
)

func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusCreated, OrderStatusSettled, OrderStatusCancelled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// ComponentHealth is the status reported for a service/dependency on the
// status page. "Unknown" is used when a check has not actually been
// performed — never defaulted to "Operational".
type ComponentHealth string

const (
	ComponentOperational ComponentHealth = "Operational"
	ComponentDegraded    ComponentHealth = "Degraded"
	ComponentUnavailable ComponentHealth = "Unavailable"
	ComponentUnknown     ComponentHealth = "Unknown"
)
