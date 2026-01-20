import React from 'react';
import { Select, Input, Button, Space, Popconfirm } from '@arco-design/web-react';
import { IconDelete, IconPlus } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from '@/pages/dashboard/agent-flow/locale';
import type { Node } from 'reactflow';
import type { NodeConnection } from '@/pages/dashboard/agent-flow/types';
import styles from './style/components.module.less';

interface NodeConnectionEditorProps {
  nodes: Node[];
  connections: NodeConnection[];
  onChange: (connections: NodeConnection[]) => void;
  excludeNodeId?: string;
}

function NodeConnectionEditor({
  nodes,
  connections = [],
  onChange,
  excludeNodeId,
}: NodeConnectionEditorProps) {
  const t = useLocale(locale);

  const handleAdd = () => {
    const newConnection: NodeConnection = {
      targetNodeId: '',
      logicDescription: '',
    };
    onChange([...connections, newConnection]);
  };

  const handleUpdate = (
    index: number,
    field: keyof NodeConnection,
    value: string
  ) => {
    const updated = [...connections];
    updated[index] = { ...updated[index], [field]: value };
    onChange(updated);
  };

  const handleDelete = (index: number) => {
    const updated = connections.filter((_, i) => i !== index);
    onChange(updated);
  };

  // 可选的节点（排除当前节点）
  // 使用 useMemo 确保在 nodes 更新时重新计算
  const availableNodes = React.useMemo(
    () => nodes.filter((node) => node.id !== excludeNodeId),
    [nodes, excludeNodeId]
  );

  return (
    <div className={styles.connectionEditor}>
      <div className={styles.connectionEditorHeader}>
        <span>{t['agentFlow.nodeConnections'] || '节点关联'}</span>
        <Button
          type="text"
          size="small"
          icon={<IconPlus />}
          onClick={handleAdd}
        >
          {t['agentFlow.addConnection'] || '添加关联'}
        </Button>
      </div>
      <div className={styles.connectionList}>
        {connections.map((connection, index) => (
          <div key={index} className={styles.connectionItem}>
            <Space direction="vertical" style={{ width: '100%' }} size="small">
              <Space size="small" style={{ width: '100%', display: 'flex', alignItems: 'center' }} className={styles.selectSpace}>
                <div style={{ flex: 1, minWidth: 0, width: '100%' }}>
                  <Select
                    placeholder={t['agentFlow.selectTargetNode'] || '选择目标节点'}
                    value={connection.targetNodeId}
                    onChange={(value) => handleUpdate(index, 'targetNodeId', value)}
                    className={styles.nodeSelect}
                    style={{ width: '100%' }}
                    triggerProps={{
                      autoAlignPopupWidth: true,
                      autoAlignPopupMinWidth: true,
                    }}
                    getPopupContainer={(triggerNode) => {
                      const drawer = triggerNode.closest('.arco-drawer-body');
                      return drawer || document.body;
                    }}
                  >
                    {availableNodes.map((node) => (
                      <Select.Option
                        key={node.id}
                        value={node.id}
                      >
                        {node.data?.label || node.id}
                      </Select.Option>
                    ))}
                  </Select>
                </div>
                <Popconfirm
                  title={t['agentFlow.deleteConnectionConfirm'] || '确定删除此关联吗？'}
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
                placeholder={t['agentFlow.logicDescriptionPlaceholder'] || '输入逻辑描述词（描述为什么会链接到这个节点）'}
                value={connection.logicDescription}
                onChange={(value) => handleUpdate(index, 'logicDescription', value)}
                style={{ width: '100%' }}
              />
            </Space>
          </div>
        ))}
        {connections.length === 0 && (
          <div className={styles.emptyHint}>
            {t['agentFlow.noConnections'] || '暂无关联，点击上方按钮添加'}
          </div>
        )}
      </div>
    </div>
  );
}

export default NodeConnectionEditor;

