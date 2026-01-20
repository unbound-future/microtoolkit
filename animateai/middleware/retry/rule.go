package retry

import (
	"fmt"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/common/ruleengine"
)

type Rule struct {
	Describe                string           `json:"describe"`
	MaxRetry                int              `json:"max-retry"`
	RetryIntervalTimeSecond int              `json:"retry-interval-time-seconds"`
	Regs                    []ruleengine.Reg `json:"reg"`
}

type RuleConfig struct {
	Rules []Rule `json:"rules"`
}

var holder = ruleengine.NewConfigHolder[RuleConfig](consts.RetryRuleConfigKey)

func InitRuleConfig() error {
	return holder.Init()
}

func GetActiveRuleConfig() RuleConfig {
	return holder.Get()
}

func ParseRetryInfo(content map[string]string) (map[string]string, error) {
	cfg := GetActiveRuleConfig()
	for _, rule := range cfg.Rules {
		if ruleengine.MatchRegs(rule.Regs, content) {
			return map[string]string{
				consts.MaxRetry:      fmt.Sprintf("%d", rule.MaxRetry),
				consts.RetryInterval: fmt.Sprintf("%d", rule.RetryIntervalTimeSecond),
			}, nil
		}
	}
	return map[string]string{}, fmt.Errorf("no matching rule found for content: %+v", content)
}
