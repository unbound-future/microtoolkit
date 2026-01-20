import React, { useState } from 'react';
import { Select, Input, Button, Space, Popconfirm } from '@arco-design/web-react';
import { IconDelete, IconPlus } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import type { Node } from 'reactflow';
import type { UpstreamNodeBinding } from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface UpstreamBindingEditorProps {
  nodes: Node[];
  bindings?: UpstreamNodeBinding[];
  onChange?: (bindings: UpstreamNodeBinding[]) => void;
  upstreamNodeIds?: string[];
}

function UpstreamBindingEditor({
  nodes,
  bindings = [],
  onChange,
  upstreamNodeIds = [],
}: UpstreamBindingEditorProps) {
  const t = useLocale(locale);

  const handleAdd = () => {
    const newBinding: UpstreamNodeBinding = {
      nodeId: '',
      variableName: '',
      bindingName: '',
    };
    onChange?.([...bindings, newBinding]);
  };

  const handleUpdate = (
    index: number,
    field: keyof UpstreamNodeBinding,
    value: string
  ) => {
    const updated = [...bindings];
    updated[index] = { ...updated[index], [field]: value };
    onChange?.(updated);
  };

  const handleDelete = (index: number) => {
    const updated = bindings.filter((_, i) => i !== index);
    onChange?.(updated);
  };

  // 获取上游节点的变量列表
  const getNodeVariables = (nodeId: string) => {
    const node = nodes.find((n) => n.id === nodeId);
    return node?.data?.variables || [];
  };

  // 可选的上游节点（只显示已选择的上游节点）
  const availableUpstreamNodes = nodes.filter((node) =>
    upstreamNodeIds.includes(node.id)
  );

  return (
    <div className={styles.bindingEditor}>
      <div className={styles.bindingEditorHeader}>
        <span>{t['agentFlow.upstreamBinding']}</span>
        <Button
          type="text"
          size="small"
          icon={<IconPlus />}
          onClick={handleAdd}
          disabled={upstreamNodeIds.length === 0}
        >
          {t['agentFlow.addBinding']}
        </Button>
      </div>
      {upstreamNodeIds.length === 0 && (
        <div className={styles.emptyHint}>
          {t['agentFlow.selectUpstreamNodeFirst']}
        </div>
      )}
      <div className={styles.bindingList}>
        {bindings.map((binding, index) => {
          const nodeVariables = binding.nodeId
            ? getNodeVariables(binding.nodeId)
            : [];

          return (
            <div key={index} className={styles.bindingItem}>
              <Space direction="vertical" style={{ width: '100%' }} size="small">
                <Space size="small" style={{ width: '100%' }}>
                  <Select
                    placeholder={t['agentFlow.selectUpstreamNode']}
                    value={binding.nodeId}
                    onChange={(value) => {
                      handleUpdate(index, 'nodeId', value);
                      // 清空变量名，因为节点变了
                      handleUpdate(index, 'variableName', '');
                    }}
                    style={{ flex: 1 }}
                  >
                    {availableUpstreamNodes.map((node) => (
                      <Select.Option key={node.id} value={node.id}>
                        {node.data?.label || node.id}
                      </Select.Option>
                    ))}
                  </Select>
                  <Popconfirm
                    title={t['agentFlow.deleteBindingConfirm']}
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
                {binding.nodeId && (
                  <>
                    <Select
                      placeholder={t['agentFlow.selectVariable']}
                      value={binding.variableName}
                      onChange={(value) => handleUpdate(index, 'variableName', value)}
                      style={{ width: '100%' }}
                    >
                      {nodeVariables.map((variable) => (
                        <Select.Option
                          key={variable.name}
                          value={variable.name}
                        >
                          {variable.name} ({variable.type})
                        </Select.Option>
                      ))}
                    </Select>
                    <Input
                      placeholder={t['agentFlow.bindingVariableNamePlaceholder']}
                      value={binding.bindingName}
                      onChange={(value) =>
                        handleUpdate(index, 'bindingName', value)
                      }
                      style={{ width: '100%' }}
                    />
                  </>
                )}
              </Space>
            </div>
          );
        })}
        {bindings.length === 0 && upstreamNodeIds.length > 0 && (
          <div className={styles.emptyHint}>{t['agentFlow.noBindings']}</div>
        )}
      </div>
    </div>
  );
}

export default UpstreamBindingEditor;

