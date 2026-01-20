import React, { useState, useEffect } from 'react';
import {
  Button,
  Input,
  Table,
  Modal,
  Message,
  Tabs,
  Upload,
  Space,
  Tag,
  Popconfirm,
  Spin,
} from '@arco-design/web-react';

const { TextArea } = Input;
import { IconPlus, IconDelete, IconEye, IconEdit } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import type { Asset, AssetSource } from './types';
import {
  isValidUrl,
  formatFileSize,
} from './utils/assetUtils';
import request from '@/utils/request';
import styles from './style/index.module.less';

const { TabPane } = Tabs;

// 后端返回的资产类型
interface BackendAsset {
  id?: number;
  asset_id: string;
  name: string;
  description?: string;
  url: string;
  source: 'url' | 'file';
  type: string;
  size?: number;
  mime_type?: string;
  created_at?: string;
  updated_at?: string;
  upload_time?: number;
}

function AssetManagement() {
  const t = useLocale(locale);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [activeTab, setActiveTab] = useState<'url' | 'file'>('url');
  
  // URL 添加表单
  const [url, setUrl] = useState('');
  const [urlAssetName, setUrlAssetName] = useState('');
  const [urlDescription, setUrlDescription] = useState('');
  
  // 文件上传 - 为每个文件存储自定义名称和描述
  const [fileList, setFileList] = useState<any[]>([]);
  const [fileMetadata, setFileMetadata] = useState<Record<string, { name: string; description: string }>>({});
  const [previewVisible, setPreviewVisible] = useState(false);
  const [previewUrl, setPreviewUrl] = useState<string>('');
  const [previewMimeType, setPreviewMimeType] = useState<string>('');
  
  // 编辑资产相关状态
  const [editVisible, setEditVisible] = useState(false);
  const [editingAsset, setEditingAsset] = useState<Asset | null>(null);
  const [editName, setEditName] = useState('');
  const [editDescription, setEditDescription] = useState('');
  const [editUrl, setEditUrl] = useState('');
  const [uploading, setUploading] = useState(false);

  // 转换后端资产数据到前端格式
  const convertBackendAssetToFrontend = (backendAsset: BackendAsset): Asset => {
    return {
      id: backendAsset.asset_id,
      name: backendAsset.name,
      description: backendAsset.description,
      url: backendAsset.url,
      source: backendAsset.source,
      size: backendAsset.size,
      mimeType: backendAsset.mime_type,
      uploadTime: backendAsset.upload_time || (backendAsset.created_at ? new Date(backendAsset.created_at).getTime() : Date.now()),
    };
  };

  // 加载资产列表
  const fetchAssets = async () => {
    setLoading(true);
    try {
      const response = await request.get<{ status: string; data: BackendAsset[] }>('/api/asset/list');
      if (response.data.status === 'ok' && response.data.data) {
        const convertedAssets = response.data.data.map(convertBackendAssetToFrontend);
        setAssets(convertedAssets);
      } else {
        Message.error(response.data.status === 'error' ? (response.data as any).msg : 'Failed to load assets');
      }
    } catch (error: any) {
      console.error('Failed to fetch assets:', error);
      Message.error(error.response?.data?.msg || 'Failed to load assets');
    } finally {
      setLoading(false);
    }
  };

  // 页面加载时获取资产列表
  useEffect(() => {
    fetchAssets();
  }, []);

  const handleAddByUrl = async () => {
    if (!url.trim()) {
      Message.warning(t['assetManagement.urlRequired']);
      return;
    }

    if (!isValidUrl(url)) {
      Message.error(t['assetManagement.invalidUrl']);
      return;
    }

    if (!urlAssetName.trim()) {
      Message.warning(t['assetManagement.nameRequired']);
      return;
    }

    try {
      const response = await request.post<{ status: string; data: BackendAsset; msg?: string }>('/api/asset/add-by-url', {
        name: urlAssetName.trim(),
        description: urlDescription.trim() || '',
        url: url.trim(),
      });

      if (response.data.status === 'ok' && response.data.data) {
        const newAsset = convertBackendAssetToFrontend(response.data.data);
        setAssets([...assets, newAsset]);
        Message.success(t['assetManagement.addSuccess']);
        setUrl('');
        setUrlAssetName('');
        setUrlDescription('');
        setVisible(false);
      } else {
        Message.error(response.data.msg || 'Failed to add asset');
      }
    } catch (error: any) {
      console.error('Failed to add asset by URL:', error);
      Message.error(error.response?.data?.msg || 'Failed to add asset');
    }
  };

  const handleFileChange = (fileList: any[]) => {
    setFileList(fileList);
    // 为新添加的文件初始化元数据，保留已有的元数据
    const newMetadata: Record<string, { name: string; description: string }> = { ...fileMetadata };
    fileList.forEach((fileItem, index) => {
      const file = fileItem.originFile || fileItem.file;
      if (file) {
        // 使用一致的 fileUid 生成方式：优先使用 fileItem.uid（Arco Upload 组件提供的唯一ID）
        // 注意：不要使用 Date.now()，因为会导致每次文件变化时生成不同的 ID
        const fileUid = fileItem.uid || file.uid || `file_${index}_${file.name}`;
        // 如果已有元数据，保留；否则初始化
        if (!newMetadata[fileUid]) {
          newMetadata[fileUid] = {
            name: file.name || '',
            description: '',
          };
        }
      }
    });
    // 清理已删除文件的元数据
    const currentFileUids = new Set(
      fileList
        .map((item) => {
          const file = item.originFile || item.file;
          return file ? (item.uid || file.uid || `${file.name}_${fileList.indexOf(item)}`) : null;
        })
        .filter((uid) => uid !== null)
    );
    Object.keys(newMetadata).forEach((uid) => {
      if (!currentFileUids.has(uid)) {
        delete newMetadata[uid];
      }
    });
    setFileMetadata(newMetadata);
  };

  const handleFileMetadataChange = (fileUid: string, field: 'name' | 'description', value: string) => {
    const currentMetadata = fileMetadata[fileUid] || { name: '', description: '' };
    setFileMetadata({
      ...fileMetadata,
      [fileUid]: {
        ...currentMetadata,
        [field]: value,
      },
    });
  };

  const handleUpload = async () => {
    if (fileList.length === 0) {
      Message.warning(t['assetManagement.fileRequired']);
      return;
    }

    setUploading(true);
    const uploadedAssets: Asset[] = [];
    let hasError = false;

    try {
      for (let index = 0; index < fileList.length; index++) {
        const fileItem = fileList[index];
        const file = fileItem.originFile || fileItem.file;
        if (!file) {
          console.warn(`File item at index ${index} has no file object`);
          continue;
        }

        // 验证文件对象是否为有效的 File 或 Blob 对象
        if (!(file instanceof File) && !(file instanceof Blob)) {
          console.error(`File at index ${index} is not a valid File or Blob object:`, file);
          hasError = true;
          Message.error(`Invalid file object: ${fileItem.name || 'Unknown file'}`);
          continue;
        }

        // 验证文件大小
        if (file.size === 0) {
          console.warn(`File ${file.name} has zero size`);
          Message.warning(`File ${file.name} is empty, skipping...`);
          continue;
        }

        // 使用一致的 fileUid 生成方式：优先使用 fileItem.uid（Arco Upload 组件提供的唯一ID）
        const fileUid = fileItem.uid || file.uid || `file_${index}_${file.name}`;
        const metadata = fileMetadata[fileUid] || { name: file.name || '', description: '' };

        // 创建 FormData 上传文件
        const formData = new FormData();
        // 确保传递的是 File 对象，而不是文件项
        formData.append('file', file, file.name);
        formData.append('name', metadata.name.trim() || file.name);
        formData.append('description', metadata.description.trim() || '');

        console.log(`Uploading file: ${file.name}, size: ${file.size}, type: ${file.type}`);

        try {
          // 对于FormData，axios会自动设置Content-Type为multipart/form-data并包含boundary
          // 不需要手动设置Content-Type
          const response = await request.post<{ status: string; data: BackendAsset; msg?: string }>(
            '/api/asset/upload',
            formData
          );

          if (response.data.status === 'ok' && response.data.data) {
            const newAsset = convertBackendAssetToFrontend(response.data.data);
            uploadedAssets.push(newAsset);
          } else {
            hasError = true;
            Message.error(response.data.msg || `Failed to upload file: ${file.name}`);
          }
        } catch (fileError: any) {
          hasError = true;
          console.error(`Failed to upload file ${file.name}:`, fileError);
          Message.error(fileError.response?.data?.msg || `Failed to upload file: ${file.name}`);
        }
      }

      if (uploadedAssets.length > 0) {
        setAssets((prevAssets) => [...prevAssets, ...uploadedAssets]);
        Message.success(t['assetManagement.addSuccess']);
        setFileList([]);
        setFileMetadata({});
        setVisible(false);
      } else if (!hasError) {
        Message.error('No files were uploaded successfully');
      }
    } catch (error: any) {
      console.error('Failed to upload files:', error);
      Message.error(error.response?.data?.msg || 'Failed to upload files');
    } finally {
      setUploading(false);
    }
  };

  const handleDelete = async (assetId: string) => {
    try {
      const response = await request.delete<{ status: string; msg?: string }>(`/api/asset/${assetId}`);
      
      if (response.data.status === 'ok') {
        setAssets(assets.filter((a) => a.id !== assetId));
        Message.success(t['assetManagement.deleteSuccess']);
      } else {
        Message.error(response.data.msg || 'Failed to delete asset');
      }
    } catch (error: any) {
      console.error('Failed to delete asset:', error);
      Message.error(error.response?.data?.msg || 'Failed to delete asset');
    }
  };

  const handlePreview = async (asset: Asset) => {
    // 如果是文件类型的资产，调用后端API生成预签名URL
    if (asset.source === 'file') {
      try {
        const response = await request.get<{ status: string; data?: { url: string }; msg?: string }>(
          `/api/asset/${asset.id}/presigned-url`
        );
        
        if (response.data.status === 'ok' && response.data.data?.url) {
          // 在新标签页打开预签名URL
          window.open(response.data.data.url, '_blank');
          return;
        } else {
          Message.error(response.data.msg || 'Failed to generate presigned URL');
          return;
        }
      } catch (error: any) {
        console.error('Failed to generate presigned URL:', error);
        Message.error(error.response?.data?.msg || 'Failed to generate presigned URL');
        return;
      }
    }
    
    // URL类型的资产直接预览
    setPreviewUrl(asset.url);
    setPreviewMimeType(asset.mimeType || '');
    setPreviewVisible(true);
  };

  const handleEdit = (asset: Asset) => {
    setEditingAsset(asset);
    setEditName(asset.name);
    setEditDescription(asset.description || '');
    setEditUrl(asset.url);
    setEditVisible(true);
  };

  const handleSaveEdit = async () => {
    if (!editingAsset) return;

    if (!editName.trim()) {
      Message.warning(t['assetManagement.nameRequired']);
      return;
    }

    if (editingAsset.source === 'url') {
      if (!editUrl.trim()) {
        Message.warning(t['assetManagement.urlRequired']);
        return;
      }

      if (!isValidUrl(editUrl)) {
        Message.error(t['assetManagement.invalidUrl']);
        return;
      }
    }

    try {
      const response = await request.put<{ status: string; data: BackendAsset; msg?: string }>(
        `/api/asset/${editingAsset.id}`,
        {
          name: editName.trim(),
          description: editDescription.trim() || '',
          url: editingAsset.source === 'url' ? editUrl.trim() : undefined,
        }
      );

      if (response.data.status === 'ok' && response.data.data) {
        const updatedAsset = convertBackendAssetToFrontend(response.data.data);
        setAssets(assets.map((asset) => (asset.id === editingAsset.id ? updatedAsset : asset)));
        Message.success(t['assetManagement.editSuccess']);
        setEditVisible(false);
        setEditingAsset(null);
        setEditName('');
        setEditDescription('');
        setEditUrl('');
      } else {
        Message.error(response.data.msg || 'Failed to update asset');
      }
    } catch (error: any) {
      console.error('Failed to update asset:', error);
      Message.error(error.response?.data?.msg || 'Failed to update asset');
    }
  };

  const handleCancelEdit = () => {
    setEditVisible(false);
    setEditingAsset(null);
    setEditName('');
    setEditDescription('');
    setEditUrl('');
  };

  const columns = [
    {
      title: t['assetManagement.assetName'],
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Asset) => (
        <div className={styles.assetNameCell}>
          {record.mimeType?.startsWith('image/') && (
            <img
              src={record.url}
              alt={text}
              className={styles.thumbnail}
              onError={(e) => {
                (e.target as HTMLImageElement).style.display = 'none';
              }}
            />
          )}
          <span>{text}</span>
        </div>
      ),
    },
    {
      title: t['assetManagement.assetId'] || '资产ID',
      dataIndex: 'id',
      key: 'assetId',
      render: (id: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: '12px' }}>{id}</span>
      ),
    },
    {
      title: t['assetManagement.assetDescription'],
      dataIndex: 'description',
      key: 'description',
      render: (description: string) => (
        <span style={{ color: description ? 'var(--color-text-1)' : 'var(--color-text-3)' }}>
          {description || '-'}
        </span>
      ),
    },
    {
      title: t['assetManagement.mimeType'] || 'MIME Type',
      dataIndex: 'mimeType',
      key: 'mimeType',
      render: (mimeType: string) => (
        <Tag>{mimeType || '-'}</Tag>
      ),
    },
    {
      title: t['assetManagement.assetSource'],
      dataIndex: 'source',
      key: 'source',
      render: (source: AssetSource) => (
        <Tag>{source === 'url' ? t['assetManagement.assetSource.url'] : t['assetManagement.assetSource.file']}</Tag>
      ),
    },
    {
      title: t['assetManagement.assetSize'],
      dataIndex: 'size',
      key: 'size',
      render: (size: number) => (size ? formatFileSize(size) : '-'),
    },
    {
      title: t['assetManagement.uploadTime'],
      dataIndex: 'uploadTime',
      key: 'uploadTime',
      render: (timestamp: number) => new Date(timestamp).toLocaleString(),
    },
    {
      title: t['assetManagement.actions'],
      key: 'actions',
      render: (_: any, record: Asset) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<IconEye />}
            onClick={() => handlePreview(record)}
          >
            {t['assetManagement.preview']}
          </Button>
          <Button
            type="text"
            size="small"
            icon={<IconEdit />}
            onClick={() => handleEdit(record)}
          >
            {t['assetManagement.edit']}
          </Button>
          <Popconfirm
            title={t['assetManagement.deleteConfirm']}
            onOk={() => handleDelete(record.id)}
          >
            <Button type="text" status="danger" size="small" icon={<IconDelete />}>
              {t['assetManagement.delete']}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>{t['assetManagement.title']}</h2>
        <Button
          type="primary"
          icon={<IconPlus />}
          onClick={() => setVisible(true)}
        >
          {t['assetManagement.addAsset']}
        </Button>
      </div>

      <div className={styles.content}>
        <Spin loading={loading} style={{ width: '100%' }}>
          <Table
            columns={columns}
            data={assets}
            pagination={{ pageSize: 10 }}
            noDataElement={<div className={styles.empty}>{t['assetManagement.noAssets']}</div>}
          />
        </Spin>
      </div>

      <Modal
        title={t['assetManagement.addAsset']}
        visible={visible}
        onCancel={() => {
          setVisible(false);
          setUrl('');
          setUrlAssetName('');
          setUrlDescription('');
          setFileList([]);
          setFileMetadata({});
          setActiveTab('url');
        }}
        footer={null}
        className={styles.addModal}
      >
        <Tabs activeTab={activeTab} onChange={(key) => setActiveTab(key as 'url' | 'file')}>
          <TabPane key="url" title={t['assetManagement.urlTab']}>
            <Space direction="vertical" size="medium" style={{ width: '100%' }}>
              <div>
                <label className={styles.label}>{t['assetManagement.assetName']}：</label>
                <Input
                  value={urlAssetName}
                  onChange={(value) => setUrlAssetName(value)}
                  placeholder={t['assetManagement.assetName']}
                  style={{ width: '100%' }}
                />
              </div>
              <div>
                <label className={styles.label}>{t['assetManagement.assetDescription']}：</label>
                <TextArea
                  value={urlDescription}
                  onChange={(value) => setUrlDescription(value)}
                  placeholder={t['assetManagement.assetDescriptionPlaceholder']}
                  style={{ width: '100%' }}
                  autoSize={{ minRows: 2, maxRows: 4 }}
                />
              </div>
              <div>
                <label className={styles.label}>{t['assetManagement.assetUrl']}：</label>
                <Input
                  value={url}
                  onChange={(value) => setUrl(value)}
                  placeholder={t['assetManagement.urlPlaceholder']}
                  style={{ width: '100%' }}
                />
                <div className={styles.hint}>{t['assetManagement.urlHint']}</div>
              </div>
              <div className={styles.supportedTypes}>
                {t['assetManagement.allTypesSupported'] || '支持所有类型的文件'}
              </div>
              <Button type="primary" onClick={handleAddByUrl} style={{ width: '100%' }}>
                {t['assetManagement.addByUrl']}
              </Button>
            </Space>
          </TabPane>
          <TabPane key="file" title={t['assetManagement.fileTab']}>
            <Space direction="vertical" size="medium" style={{ width: '100%' }}>
              <Upload
                fileList={fileList}
                onChange={handleFileChange}
                multiple
                autoUpload={false}
              >
                <Button>{t['assetManagement.selectFile']}</Button>
              </Upload>
              
              {/* 为每个文件显示文件名和描述输入框 */}
              {fileList.length > 0 && (
                <div style={{ width: '100%', marginTop: '16px' }}>
                  <Space direction="vertical" size="medium" style={{ width: '100%' }}>
                    {fileList.map((fileItem, index) => {
                      const file = fileItem.originFile || fileItem.file;
                      if (!file) return null;
                      // 使用一致的 fileUid 生成方式：优先使用 fileItem.uid（Arco Upload 组件提供的唯一ID）
                      // 注意：不要使用 Date.now()，因为会导致每次渲染时生成不同的 ID
                      const fileUid = fileItem.uid || file.uid || `file_${index}_${file.name}`;
                      const metadata = fileMetadata[fileUid] || { name: file.name || '', description: '' };
                      
                      return (
                        <div key={fileUid} style={{ border: '1px solid var(--color-border-2)', padding: '12px', borderRadius: '4px' }}>
                          <div style={{ marginBottom: '12px', fontWeight: 500, color: 'var(--color-text-2)', fontSize: '12px' }}>
                            {t['assetManagement.originalFileName'] || '原始文件名'}：{file.name}
                          </div>
                          <Space direction="vertical" size="small" style={{ width: '100%' }}>
                            <div>
                              <label className={styles.label} style={{ fontSize: '12px' }}>
                                {t['assetManagement.assetName']}：
                              </label>
                              <Input
                                value={metadata.name}
                                onChange={(value) => handleFileMetadataChange(fileUid, 'name', value)}
                                placeholder={t['assetManagement.assetName']}
                                style={{ width: '100%' }}
                              />
                            </div>
                            <div>
                              <label className={styles.label} style={{ fontSize: '12px' }}>
                                {t['assetManagement.assetDescription']}：
                              </label>
                              <TextArea
                                value={metadata.description}
                                onChange={(value) => handleFileMetadataChange(fileUid, 'description', value)}
                                placeholder={t['assetManagement.assetDescriptionPlaceholder']}
                                style={{ width: '100%' }}
                                autoSize={{ minRows: 2, maxRows: 3 }}
                              />
                            </div>
                          </Space>
                        </div>
                      );
                    })}
                  </Space>
                </div>
              )}
              
              <div className={styles.supportedTypes}>
                {t['assetManagement.allTypesSupported'] || '支持所有类型的文件'}
              </div>
              <Button 
                type="primary" 
                onClick={handleUpload} 
                style={{ width: '100%' }}
                loading={uploading}
                disabled={uploading || fileList.length === 0}
              >
                {uploading ? 'Uploading...' : t['assetManagement.addByFile']}
              </Button>
            </Space>
          </TabPane>
        </Tabs>
      </Modal>

      <Modal
        title={t['assetManagement.preview']}
        visible={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        footer={null}
        className={styles.previewModal}
        style={{ maxWidth: '90vw' }}
      >
        {previewMimeType.startsWith('image/') && (
          <img
            src={previewUrl}
            alt="Preview"
            style={{ width: '100%', maxHeight: '70vh', objectFit: 'contain' }}
          />
        )}
        {previewMimeType.startsWith('audio/') && (
          <audio
            src={previewUrl}
            controls
            style={{ width: '100%' }}
          />
        )}
        {previewMimeType.startsWith('video/') && (
          <video
            src={previewUrl}
            controls
            style={{ width: '100%', maxHeight: '70vh' }}
          />
        )}
        {!previewMimeType.startsWith('image/') && 
         !previewMimeType.startsWith('audio/') && 
         !previewMimeType.startsWith('video/') && (
          <div style={{ padding: '40px', textAlign: 'center' }}>
            <p>{t['assetManagement.previewNotSupported'] || '此文件类型不支持预览'}</p>
            <Button type="primary" onClick={() => window.open(previewUrl, '_blank')}>
              {t['assetManagement.openInNewTab'] || '在新标签页中打开'}
            </Button>
          </div>
        )}
      </Modal>

      {/* 编辑资产弹窗 */}
      <Modal
        title={t['assetManagement.editAsset']}
        visible={editVisible}
        onCancel={handleCancelEdit}
        onOk={handleSaveEdit}
        className={styles.addModal}
        okText={t['assetManagement.save']}
        cancelText={t['assetManagement.cancel']}
      >
        <Space direction="vertical" size="medium" style={{ width: '100%' }}>
          <div>
            <label className={styles.label}>{t['assetManagement.assetName']}：</label>
            <Input
              value={editName}
              onChange={(value) => setEditName(value)}
              placeholder={t['assetManagement.assetName']}
              style={{ width: '100%' }}
            />
          </div>
          <div>
            <label className={styles.label}>{t['assetManagement.assetDescription']}：</label>
            <TextArea
              value={editDescription}
              onChange={(value) => setEditDescription(value)}
              placeholder={t['assetManagement.assetDescriptionPlaceholder']}
              style={{ width: '100%' }}
              autoSize={{ minRows: 2, maxRows: 4 }}
            />
          </div>
          {editingAsset?.source === 'url' && (
            <div>
              <label className={styles.label}>{t['assetManagement.assetUrl']}：</label>
              <Input
                value={editUrl}
                onChange={(value) => setEditUrl(value)}
                placeholder={t['assetManagement.urlPlaceholder']}
                style={{ width: '100%' }}
              />
              <div className={styles.hint}>{t['assetManagement.urlHint']}</div>
            </div>
          )}
          {editingAsset?.source === 'file' && (
            <div>
              <label className={styles.label}>{t['assetManagement.assetUrl']}：</label>
              <Input
                value={editingAsset.url}
                disabled
                style={{ width: '100%' }}
              />
              <div className={styles.hint} style={{ color: 'var(--color-text-3)', fontSize: '12px' }}>
                {t['assetManagement.fileUrlReadOnly']}
              </div>
            </div>
          )}
        </Space>
      </Modal>
    </div>
  );
}

export default AssetManagement;

