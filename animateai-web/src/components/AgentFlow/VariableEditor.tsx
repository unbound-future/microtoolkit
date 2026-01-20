import React, { useState } from 'react';
import { Input, Select, Button, Space, Popconfirm } from '@arco-design/web-react';
import { IconDelete, IconPlus } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import type { NodeVariable } from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface VariableEditorProps {
  variables?: NodeVariable[];
  onChange?: (variables: NodeVariable[]) => void;
}

function VariableEditor({ variables = [], onChange }: VariableEditorProps) {
  const t = useLocale(locale);

  const variableTypes = [
    { label: t['agentFlow.variableTypes.string'], value: 'string' },
    { label: t['agentFlow.variableTypes.number'], value: 'number' },
    { label: t['agentFlow.variableTypes.boolean'], value: 'boolean' },
    { label: t['agentFlow.variableTypes.object'], value: 'object' },
    { label: t['agentFlow.variableTypes.array'], value: 'array' },
  ];

  const handleAdd = () => {
    const newVar: NodeVariable = {
      name: '',
      type: 'string',
      value: '',
    };
    onChange?.([...variables, newVar]);
  };

  const handleUpdate = (index: number, field: keyof NodeVariable, value: any) => {
    const updated = [...variables];
    updated[index] = { ...updated[index], [field]: value };
    onChange?.(updated);
  };

  const handleDelete = (index: number) => {
    const updated = variables.filter((_, i) => i !== index);
    onChange?.(updated);
  };

  return (
    <div className={styles.variableEditor}>
      <div className={styles.variableEditorHeader}>
        <span>{t['agentFlow.variableDefinition']}</span>
        <Button
          type="text"
          size="small"
          icon={<IconPlus />}
          onClick={handleAdd}
        >
          {t['agentFlow.addVariable']}
        </Button>
      </div>
      <div className={styles.variableList}>
        {variables.map((variable, index) => (
          <div key={index} className={styles.variableItem}>
            <Space direction="vertical" style={{ width: '100%' }} size="small">
              <Space size="small" style={{ width: '100%' }}>
                <Input
                  placeholder={t['agentFlow.variableNamePlaceholder']}
                  value={variable.name}
                  onChange={(value) => handleUpdate(index, 'name', value)}
                  style={{ flex: 1 }}
                />
                <Select
                  value={variable.type}
                  onChange={(value) => handleUpdate(index, 'type', value)}
                  style={{ width: 120 }}
                >
                  {variableTypes.map((type) => (
                    <Select.Option key={type.value} value={type.value}>
                      {type.label}
                    </Select.Option>
                  ))}
                </Select>
                <Popconfirm
                  title={t['agentFlow.deleteVariableConfirm']}
                  onOk={() => handleDelete(index)}
                >
                  <Button
                    type="text"
                    status="danger"
                    icon={<IconDelete />}
                    size="small"
                  />
                </Popconfirm>
              </Space>
              <Input
                placeholder={t['agentFlow.defaultValuePlaceholder']}
                value={variable.value}
                onChange={(value) => handleUpdate(index, 'value', value)}
              />
              <Input
                placeholder={t['agentFlow.descriptionPlaceholder']}
                value={variable.description}
                onChange={(value) => handleUpdate(index, 'description', value)}
              />
            </Space>
          </div>
        ))}
        {variables.length === 0 && (
          <div className={styles.emptyHint}>{t['agentFlow.noVariables']}</div>
        )}
      </div>
    </div>
  );
}

export default VariableEditor;

