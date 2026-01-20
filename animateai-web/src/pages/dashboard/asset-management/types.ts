// 资产来源
export type AssetSource = 'url' | 'file';

// 资产数据
export interface Asset {
  id: string;
  name: string;
  description?: string; // 资产描述
  url: string;
  source: AssetSource;
  file?: File; // 本地文件对象
  size?: number; // 文件大小（字节）
  uploadTime: number; // 上传时间戳
  mimeType?: string; // MIME 类型（用于预览）
}

