import React from 'react';
import { Radio } from '@arco-design/web-react';
import LogicGateSelector from './LogicGateSelector';
import type { LogicGateType } from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface LogicGateTypeSelectorProps {
  logicGate: LogicGateType;
  onChange: (gate: LogicGateType) => void;
}

function LogicGateTypeSelector({
  logicGate,
  onChange,
}: LogicGateTypeSelectorProps) {
  return (
    <div className={styles.logicGateTypeSelector}>
      <LogicGateSelector
        value={logicGate}
        onChange={(value) => onChange(value as LogicGateType)}
      />
    </div>
  );
}

export default LogicGateTypeSelector;

