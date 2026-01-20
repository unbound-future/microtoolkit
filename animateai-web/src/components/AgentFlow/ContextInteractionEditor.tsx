import React from 'react';
import { Radio, Space } from '@arco-design/web-react';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import VariableEditor from './VariableEditor';
import type { ContextInteractionMode, NodeVariable } from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface ContextInteractionEditorProps {
  mode: ContextInteractionMode;
  variables: NodeVariable[];
  onModeChange: (mode: ContextInteractionMode) => void;
  onVariablesChange: (variables: NodeVariable[]) => void;
}

function ContextInteractionEditor({
  mode,
  variables,
  onModeChange,
  onVariablesChange,
}: ContextInteractionEditorProps) {
  const t = useLocale(locale);

  return (
    <div className={styles.contextInteractionEditor}>
      <Space direction="vertical" size="medium" style={{ width: '100%' }}>
        <div className={styles.formItem}>
          <label>{t['agentFlow.contextInteractionMode'] || '上下文交互模式'}：</label>
          <Radio.Group
            type="button"
            value={mode}
            onChange={(value) => onModeChange(value as ContextInteractionMode)}
          >
            <Radio value="full">
              {t['agentFlow.contextModeFull'] || '全量传递，持续上下文交互'}
            </Radio>
            <Radio value="incremental">
              {t['agentFlow.contextModeIncremental'] || '增量传递指定上下文信息'}
            </Radio>
          </Radio.Group>
          <div className={styles.hint}>
            {mode === 'full'
              ? t['agentFlow.contextModeFullHint'] || '所有上下文信息都会被传递，保持持续交互'
              : t['agentFlow.contextModeIncrementalHint'] || '仅传递指定的上下文信息，需要为每个变量添加描述'}
          </div>
        </div>

        {mode === 'incremental' && (
          <div className={styles.formItem}>
            <label>{t['agentFlow.contextVariables'] || '上下文信息变量'}：</label>
            <VariableEditor
              variables={variables}
              onChange={onVariablesChange}
            />
            <div className={styles.hint}>
              {t['agentFlow.contextVariablesHint'] || '为每个变量添加描述，说明该上下文信息的用途'}
            </div>
          </div>
        )}
      </Space>
    </div>
  );
}

export default ContextInteractionEditor;

