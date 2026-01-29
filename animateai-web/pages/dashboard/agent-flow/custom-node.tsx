import React from 'react';
import { Handle, Position } from 'reactflow';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import type { NodeConfig } from './types';
import styles from './style/custom-node.module.less';

interface CustomNodeProps {
  data: NodeConfig;
  selected?: boolean;
}

function CustomNode({ data, selected }: CustomNodeProps) {
  const t = useLocale(locale);

  return (
    <div className={`${styles.node} ${selected ? styles.selected : ''}`}>
      {/* 四个方向的连接点，每个方向都可以作为 source 和 target */}
      <Handle type="target" position={Position.Top} id="top" />
      <Handle type="source" position={Position.Top} id="top" />
      
      <Handle type="target" position={Position.Right} id="right" />
      <Handle type="source" position={Position.Right} id="right" />
      
      <Handle type="target" position={Position.Bottom} id="bottom" />
      <Handle type="source" position={Position.Bottom} id="bottom" />
      
      <Handle type="target" position={Position.Left} id="left" />
      <Handle type="source" position={Position.Left} id="left" />
      
      <div className={styles.nodeContent}>
        <div className={styles.nodeTitle}>{data.label}</div>
        {data.description && (
          <div className={styles.nodeDescription}>{data.description}</div>
        )}
        {data.upstreamCallDescriptions && data.upstreamCallDescriptions.length > 0 && (
          <div className={styles.nodeUpstreamCall}>
            <div className={styles.nodeUpstreamCallTitle}>
              {t['agentFlow.upstreamCallDescriptions'] || '上游调用'}:
            </div>
            {data.upstreamCallDescriptions.map((desc, index) => (
              <div key={index} className={styles.nodeUpstreamCallItem}>
                {desc}
              </div>
            ))}
          </div>
        )}
        {data.variables && data.variables.length > 0 && (
          <div className={styles.nodeVariables}>
            {t['agentFlow.variables']}: {data.variables.length}
          </div>
        )}
        {data.connections && data.connections.length > 0 && (
          <div className={styles.nodeConnections}>
            {t['agentFlow.connections'] || '关联'}: {data.connections.length}
          </div>
        )}
      </div>
    </div>
  );
}

export default CustomNode;

