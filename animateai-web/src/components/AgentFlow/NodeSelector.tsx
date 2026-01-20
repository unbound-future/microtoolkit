import React from 'react';
import { Select } from '@arco-design/web-react';
import type { Node } from 'reactflow';

interface NodeSelectorProps {
  nodes: Node[];
  value?: string[];
  onChange?: (value: string[]) => void;
  placeholder?: string;
  disabled?: boolean;
  excludeNodeId?: string; // 排除的节点ID（例如当前节点本身）
}

function NodeSelector({
  nodes,
  value = [],
  onChange,
  placeholder = '选择节点',
  disabled,
  excludeNodeId,
}: NodeSelectorProps) {
  // 过滤掉排除的节点
  const availableNodes = nodes.filter(
    (node) => node.id !== excludeNodeId
  );

  const options = availableNodes.map((node) => ({
    label: node.data?.label || node.id,
    value: node.id,
  }));

  return (
    <Select
      mode="multiple"
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      disabled={disabled}
      allowClear
      style={{ width: '100%' }}
      renderFormat={(option) => {
        const node = nodes.find((n) => n.id === option.value);
        return node?.data?.label || option.value;
      }}
    >
      {options.map((option) => (
        <Select.Option key={option.value} value={option.value}>
          {option.label}
        </Select.Option>
      ))}
    </Select>
  );
}

export default NodeSelector;

