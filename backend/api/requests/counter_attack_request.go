package requests

type CounterAttackRequest struct {
	TargetIP   string `json:"targetIp" binding:"required"`
	AttackType string `json:"attackType" binding:"required"`
	Intensity  int    `json:"intensity" binding:"required,min=1,max=5"`
}
