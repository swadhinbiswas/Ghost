package notify
type Type string

const(
	TypeAgentFinished Type = "agent_finished"
)

type Notification struct{
	SessionID string
	SessionTitle string
	Type Type
}


