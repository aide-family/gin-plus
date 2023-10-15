package ginplus

import (
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"
)

func Test_getTag(t *testing.T) {
	type My struct {
		Name    string `json:"name"`
		Id      uint   `uri:"id"`
		Keyword string `form:"keyword"`
	}

	fieldList := getTag(reflect.TypeOf(&My{}))

	t.Log(fieldList)
}

func Test_getTag2(t *testing.T) {
	var b bool
	t.Logf("%T", b)
}

type PromStrategy struct {
	ID           int32            `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	GroupID      int32            `gorm:"column:group_id;type:int unsigned;not null;comment:所属规则组ID" json:"group_id"`                                             // 所属规则组ID
	Alert        string           `gorm:"column:alert;type:varchar(64);not null;comment:规则名称" json:"alert"`                                                       // 规则名称
	Expr         string           `gorm:"column:expr;type:text;not null;comment:prom ql" json:"expr"`                                                             // prom ql
	For          string           `gorm:"column:for;type:varchar(64);not null;default:10s;comment:持续时间" json:"for"`                                               // 持续时间
	Labels       string           `gorm:"column:labels;type:json;not null;comment:标签" json:"labels"`                                                              // 标签
	Annotations  string           `gorm:"column:annotations;type:json;not null;comment:告警文案" json:"annotations"`                                                  // 告警文案
	AlertLevelID int32            `gorm:"column:alert_level_id;type:int;not null;index:idx__alart_level_id,priority:1;comment:告警等级dict ID" json:"alert_level_id"` // 告警等级dict ID
	Status       int32            `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态: 1启用;2禁用" json:"status"`                                      // 启用状态: 1启用;2禁用
	CreatedAt    time.Time        `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`                     // 创建时间
	UpdatedAt    time.Time        `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`                     // 更新时间
	DeletedAt    gorm.DeletedAt   `gorm:"column:deleted_at;type:timestamp;comment:删除时间" json:"deleted_at"`                                                        // 删除时间
	AlarmPages   []*PromAlarmPage `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:AlarmPageID;many2many:prom_strategy_alarm_pages" json:"alarm_pages"`
	Categories   []*PromDict      `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromStrategyID;joinReferences:DictID;many2many:prom_strategy_categories" json:"categories"`
	AlertLevel   *PromDict        `gorm:"foreignKey:AlertLevelID" json:"alert_level"`
	GroupInfo    *PromGroup       `gorm:"foreignKey:GroupID" json:"group_info"`
}

type PromGroup struct {
	ID             int32           `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name           string          `gorm:"column:name;type:varchar(64);not null;comment:规则组名称" json:"name"`                                    // 规则组名称
	StrategyCount  int64           `gorm:"column:strategy_count;type:bigint;not null;comment:规则数量" json:"strategy_count"`                      // 规则数量
	Status         int32           `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用" json:"status"`                   // 启用状态1:启用;2禁用
	Remark         string          `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`                                // 描述信息
	CreatedAt      time.Time       `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt      time.Time       `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt      gorm.DeletedAt  `gorm:"column:deleted_at;type:timestamp;comment:删除时间" json:"deleted_at"`                                    // 删除时间
	PromStrategies []*PromStrategy `gorm:"foreignKey:GroupID" json:"prom_strategies"`
	Categories     []*PromDict     `gorm:"References:ID;foreignKey:ID;joinForeignKey:PromGroupID;joinReferences:DictID;many2many:prom_group_categories" json:"categories"`
}

type PromDict struct {
	ID        int32          `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name      string         `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name__category,priority:1;comment:字典名称" json:"name"`                                    // 字典名称
	Category  int32          `gorm:"column:category;type:tinyint;not null;uniqueIndex:idx__name__category,priority:2;index:idx__category,priority:1;comment:字典类型" json:"category"` // 字典类型
	Color     string         `gorm:"column:color;type:varchar(32);not null;default:#165DFF;comment:字典tag颜色" json:"color"`                                                          // 字典tag颜色
	Status    int32          `gorm:"column:status;type:tinyint;not null;default:1;comment:状态" json:"status"`                                                                       // 状态
	Remark    string         `gorm:"column:remark;type:varchar(255);not null;comment:字典备注" json:"remark"`                                                                          // 字典备注
	CreatedAt time.Time      `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`                                           // 创建时间
	UpdatedAt time.Time      `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"`                                           // 更新时间
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;type:timestamp;comment:删除时间" json:"deleted_at"`                                                                              // 删除时间
}

type PromAlarmPage struct {
	ID             int32               `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Name           string              `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:报警页面名称" json:"name"`  // 报警页面名称
	Remark         string              `gorm:"column:remark;type:varchar(255);not null;comment:描述信息" json:"remark"`                                // 描述信息
	Icon           string              `gorm:"column:icon;type:varchar(1024);not null;comment:图表" json:"icon"`                                     // 图表
	Color          string              `gorm:"column:color;type:varchar(64);not null;comment:tab颜色" json:"color"`                                  // tab颜色
	Status         int32               `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态,1启用;2禁用" json:"status"`                   // 启用状态,1启用;2禁用
	CreatedAt      time.Time           `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt      time.Time           `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt      gorm.DeletedAt      `gorm:"column:deleted_at;type:timestamp;comment:删除时间" json:"deleted_at"`                                    // 删除时间
	PromStrategies []*PromStrategy     `gorm:"References:ID;foreignKey:ID;joinForeignKey:AlarmPageID;joinReferences:PromStrategyID;many2many:prom_strategy_alarm_pages" json:"prom_strategies"`
	Histories      []*PromAlarmHistory `gorm:"References:ID;foreignKey:ID;joinForeignKey:AlarmPageID;joinReferences:HistoryID;many2many:prom_alarm_page_histories" json:"histories"`
}

type PromAlarmHistory struct {
	ID         int32            `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	Node       string           `gorm:"column:node;type:varchar(64);not null;comment:node名称" json:"node"`                                                        // node名称
	Status     string           `gorm:"column:status;type:varchar(16);not null;comment:告警消息状态, 报警和恢复" json:"status"`                                             // 告警消息状态, 报警和恢复
	Info       string           `gorm:"column:info;type:json;not null;comment:原始告警消息" json:"info"`                                                               // 原始告警消息
	CreatedAt  time.Time        `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"`                      // 创建时间
	StartAt    int64            `gorm:"column:start_at;type:bigint unsigned;not null;comment:报警开始时间" json:"start_at"`                                            // 报警开始时间
	EndAt      int64            `gorm:"column:end_at;type:bigint unsigned;not null;comment:报警恢复时间" json:"end_at"`                                                // 报警恢复时间
	Duration   int64            `gorm:"column:duration;type:bigint unsigned;not null;comment:持续时间时间戳, 没有恢复, 时间戳是0" json:"duration"`                              // 持续时间时间戳, 没有恢复, 时间戳是0
	StrategyID int32            `gorm:"column:strategy_id;type:int unsigned;not null;index:idx__strategy_id,priority:1;comment:规则ID, 用于查询时候" json:"strategy_id"` // 规则ID, 用于查询时候
	LevelID    int32            `gorm:"column:level_id;type:int unsigned;not null;index:idx__level_id,priority:1;comment:报警等级ID" json:"level_id"`                // 报警等级ID
	Md5        string           `gorm:"column:md5;type:char(32);not null;comment:md5" json:"md5"`                                                                // md5
	Pages      []*PromAlarmPage `gorm:"References:ID;foreignKey:ID;joinForeignKey:AlarmPageID;joinReferences:PageID;many2many:prom_prom_alarm_page_histories" json:"pages"`
}

type Page struct {
	Curr  int   `json:"curr"`
	Size  int   `json:"size"`
	Total int64 `json:"total"`
}

func Test_getTag1(t *testing.T) {
	// ListStrategyResp ...
	type ListStrategyResp struct {
		List []*PromStrategy `json:"list"`
		Page Page            `json:"page"`
	}

	fieldList := getTag(reflect.TypeOf(&ListStrategyResp{}))
	t.Log(fieldList)
}
