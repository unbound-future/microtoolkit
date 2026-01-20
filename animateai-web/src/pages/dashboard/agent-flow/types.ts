// 上下文交互模式
export type ContextInteractionMode = 'full' | 'incremental';

// 节点数据类型定义（用于增量传递模式下的上下文信息）
export interface NodeVariable {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'object' | 'array';
  value?: any;
  description?: string; // 上下文信息描述
}

export interface UpstreamNodeBinding {
  nodeId: string;
  variableName: string;
  bindingName: string; // 绑定到当前节点的变量名
}

// 节点关联关系（当前节点到下一个节点的链接）
export interface NodeConnection {
  // 目标节点ID
  targetNodeId: string;
  // 逻辑描述词（描述为什么选择这个节点）
  logicDescription: string;
}

// 组件输入参数
export interface ComponentInputParam {
  name: string;
  value: string;
  description?: string;
}

// 节点关联的组件配置
export interface NodeComponent {
  componentId: string; // 组件ID
  description?: string; // 组件描述
  inputParams?: ComponentInputParam[]; // 组件输入参数
}

export interface NodeConfig {
  // 节点基本信息
  label: string;
  description?: string;
  
  // 关联的资产ID（可选）
  assetId?: string;
  
  // 关联的工具组件列表（可选，支持多个组件组合）
  components?: NodeComponent[];
  
  // 上游调用描述列表（表述节点可以对外提供哪些功能，可以被哪些上游节点调用）
  upstreamCallDescriptions?: string[];
  
  // 上下文交互模式：full-全量传递持续上下文交互（默认），incremental-增量传递指定上下文信息
  contextInteractionMode?: ContextInteractionMode;
  
  // 增量传递模式下的上下文信息变量列表（只有在 contextInteractionMode === 'incremental' 时使用）
  variables?: NodeVariable[];
  
  // 节点关联关系（当前节点到下一个节点的链接）
  connections?: NodeConnection[];
  
  // 上游节点变量绑定（保留用于兼容性）
  upstreamBindings?: UpstreamNodeBinding[];
}

export interface FlowNodeData extends NodeConfig {
  // 可以扩展其他数据
}

