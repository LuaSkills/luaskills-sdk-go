package luaskills

import "encoding/json"

// SkillConfigEntry is one raw persisted package configuration record.
// SkillConfigEntry 是单条原始持久化技能包配置记录。
type SkillConfigEntry struct {
	// StoreScope identifies the concrete file-backed store containing this record.
	// StoreScope 标识包含当前记录的具体文件存储。
	StoreScope SkillConfigStoreScope `json:"store_scope"`
	// SkillID is the stable package identifier owning this record.
	// SkillID 是拥有当前记录的稳定技能包标识。
	SkillID string `json:"skill_id"`
	// Key is the stable package-level configuration key.
	// Key 是稳定的技能包级配置键。
	Key string `json:"key"`
	// Value is the unmasked raw persisted string.
	// Value 是未遮罩的原始持久化字符串。
	Value string `json:"value"`
}

// SkillConfigGetResult is one raw package configuration lookup result.
// SkillConfigGetResult 是单个原始技能包配置查询结果。
type SkillConfigGetResult struct {
	// Found reports whether one explicit persisted value exists.
	// Found 表示是否存在一个显式持久化值。
	Found bool `json:"found"`
	// SkillID is the stable queried package identifier.
	// SkillID 是被查询的稳定技能包标识。
	SkillID string `json:"skill_id"`
	// Key is the stable queried configuration key.
	// Key 是被查询的稳定配置键。
	Key string `json:"key"`
	// Value contains the raw persisted string only when Found is true.
	// Value 仅在 Found 为 true 时包含原始持久化字符串。
	Value *string `json:"value,omitempty"`
}

// SkillPackageConfigConstraints contains type-specific declaration constraints.
// SkillPackageConfigConstraints 包含类型专属的声明约束。
type SkillPackageConfigConstraints struct {
	// Minimum is the inclusive numeric lower bound, preserving the original JSON number.
	// Minimum 是包含式数值下界，并保留原始 JSON 数字精度。
	Minimum *json.Number `json:"minimum,omitempty"`
	// Maximum is the inclusive numeric upper bound, preserving the original JSON number.
	// Maximum 是包含式数值上界，并保留原始 JSON 数字精度。
	Maximum *json.Number `json:"maximum,omitempty"`
	// MinLength is the minimum Unicode scalar-value count for a string.
	// MinLength 是字符串的最小 Unicode 标量值数量。
	MinLength *uint64 `json:"min_length,omitempty"`
	// MaxLength is the maximum Unicode scalar-value count for a string.
	// MaxLength 是字符串的最大 Unicode 标量值数量。
	MaxLength *uint64 `json:"max_length,omitempty"`
}

// SkillPackageConfigEnumOption is one enumeration option returned by LuaSkills.
// SkillPackageConfigEnumOption 是 LuaSkills 返回的单个枚举选项。
type SkillPackageConfigEnumOption struct {
	// Value is the stable machine value persisted by this option.
	// Value 是当前选项持久化的稳定机器值。
	Value string `json:"value"`
	// Label is the human-readable display label supplied by the package author.
	// Label 是技能包作者提供的人类可读显示名称。
	Label string `json:"label"`
	// Description is the human-readable explanation supplied by the package author.
	// Description 是技能包作者提供的人类可读说明。
	Description string `json:"description"`
}

// SkillPackageConfigValidationError is one structured value-validation failure.
// SkillPackageConfigValidationError 是单个结构化值校验失败信息。
type SkillPackageConfigValidationError struct {
	// Code is the stable machine-readable validation code.
	// Code 是稳定机器可读校验代码。
	Code string `json:"code"`
	// Message is the human-readable validation explanation.
	// Message 是人类可读校验说明。
	Message string `json:"message"`
}

// SkillPackageConfigItemDescriptor describes one declared package configuration item and its effective state.
// SkillPackageConfigItemDescriptor 描述单个已声明技能包配置项及其有效状态。
type SkillPackageConfigItemDescriptor struct {
	// Key is the stable package-level configuration key.
	// Key 是稳定的包级配置键。
	Key string `json:"key"`
	// Type is the declared configuration value type.
	// Type 是声明的配置值类型。
	Type SkillPackageConfigType `json:"type"`
	// Required reports whether one effective value is required for completeness.
	// Required 表示完整性是否要求存在一个有效值。
	Required bool `json:"required"`
	// Sensitive is a hint consumed exclusively by host-side authorization policy.
	// Sensitive 是仅由宿主侧授权策略使用的提示。
	Sensitive bool `json:"sensitive"`
	// Description is the human-readable explanation supplied by the package author.
	// Description 是技能包作者提供的人类可读说明。
	Description string `json:"description"`
	// Constraints contains the type-specific declaration constraints.
	// Constraints 包含类型专属声明约束。
	Constraints SkillPackageConfigConstraints `json:"constraints"`
	// Options contains declared enumeration choices and is empty for non-enum types.
	// Options 包含已声明的枚举选项，非枚举类型时为空。
	Options []SkillPackageConfigEnumOption `json:"options"`
	// Default is the typed declaration default when one exists.
	// Default 是存在时声明提供的类型化默认值。
	Default any `json:"default,omitempty"`
	// Title is the optional short host-facing title.
	// Title 是可选的宿主短标题。
	Title *string `json:"title,omitempty"`
	// Group is the optional host grouping hint.
	// Group 是可选的宿主分组提示。
	Group *string `json:"group,omitempty"`
	// Order is the optional display order.
	// Order 是可选显示顺序。
	Order *int32 `json:"order,omitempty"`
	// Advanced marks one host-facing advanced item.
	// Advanced 标记一个宿主高级项。
	Advanced bool `json:"advanced"`
	// Placeholder is the optional host input placeholder.
	// Placeholder 是可选宿主输入占位文本。
	Placeholder *string `json:"placeholder,omitempty"`
	// Example is the optional typed example.
	// Example 是可选类型化示例。
	Example any `json:"example,omitempty"`
	// Format is the optional host rendering format.
	// Format 是可选宿主渲染格式。
	Format *string `json:"format,omitempty"`
	// RestartRequired reports whether a host-managed restart may be required.
	// RestartRequired 表示是否可能需要宿主管理的重启。
	RestartRequired bool `json:"restart_required"`
	// Deprecated reports whether the declaration is deprecated.
	// Deprecated 表示当前声明是否已弃用。
	Deprecated bool `json:"deprecated"`
	// DeprecationMessage is optional migration guidance.
	// DeprecationMessage 是可选迁移说明。
	DeprecationMessage *string `json:"deprecation_message,omitempty"`
	// State is the current unambiguous runtime state.
	// State 是当前无歧义运行时状态。
	State SkillPackageConfigItemState `json:"state"`
	// Satisfied reports whether the item satisfies package completeness.
	// Satisfied 表示当前项是否满足技能包完整性。
	Satisfied bool `json:"satisfied"`
	// ValidationError describes a configured value invalidated by the current declaration.
	// ValidationError 描述被当前声明判定为非法的已配置值。
	ValidationError *SkillPackageConfigValidationError `json:"validation_error,omitempty"`
	// Value is present only when the host explicitly requests unmasked effective values.
	// Value 仅在宿主显式请求未脱敏有效值时存在。
	Value *string `json:"value,omitempty"`
}

// SkillPackageConfigIssue is one package configuration completeness or validity issue.
// SkillPackageConfigIssue 是单个技能包配置完整性或合法性问题。
type SkillPackageConfigIssue struct {
	// Key is the stable configuration key that owns the issue.
	// Key 是拥有当前问题的稳定配置键。
	Key string `json:"key"`
	// Code is the stable machine-readable issue code.
	// Code 是稳定机器可读问题代码。
	Code string `json:"code"`
	// Message is the human-readable issue explanation.
	// Message 是人类可读问题说明。
	Message string `json:"message"`
}

// SkillPackageConfigStatus reports completeness and validity for one effective package configuration.
// SkillPackageConfigStatus 报告单个有效技能包配置的完整性与合法性。
type SkillPackageConfigStatus struct {
	// SkillID is the stable owning package identifier.
	// SkillID 是稳定的所属技能包标识。
	SkillID string `json:"skill_id"`
	// Complete reports whether every required and configured declared item is valid.
	// Complete 表示每个必填项和已配置声明项是否均合法。
	Complete bool `json:"complete"`
	// Revision is the immutable snapshot revision used by this status.
	// Revision 是当前状态使用的不可变快照修订号。
	Revision string `json:"revision"`
	// StoreScope is either skills or system-skills.
	// StoreScope 是 skills 或 system-skills。
	StoreScope SkillConfigStoreScope `json:"store_scope"`
	// Missing contains required declarations without an explicit or default value.
	// Missing 包含缺少显式值和默认值的必填声明。
	Missing []SkillPackageConfigIssue `json:"missing"`
	// Invalid contains persisted declared values that fail the current declaration.
	// Invalid 包含不满足当前声明的已持久化声明值。
	Invalid []SkillPackageConfigIssue `json:"invalid"`
	// BusinessIssues contains isolated cross-field validation issues.
	// BusinessIssues 包含隔离的跨字段校验问题。
	BusinessIssues []SkillPackageConfigBusinessIssue `json:"business_issues"`
	// Orphaned contains persisted keys no longer declared.
	// Orphaned 包含不再声明的持久化键。
	Orphaned []string `json:"orphaned"`
	// OrphanedCount is the number of persisted keys no longer declared by the effective package.
	// OrphanedCount 是当前有效技能包不再声明的持久化键数量。
	OrphanedCount uint64 `json:"orphaned_count"`
}

// SkillPackageConfigDescriptor is the full declared configuration structure of one skill package.
// SkillPackageConfigDescriptor 是单个技能包的完整已声明配置结构。
type SkillPackageConfigDescriptor struct {
	// SkillID is the stable package identifier.
	// SkillID 是稳定技能包标识。
	SkillID string `json:"skill_id"`
	// SkillVersion is the semantic package version that owns the declaration.
	// SkillVersion 是拥有当前声明的语义化技能包版本。
	SkillVersion string `json:"skill_version"`
	// Complete reports whether the effective package configuration is complete and valid.
	// Complete 表示有效技能包配置是否完整且合法。
	Complete bool `json:"complete"`
	// Revision is the immutable snapshot revision.
	// Revision 是不可变快照修订号。
	Revision string `json:"revision"`
	// StoreScope identifies the routed persisted store.
	// StoreScope 标识路由后的持久化存储。
	StoreScope SkillConfigStoreScope `json:"store_scope"`
	// MissingCount is the number of missing required declarations.
	// MissingCount 是缺失必填声明数量。
	MissingCount uint64 `json:"missing_count"`
	// InvalidCount is the number of invalid configured declarations.
	// InvalidCount 是非法已配置声明数量。
	InvalidCount uint64 `json:"invalid_count"`
	// BusinessIssueCount is the number of cross-field issues.
	// BusinessIssueCount 是跨字段问题数量。
	BusinessIssueCount uint64 `json:"business_issue_count"`
	// OrphanedCount is the number of persisted keys no longer declared by this package.
	// OrphanedCount 是当前技能包不再声明的持久化键数量。
	OrphanedCount uint64 `json:"orphaned_count"`
	// Orphaned contains persisted keys no longer declared.
	// Orphaned 包含不再声明的持久化键。
	Orphaned []string `json:"orphaned"`
	// Items contains package-level configuration item descriptors.
	// Items 包含包级配置项描述列表。
	Items []SkillPackageConfigItemDescriptor `json:"items"`
}

// SkillPackageConfigDescribeOptions controls one package configuration structure query.
// SkillPackageConfigDescribeOptions 控制单次技能包配置结构查询。
type SkillPackageConfigDescribeOptions struct {
	// SkillID optionally limits the query to one effective package.
	// SkillID 可选地将查询限制到单个有效技能包。
	SkillID string
	// IncludeValues explicitly requests unmasked effective values; the host must authorize this choice.
	// IncludeValues 显式请求未脱敏有效值；宿主必须对此选择执行授权。
	IncludeValues bool
	// Mode selects effective or installed declaration discovery.
	// Mode 选择有效或已安装声明发现。
	Mode SkillPackageConfigDescribeMode
	// RootName optionally filters installed discovery by physical root.
	// RootName 可选地按物理根过滤已安装发现。
	RootName string
}
