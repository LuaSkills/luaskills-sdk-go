package luaskills

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strings"
	"time"
)

// ConfigClient is the skill-config namespace backed by the unified runtime config store.
// ConfigClient 是基于统一运行时配置存储的 skill 配置命名空间。
type ConfigClient struct {
	client *Client
}

// Describe returns declared package configuration structures with package-authored descriptions and optional values.
// Describe 返回带有技能包作者所写说明和可选值的已声明技能包配置结构。
//
// options selects an optional package and whether unmasked effective values are returned.
// options 选择可选技能包，以及是否返回未脱敏有效值。
// The caller receives one descriptor per effective package or an FFI error.
// 调用方会收到每个有效技能包的描述符，或者一个 FFI 错误。
// Hosts must authorize IncludeValues before setting it because LuaSkills intentionally performs no disclosure policy.
// 宿主必须在设置 IncludeValues 前完成授权，因为 LuaSkills 有意不执行值披露策略。
func (c *ConfigClient) Describe(options SkillPackageConfigDescribeOptions) ([]SkillPackageConfigDescriptor, error) {
	if options.Mode == SkillPackageConfigDescribeModeInstalled {
		return nil, fmt.Errorf("installed mode requires DescribeInstalled")
	}
	var result []SkillPackageConfigDescriptor
	err := c.client.call(
		"luaskills_ffi_skill_config_describe_json",
		skillPackageConfigDescribePayload(c.client.engineID, options),
		&result,
	)
	return result, err
}

// DescribeInstalled returns physical package declarations without executing package Lua.
// DescribeInstalled 在不执行技能包 Lua 的情况下返回物理技能包声明。
func (c *ConfigClient) DescribeInstalled(options SkillPackageConfigDescribeOptions) ([]InstalledSkillPackageConfigDescriptor, error) {
	options.Mode = SkillPackageConfigDescribeModeInstalled
	options.IncludeValues = false
	var result []InstalledSkillPackageConfigDescriptor
	err := c.client.call(
		"luaskills_ffi_skill_config_describe_json",
		skillPackageConfigDescribePayload(c.client.engineID, options),
		&result,
	)
	return result, err
}

// Validate returns completeness and validity status for one effective skill package configuration.
// Validate 返回单个有效技能包配置的完整性与合法性状态。
//
// skillID is the exact owning package identifier.
// skillID 是精确的所属技能包标识。
// The caller receives one structured status or an FFI error.
// 调用方会收到一个结构化状态，或者一个 FFI 错误。
func (c *ConfigClient) Validate(skillID string) (SkillPackageConfigStatus, error) {
	var result SkillPackageConfigStatus
	err := c.client.call(
		"luaskills_ffi_skill_config_validate_json",
		skillPackageConfigValidatePayload(c.client.engineID, skillID),
		&result,
	)
	return result, err
}

// List lists flattened config records, optionally limited to one skill id.
// List 列出扁平化配置记录，并可选限制到单个 skill id。
func (c *ConfigClient) List(skillID string) ([]SkillConfigEntry, error) {
	var result []SkillConfigEntry
	payload := map[string]any{"engine_id": c.client.engineID}
	if skillID != "" {
		payload["skill_id"] = skillID
	}
	err := c.client.call("luaskills_ffi_skill_config_list_json", payload, &result)
	return result, err
}

// Get reads one config value by skill id and key.
// Get 按 skill id 与 key 读取单个配置值。
func (c *ConfigClient) Get(skillID string, key string) (SkillConfigGetResult, error) {
	var result SkillConfigGetResult
	err := c.client.call("luaskills_ffi_skill_config_get_json", map[string]any{
		"engine_id": c.client.engineID,
		"skill_id":  skillID,
		"key":       key,
	}, &result)
	return result, err
}

// Set atomically writes one typed configuration value through the unique batch transaction.
// Set 通过唯一批量事务原子写入单个类型化配置值。
func (c *ConfigClient) Set(skillID string, key string, value any, expectedRevision string) (SkillConfigWriteResult, error) {
	return c.SetValues(skillID, map[string]any{key: value}, expectedRevision)
}

// SetValues atomically writes one nonempty typed package configuration batch.
// SetValues 原子写入一个非空类型化技能包配置批次。
func (c *ConfigClient) SetValues(skillID string, values map[string]any, expectedRevision string) (SkillConfigWriteResult, error) {
	var result SkillConfigWriteResult
	if len(values) == 0 {
		return result, fmt.Errorf("configuration batch must not be empty")
	}
	for key, value := range values {
		if key == "" {
			return result, fmt.Errorf("configuration keys must be nonempty strings")
		}
		if err := validateSkillConfigValue(key, value); err != nil {
			return result, err
		}
	}
	err := c.client.call("luaskills_ffi_skill_config_set_json", map[string]any{
		"engine_id":         c.client.engineID,
		"skill_id":          skillID,
		"values":            values,
		"expected_revision": optionalConfigString(expectedRevision),
	}, &result)
	return result, err
}

// Delete removes one config value by skill id and key.
// Delete 按 skill id 与 key 删除单个配置值。
func (c *ConfigClient) Delete(skillID string, key string, expectedRevision string) (SkillConfigDeleteResult, error) {
	var result SkillConfigDeleteResult
	err := c.client.call("luaskills_ffi_skill_config_delete_json", map[string]any{
		"engine_id":         c.client.engineID,
		"skill_id":          skillID,
		"key":               key,
		"expected_revision": optionalConfigString(expectedRevision),
	}, &result)
	return result, err
}

// Refresh explicitly refreshes one selected store or both stores when storeScope is empty.
// Refresh 显式刷新一个选定存储，storeScope 为空时刷新两个存储。
func (c *ConfigClient) Refresh(storeScope SkillConfigStoreScope) ([]SkillConfigStoreRefresh, error) {
	var result []SkillConfigStoreRefresh
	err := c.client.call("luaskills_ffi_skill_config_refresh_json", map[string]any{
		"engine_id":   c.client.engineID,
		"store_scope": optionalConfigString(string(storeScope)),
	}, &result)
	return result, err
}

// PollEvents returns ordered configuration events after one optional cursor.
// PollEvents 返回一个可选游标之后的有序配置事件。
func (c *ConfigClient) PollEvents(afterSequence string, limit uint64) (SkillConfigEventBatch, error) {
	var result SkillConfigEventBatch
	if limit < 1 || limit > SkillConfigMaximumEventPollLimit {
		return result, fmt.Errorf(
			"event poll limit must be between 1 and %d",
			SkillConfigMaximumEventPollLimit,
		)
	}
	err := c.client.call("luaskills_ffi_skill_config_events_poll_json", map[string]any{
		"engine_id":      c.client.engineID,
		"after_sequence": optionalConfigString(afterSequence),
		"limit":          limit,
	}, &result)
	return result, err
}

// WaitEvents waits until an event is available, the timeout expires, or the context is canceled.
// WaitEvents 等待事件可用、超时到期或上下文取消。
func (c *ConfigClient) WaitEvents(
	ctx context.Context,
	afterSequence string,
	limit uint64,
	timeout time.Duration,
	pollInterval time.Duration,
) (SkillConfigEventBatch, error) {
	var empty SkillConfigEventBatch
	if timeout < 0 {
		return empty, fmt.Errorf("timeout must not be negative")
	}
	if pollInterval <= 0 || pollInterval > time.Minute {
		return empty, fmt.Errorf("poll interval must be within (0, 1m]")
	}
	deadline := time.Now().Add(timeout)
	for {
		batch, err := c.PollEvents(afterSequence, limit)
		if err != nil || len(batch.Events) > 0 || !time.Now().Before(deadline) {
			return batch, err
		}
		wait := pollInterval
		if remaining := time.Until(deadline); remaining < wait {
			wait = remaining
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return empty, ctx.Err()
		case <-timer.C:
		}
	}
}

// WatchEvents starts serial callback delivery and returns a channel containing the terminal error.
// WatchEvents 启动串行回调投递并返回包含终止错误的通道。
func (c *ConfigClient) WatchEvents(
	ctx context.Context,
	afterSequence string,
	limit uint64,
	pollInterval time.Duration,
	handler func(SkillConfigEvent),
) <-chan error {
	result := make(chan error, 1)
	go func() {
		defer close(result)
		cursor := afterSequence
		for {
			batch, err := c.WaitEvents(ctx, cursor, limit, pollInterval, pollInterval)
			if err != nil {
				result <- err
				return
			}
			for _, event := range batch.Events {
				handler(event)
			}
			cursor = batch.NextSequence
		}
	}()
	return result
}

// skillPackageConfigDescribePayload builds the exact JSON FFI request for one structure query.
// skillPackageConfigDescribePayload 为单次结构查询构造精确 JSON FFI 请求。
//
// engineID is the active native engine handle and options contains optional query controls.
// engineID 是活动原生引擎句柄，options 包含可选查询控制项。
// The returned map omits absent optional fields and always makes value disclosure explicit.
// 返回的映射省略缺失的可选字段，并始终显式指定值披露选择。
func skillPackageConfigDescribePayload(
	engineID uint64,
	options SkillPackageConfigDescribeOptions,
) map[string]any {
	payload := map[string]any{
		"engine_id":      engineID,
		"include_values": options.IncludeValues,
		"mode":           string(configDescribeMode(options.Mode)),
	}
	if options.SkillID != "" {
		payload["skill_id"] = options.SkillID
	}
	if options.RootName != "" {
		payload["root_name"] = options.RootName
	}
	return payload
}

// skillPackageConfigValidatePayload builds the exact JSON FFI request for one package validation.
// skillPackageConfigValidatePayload 为单个技能包校验构造精确 JSON FFI 请求。
//
// engineID is the active native engine handle and skillID is the exact owning package identifier.
// engineID 是活动原生引擎句柄，skillID 是精确的所属技能包标识。
// The returned map contains only fields accepted by the validation endpoint.
// 返回的映射仅包含校验端点接受的字段。
func skillPackageConfigValidatePayload(engineID uint64, skillID string) map[string]any {
	return map[string]any{
		"engine_id": engineID,
		"skill_id":  skillID,
	}
}

// configDescribeMode returns the unique default effective mode.
// configDescribeMode 返回唯一默认有效模式。
func configDescribeMode(mode SkillPackageConfigDescribeMode) SkillPackageConfigDescribeMode {
	if mode == "" {
		return SkillPackageConfigDescribeModeEffective
	}
	return mode
}

// optionalConfigString encodes one omitted protocol string as JSON null.
// optionalConfigString 把一个省略的协议字符串编码为 JSON null。
func optionalConfigString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

// validateSkillConfigValue rejects values outside the common lossless scalar contract.
// validateSkillConfigValue 拒绝超出公共无损标量契约的值。
func validateSkillConfigValue(key string, value any) error {
	if value == nil {
		return fmt.Errorf("configuration %q does not accept null", key)
	}
	if number, ok := value.(json.Number); ok {
		if integer, err := number.Int64(); err == nil {
			if integer < -SkillConfigMaximumSafeInteger || integer > SkillConfigMaximumSafeInteger {
				return fmt.Errorf("configuration %q integer exceeds the common safe range", key)
			}
			return nil
		}
		if !strings.ContainsAny(number.String(), ".eE") {
			return fmt.Errorf("configuration %q integer exceeds the common safe range", key)
		}
		float, err := number.Float64()
		if err != nil || math.IsInf(float, 0) || math.IsNaN(float) {
			return fmt.Errorf("configuration %q requires one finite number", key)
		}
		return nil
	}
	kind := reflect.TypeOf(value).Kind()
	reflected := reflect.ValueOf(value)
	switch kind {
	case reflect.String, reflect.Bool:
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		integer := reflected.Int()
		if integer < -SkillConfigMaximumSafeInteger || integer > SkillConfigMaximumSafeInteger {
			return fmt.Errorf("configuration %q integer exceeds the common safe range", key)
		}
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if reflected.Uint() > uint64(SkillConfigMaximumSafeInteger) {
			return fmt.Errorf("configuration %q integer exceeds the common safe range", key)
		}
		return nil
	case reflect.Float32, reflect.Float64:
		float := reflected.Float()
		if math.IsInf(float, 0) || math.IsNaN(float) {
			return fmt.Errorf("configuration %q requires one finite number", key)
		}
		return nil
	default:
		return fmt.Errorf("configuration %q requires a string, integer, float, or boolean", key)
	}
}
