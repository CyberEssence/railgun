package requests

type CounterAttackRequest struct {
	TargetIP   string `json:"target_ip" binding:"required,ip"`
	AttackType string `json:"attack_type" binding:"required"`
	Intensity  int    `json:"intensity" binding:"required,min=1,max=5"`
}
