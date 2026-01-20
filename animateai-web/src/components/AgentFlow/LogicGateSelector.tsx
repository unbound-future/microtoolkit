import React from 'react';
import { Select } from '@arco-design/web-react';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import styles from './style/components.module.less';

interface LogicGateSelectorProps {
  value?: string;
  onChange?: (value: string) => void;
  disabled?: boolean;
}

function LogicGateSelector({ value, onChange, disabled }: LogicGateSelectorProps) {
  const t = useLocale(locale);

  const logicGateOptions = [
    { label: t['agentFlow.logicGate.AND'], value: 'AND' },
    { label: t['agentFlow.logicGate.NAND'], value: 'NAND' },
    { label: t['agentFlow.logicGate.OR'], value: 'OR' },
    { label: t['agentFlow.logicGate.NOR'], value: 'NOR' },
    { label: t['agentFlow.logicGate.XOR'], value: 'XOR' },
    { label: t['agentFlow.logicGate.XNOR'], value: 'XNOR' },
  ];

  return (
    <Select
      value={value}
      onChange={onChange}
      disabled={disabled}
      placeholder={t['agentFlow.selectLogicGate']}
      style={{ width: '100%' }}
    >
      {logicGateOptions.map((option) => (
        <Select.Option key={option.value} value={option.value}>
          {option.label}
        </Select.Option>
      ))}
    </Select>
  );
}

export default LogicGateSelector;

