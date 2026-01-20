import React from 'react';
import { Select, Input } from '@arco-design/web-react';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import type { NodeVariable, VariableEqualityConfig } from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface VariableEqualitySelectorProps {
  config: VariableEqualityConfig;
  variables: NodeVariable[];
  onChange: (config: VariableEqualityConfig) => void;
}

const operators = [
  { label: '等于 (==)', value: 'equals' },
  { label: '不等于 (!=)', value: 'notEquals' },
  { label: '大于 (>)', value: 'greaterThan' },
  { label: '小于 (<)', value: 'lessThan' },
  { label: '包含 (contains)', value: 'contains' },
];

function VariableEqualitySelector({
  config,
  variables,
  onChange,
}: VariableEqualitySelectorProps) {
  const t = useLocale(locale);

  const handleChange = (field: keyof VariableEqualityConfig, value: any) => {
    onChange({
      ...config,
      [field]: value,
    });
  };

  return (
    <div className={styles.variableEqualitySelector}>
      <div className={styles.formItem}>
        <label>{t['agentFlow.variableEquality.variableName'] || '选择变量'}</label>
        <Select
          value={config.variableName}
          onChange={(value) => handleChange('variableName', value)}
          placeholder={t['agentFlow.variableEquality.selectVariable'] || '请选择要比较的变量'}
          style={{ width: '100%' }}
        >
          {variables.map((variable) => (
            <Select.Option key={variable.name} value={variable.name}>
              {variable.name} ({variable.type})
            </Select.Option>
          ))}
        </Select>
      </div>
      <div className={styles.formItem}>
        <label>{t['agentFlow.variableEquality.operator'] || '比较操作符'}</label>
        <Select
          value={config.operator || 'equals'}
          onChange={(value) => handleChange('operator', value)}
          style={{ width: '100%' }}
        >
          {operators.map((op) => (
            <Select.Option key={op.value} value={op.value}>
              {op.label}
            </Select.Option>
          ))}
        </Select>
      </div>
      <div className={styles.formItem}>
        <label>{t['agentFlow.variableEquality.compareValue'] || '比较值'}</label>
        <Input
          value={config.compareValue}
          onChange={(value) => handleChange('compareValue', value)}
          placeholder={t['agentFlow.variableEquality.compareValuePlaceholder'] || '输入要比较的值'}
        />
      </div>
      <div className={styles.hint}>
        {t['agentFlow.variableEquality.hint'] || '当所选变量的值与比较值完全相等（或满足比较条件）时，将触发下游节点'}
      </div>
    </div>
  );
}

export default VariableEqualitySelector;

