package luaskills

// SkillPackageConfigBusinessIssue is one optional-key cross-field issue.
// SkillPackageConfigBusinessIssue 是单个可选键跨字段问题。
type SkillPackageConfigBusinessIssue struct {
	// Key is the optional associated declared key.
	// Key 是可选关联已声明键。
	Key *string `json:"key,omitempty"`
	// Code is the stable package-namespaced code.
	// Code 是稳定且带技能包命名空间的代码。
	Code string `json:"code"`
	// Message is the package-authored value-safe explanation.
	// Message 是技能包编写且不泄漏值的说明。
	Message string `json:"message"`
}

// SkillConfigWriteResult is one atomic package configuration write result.
// SkillConfigWriteResult 是单次原子技能包配置写入结果。
type SkillConfigWriteResult struct {
	// Revision is visible after the transaction.
	// Revision 是事务完成后可见的修订号。
	Revision string `json:"revision"`
	// Changed reports whether persisted data changed.
	// Changed 表示持久化数据是否发生变化。
	Changed bool `json:"changed"`
	// Values contains canonical submitted values.
	// Values 包含规范化的已提交值。
	Values map[string]string `json:"values"`
	// ChangedKeys contains stable sorted changed keys.
	// ChangedKeys 包含稳定排序的变更键。
	ChangedKeys []string `json:"changed_keys"`
}

// SkillConfigDeleteResult is one compare-and-swap deletion result.
// SkillConfigDeleteResult 是单次比较并交换删除结果。
type SkillConfigDeleteResult struct {
	// Revision is visible after deletion.
	// Revision 是删除后可见的修订号。
	Revision string `json:"revision"`
	// Deleted reports whether one value was removed.
	// Deleted 表示是否移除了一个值。
	Deleted bool `json:"deleted"`
	// Key is the exact targeted key.
	// Key 是精确目标键。
	Key string `json:"key"`
}

// SkillConfigEventError is one structured watcher or reload failure.
// SkillConfigEventError 是单个结构化监听或重载失败。
type SkillConfigEventError struct {
	// Code is the stable machine-readable error code.
	// Code 是稳定机器可读错误码。
	Code string `json:"code"`
	// Message is the value-safe human-readable message.
	// Message 是不泄漏值的人类可读消息。
	Message string `json:"message"`
}

// SkillConfigEvent is one ordered package configuration event.
// SkillConfigEvent 是单个有序技能包配置事件。
type SkillConfigEvent struct {
	// Sequence is the engine-local decimal sequence.
	// Sequence 是引擎内十进制序号。
	Sequence string `json:"sequence"`
	// Type is the stable event type.
	// Type 是稳定事件类型。
	Type string `json:"type"`
	// StoreScope is skills or system-skills.
	// StoreScope 是 skills 或 system-skills。
	StoreScope SkillConfigStoreScope `json:"store_scope"`
	// SkillID is the optional changed package.
	// SkillID 是可选变更技能包。
	SkillID *string `json:"skill_id,omitempty"`
	// Revision is the last known valid store revision.
	// Revision 是最后一个已知合法存储修订号。
	Revision string `json:"revision"`
	// ChangedKeys contains stable sorted changed keys.
	// ChangedKeys 包含稳定排序的变更键。
	ChangedKeys []string `json:"changed_keys,omitempty"`
	// Source is local_write or external_reload.
	// Source 是 local_write 或 external_reload。
	Source string `json:"source"`
	// RestartRequiredKeys contains changed keys recommending restart.
	// RestartRequiredKeys 包含建议重启的变更键。
	RestartRequiredKeys []string `json:"restart_required_keys,omitempty"`
	// Complete is optional package completeness.
	// Complete 是可选技能包完整性。
	Complete *bool `json:"complete,omitempty"`
	// Error is an optional structured failure.
	// Error 是可选结构化失败。
	Error *SkillConfigEventError `json:"error,omitempty"`
}

// SkillConfigEventBatch is one ordered event page.
// SkillConfigEventBatch 是单个有序事件页。
type SkillConfigEventBatch struct {
	// Events contains events after the requested cursor.
	// Events 包含请求游标之后的事件。
	Events []SkillConfigEvent `json:"events"`
	// NextSequence is the highest observed sequence.
	// NextSequence 是观察到的最高序号。
	NextSequence string `json:"next_sequence"`
}

// SkillConfigStoreRefresh is one explicit store refresh result.
// SkillConfigStoreRefresh 是单次显式存储刷新结果。
type SkillConfigStoreRefresh struct {
	// StoreScope identifies the refreshed store.
	// StoreScope 标识已刷新存储。
	StoreScope SkillConfigStoreScope `json:"store_scope"`
	// Revision is visible after refresh.
	// Revision 是刷新后可见的修订号。
	Revision string `json:"revision"`
	// Changed reports whether a newer snapshot was installed.
	// Changed 表示是否安装了更新快照。
	Changed bool `json:"changed"`
}

// InstalledSkillPackageConfigDescriptor is one physical package declaration.
// InstalledSkillPackageConfigDescriptor 是单个物理技能包声明。
type InstalledSkillPackageConfigDescriptor struct {
	// SkillID is the directory-derived package identifier.
	// SkillID 是目录派生的技能包标识符。
	SkillID string `json:"skill_id"`
	// RootName is the owning root name.
	// RootName 是所属根名称。
	RootName string `json:"root_name"`
	// AbsolutePath is the physical package path.
	// AbsolutePath 是物理技能包路径。
	AbsolutePath string `json:"absolute_path"`
	// Enabled reports whether the manifest enables the package.
	// Enabled 表示清单是否启用技能包。
	Enabled bool `json:"enabled"`
	// Shadowed reports whether an earlier root claims the identifier.
	// Shadowed 表示更高优先级根是否声明该标识符。
	Shadowed bool `json:"shadowed"`
	// Effective reports whether this physical instance is effective.
	// Effective 表示当前物理实例是否生效。
	Effective bool `json:"effective"`
	// ManifestValid reports whether the manifest is valid.
	// ManifestValid 表示清单是否合法。
	ManifestValid bool `json:"manifest_valid"`
	// ManifestIssue is an optional structured manifest issue.
	// ManifestIssue 是可选结构化清单问题。
	ManifestIssue *SkillConfigEventError `json:"manifest_issue,omitempty"`
	// SkillVersion is the optional semantic package version.
	// SkillVersion 是可选语义化技能包版本。
	SkillVersion *string `json:"skill_version,omitempty"`
	// Config contains valid package declarations as raw typed objects.
	// Config 包含合法技能包声明的原始类型化对象。
	Config []map[string]any `json:"config,omitempty"`
}
