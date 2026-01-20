import React from 'react';
import { Radio, Space } from '@arco-design/web-react';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import NodeSelector from './NodeSelector';
import LogicGateTypeSelector from './LogicGateTypeSelector';
import VariableEqualitySelector from './VariableEqualitySelector';
import type { Node } from 'reactflow';
import type {
  LogicSelectorType,
  LogicSelectorConfig,
  LogicGateType,
  NodeVariable,
  VariableEqualityConfig,
} from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface LogicSelectorV2Props {
  nodes: Node[];
  config: LogicSelectorConfig;
  variables: NodeVariable[];
  onConfigChange: (config: LogicSelectorConfig) => void;
  excludeNodeId?: string;
}

function LogicSelectorV2({
  nodes,
  config,
  variables,
  onConfigChange,
  excludeNodeId,
}: LogicSelectorV2Props) {
  const t = useLocale(locale);

  const handleTypeChange = (type: LogicSelectorType) => {
    const newConfig: LogicSelectorConfig = {
      type,
      downstreamNodeIds: config.downstreamNodeIds || [],
    };

    if (type === 'gate') {
      newConfig.logicGate = config.logicGate || 'AND';
    } else if (type === 'variableEquality') {
      newConfig.variableEquality = config.variableEquality || {
        variableName: '',
        compareValue: '',
        operator: 'equals',
      };
    }

    onConfigChange(newConfig);
  };

  const handleDownstreamChange = (nodeIds: string[]) => {
    onConfigChange({
      ...config,
      downstreamNodeIds: nodeIds,
    });
  };

  const handleLogicGateChange = (gate: LogicGateType) => {
    onConfigChange({
      ...config,
      logicGate: gate,
    });
  };

  const handleVariableEqualityChange = (variableEquality: VariableEqualityConfig) => {
    onConfigChange({
      ...config,
      variableEquality,
    });
  };

  return (
    <div className={styles.logicSelectorV2}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        {/* 逻辑选择器类型选择 */}
        <div className={styles.formItem}>
          <label>{t['agentFlow.logicSelector.type'] || '逻辑选择器类型'}</label>
          <Radio.Group
            value={config.type}
            onChange={handleTypeChange}
            type="button"
            style={{ width: '100%' }}
          >
            <Radio value="gate">
              {t['agentFlow.logicSelector.type.gate'] || '与非门逻辑'}
            </Radio>
            <Radio value="variableEquality">
              {t['agentFlow.logicSelector.type.variableEquality'] || '变量相等逻辑'}
            </Radio>
          </Radio.Group>
        </div>

        {/* 根据类型显示不同的逻辑选择器 */}
        {config.type === 'gate' && (
          <div className={styles.formItem}>
            <label>{t['agentFlow.triggerLogic'] || '触发逻辑（与非门）'}</label>
            <LogicGateTypeSelector
              logicGate={config.logicGate || 'AND'}
              onChange={handleLogicGateChange}
            />
            <div className={styles.hint}>
              {t['agentFlow.logicGate.hint'] ||
                '使用逻辑与非门来控制触发条件，满足逻辑条件时将触发下游节点'}
            </div>
          </div>
        )}

        {config.type === 'variableEquality' && (
          <div className={styles.formItem}>
            <VariableEqualitySelector
              config={
                config.variableEquality || {
                  variableName: '',
                  compareValue: '',
                  operator: 'equals',
                }
              }
              variables={variables}
              onChange={handleVariableEqualityChange}
            />
          </div>
        )}

        {/* 下游节点绑定 */}
        <div className={styles.formItem}>
          <label>{t['agentFlow.downstreamNodeLabel'] || '绑定下游节点'}</label>
          <NodeSelector
            nodes={nodes}
            value={config.downstreamNodeIds || []}
            onChange={handleDownstreamChange}
            placeholder={t['agentFlow.selectDownstreamNodes'] || '选择下游节点（多选）'}
            excludeNodeId={excludeNodeId}
          />
          <div className={styles.hint}>
            {t['agentFlow.downstreamBindingHint'] ||
              '选择满足逻辑条件时要触发的下游节点'}
          </div>
        </div>
      </Space>
    </div>
  );
}

export default LogicSelectorV2;

