// 工具组件类型
export type ToolComponentType = 'asset' | 'service' | 'trigger';

// 工具组件数据
export interface ToolComponent {
  id: string;
  name: string;
  description?: string;
  type: ToolComponentType;
  assetId?: string; // 资产组件类型时使用
  serviceUrl?: string; // 服务组件类型时使用
  paramDesc?: string; // 服务组件类型时使用，参数说明
  cronExpression?: string; // 时间触发器类型时使用，Cron表达式
  createdAt?: string;
  updatedAt?: string;
}

// 后端返回的工具组件类型
export interface BackendToolComponent {
  id?: number;
  component_id: string;
  name: string;
  description?: string;
  type: 'asset' | 'service' | 'trigger';
  asset_id?: string;
  service_url?: string;
  param_desc?: string;
  cron_expression?: string;
  created_at?: string;
  updated_at?: string;
}

