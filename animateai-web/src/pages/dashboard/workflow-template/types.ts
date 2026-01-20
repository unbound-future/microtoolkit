// 工作流模版类型定义
export interface WorkflowTemplate {
  id: string;
  template_id: string;
  name: string;
  description?: string;
  asset_id?: string;
  template_data?: any;
  created_at?: string;
  updated_at?: string;
}

// 后端返回的工作流模版类型
export interface BackendWorkflowTemplate {
  id: number;
  template_id: string;
  user_id: string;
  name: string;
  description?: string;
  asset_id?: string;
  template_data?: any;
  created_at: string;
  updated_at: string;
}
