package p2pclient

import (
	"encoding/json"
	"time"
)

type PlanType string

const (
	VPSPlan       PlanType = "VPS"
	DedicatedPlan PlanType = "Dedicated Server"
	OtherPlan     PlanType = "Other"
)

type Os string

const (
	Ubuntu24_04 = "Ubuntu 24.04 LTS"
	Ubuntu22_04 = "Ubuntu 22.04 LTS"
	Debian11    = "Debian 11"
)

type Plan struct {
	ID          uint            `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        PlanType        `json:"type"`
	Data        json.RawMessage `json:"data"`
	Price       float64         `json:"price"`
}

type Service struct {
	ID              uint            `json:"id"`
	AccountID       uint            `json:"account_id"`
	UserID          uint            `json:"user_id"`
	PlanID          uint            `json:"plan_id"`
	Data            json.RawMessage `json:"data"`
	Status          string          `json:"status"`
	UserDefinedName string          `json:"user_defined_name"`
	FirewallEnabled bool            `json:"firewall_enabled"`
	UniqueID        uint            `json:"unique_id"`
	IP              string          `json:"ip"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}
