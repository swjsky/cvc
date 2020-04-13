package coupon

import (
	"time"
)

// Rule 卡券的规则
type Rule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	InternalID  string    `json:"internal_id"`
	Description string    `json:"description,omitempty"`
	RuleBody    string    `json:"rule_body"`
	Creator     string    `json:"creator"`
	CreatedTime time.Time `json:"created_time,omitempty" type:"DATETIME" default:"datetime('now','localtime')"`
	UpdatedTime time.Time `json:"updated_time,omitempty" type:"DATETIME"`
	DeletedTime time.Time `json:"deleted_time,omitempty" type:"DATETIME"`
}

type offsetSpan struct {
	OffSetFromAppliedDay 	uint   `json:"offSetFromAppliedDay"`
	TimeSpan    			uint    `json:"timeSpan"`
}

type timeSpan struct {
	StartTime 			string   `json:"startTime"`
	EndTime    			string    `json:"endTime"`
}

type natureTimeUnit struct {
	Unit 				string   `json:"unit"`
	EndInAdvance    	int    `json:"endInAdvance"`
}

type applyTimes struct {
	InDays 		uint `json:"inDays"`
	Times    	uint  `json:"times"`
}

type redeemTimes struct {
	Times    	uint    `json:"times"`
}

type sameBrand struct {
	Brand    	string    `json:"brand"`
}