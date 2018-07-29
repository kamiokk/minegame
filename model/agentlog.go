package model

import (
    "time"
    "github.com/kamiokk/minegame/helper/mysql"
)

// AgentLog model of agent percentage log
type AgentLog struct {
	ID uint
    Value float64
    CreatedAt *time.Time
    UpdatedAt *time.Time
    IsDeleted uint
}

// TableName return user table name
func (AgentLog) TableName() string {
    return "mine_agent_log"
}

// Create add log
func (r *AgentLog) Create() {
    if mysql.DBInstance().NewRecord(r) {
        mysql.DBInstance().Create(r)
    }
}