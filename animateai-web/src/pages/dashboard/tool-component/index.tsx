import React, { useState, useEffect } from 'react';
import {
  Button,
  Input,
  Table,
  Modal,
  Message,
  Space,
  Tag,
  Popconfirm,
  Spin,
  Select,
  Radio,
  Form,
} from '@arco-design/web-react';

const { TextArea } = Input;
const { Option } = Select;
const RadioGroup = Radio.Group;
import { IconPlus, IconDelete, IconEdit } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import type { ToolComponent, ToolComponentType, BackendToolComponent } from './types';
import request from '@/utils/request';
import styles from './style/index.module.less';

// 后端返回的资产类型（用于资产选择下拉框）
interface BackendAsset {
  asset_id: string;
  name: string;
}

function ToolComponentManagement() {
  const t = useLocale(locale);
  const [components, setComponents] = useState<ToolComponent[]>([]);
  const [assets, setAssets] = useState<BackendAsset[]>([]);
  const [loading, setLoading] = useState(false);
  const [visible, setVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [editingComponent, setEditingComponent] = useState<ToolComponent | null>(null);
  
  // 表单状态
  const [form] = Form.useForm();
  const [editForm] = Form.useForm();
  
  // 转换后端组件数据到前端格式
  const convertBackendComponentToFrontend = (backendComponent: BackendToolComponent): ToolComponent => {
    return {
      id: backendComponent.component_id,
      name: backendComponent.name,
      description: backendComponent.description,
      type: backendComponent.type,
      assetId: backendComponent.asset_id,
      serviceUrl: backendComponent.service_url,
      paramDesc: backendComponent.param_desc,
      cronExpression: backendComponent.cron_expression,
      createdAt: backendComponent.created_at,
      updatedAt: backendComponent.updated_at,
    };
  };

  // 加载组件列表
  const fetchComponents = async () => {
    setLoading(true);
    try {
      const response = await request.get<{ status: string; data: BackendToolComponent[] }>('/api/tool-component/list');
      if (response.data.status === 'ok' && response.data.data) {
        const convertedComponents = response.data.data.map(convertBackendComponentToFrontend);
        setComponents(convertedComponents);
      } else {
        Message.error(response.data.status === 'error' ? (response.data as any).msg : 'Failed to load components');
      }
    } catch (error: any) {
      console.error('Failed to fetch components:', error);
      Message.error(error.response?.data?.msg || 'Failed to load components');
    } finally {
      setLoading(false);
    }
  };

  // 加载资产列表（用于资产组件选择）
  const fetchAssets = async () => {
    try {
      const response = await request.get<{ status: string; data: BackendAsset[] }>('/api/asset/list');
      if (response.data.status === 'ok' && response.data.data) {
        setAssets(response.data.data.map(asset => ({ asset_id: asset.asset_id, name: asset.name })));
      }
    } catch (error: any) {
      console.error('Failed to fetch assets:', error);
    }
  };

  // 页面加载时获取组件列表和资产列表
  useEffect(() => {
    fetchComponents();
    fetchAssets();
  }, []);

  // 处理添加组件
  const handleAdd = async () => {
    try {
      const values = await form.validate();
      const response = await request.post<{ status: string; data: BackendToolComponent; msg?: string }>('/api/tool-component', {
        name: values.name,
        description: values.description || '',
        type: values.type,
        asset_id: values.type === 'asset' ? values.asset_id : undefined,
        service_url: values.type === 'service' ? values.service_url : undefined,
        param_desc: values.type === 'service' ? values.param_desc || '' : undefined,
        cron_expression: values.type === 'trigger' ? values.cron_expression : undefined,
      });

      if (response.data.status === 'ok') {
        Message.success(t['toolComponent.addSuccess']);
        setVisible(false);
        form.resetFields();
        fetchComponents();
      } else {
        Message.error(response.data.msg || 'Failed to add component');
      }
    } catch (error: any) {
      if (error.fields) {
        // 表单验证错误
        return;
      }
      console.error('Failed to add component:', error);
      Message.error(error.response?.data?.msg || 'Failed to add component');
    }
  };

  // 处理编辑组件
  const handleEdit = (component: ToolComponent) => {
    setEditingComponent(component);
    editForm.setFieldsValue({
      name: component.name,
      description: component.description || '',
      type: component.type,
      asset_id: component.assetId || '',
      service_url: component.serviceUrl || '',
      param_desc: component.paramDesc || '',
      cron_expression: component.cronExpression || '',
    });
    setEditVisible(true);
  };

  // 处理保存编辑
  const handleSaveEdit = async () => {
    if (!editingComponent) return;

    try {
      const values = await editForm.validate();
      const response = await request.put<{ status: string; data: BackendToolComponent; msg?: string }>(
        `/api/tool-component/${editingComponent.id}`,
        {
          name: values.name,
          description: values.description || '',
          asset_id: editingComponent.type === 'asset' ? values.asset_id : undefined,
          service_url: editingComponent.type === 'service' ? values.service_url : undefined,
          param_desc: editingComponent.type === 'service' ? values.param_desc || '' : undefined,
          cron_expression: editingComponent.type === 'trigger' ? values.cron_expression : undefined,
        }
      );

      if (response.data.status === 'ok') {
        Message.success(t['toolComponent.editSuccess']);
        setEditVisible(false);
        setEditingComponent(null);
        editForm.resetFields();
        fetchComponents();
      } else {
        Message.error(response.data.msg || 'Failed to update component');
      }
    } catch (error: any) {
      if (error.fields) {
        // 表单验证错误
        return;
      }
      console.error('Failed to update component:', error);
      Message.error(error.response?.data?.msg || 'Failed to update component');
    }
  };

  // 处理取消编辑
  const handleCancelEdit = () => {
    setEditVisible(false);
    setEditingComponent(null);
    editForm.resetFields();
  };

  // 处理删除组件
  const handleDelete = async (componentId: string) => {
    try {
      const response = await request.delete<{ status: string; msg?: string }>(`/api/tool-component/${componentId}`);
      
      if (response.data.status === 'ok') {
        setComponents(components.filter((c) => c.id !== componentId));
        Message.success(t['toolComponent.deleteSuccess']);
      } else {
        Message.error(response.data.msg || 'Failed to delete component');
      }
    } catch (error: any) {
      console.error('Failed to delete component:', error);
      Message.error(error.response?.data?.msg || 'Failed to delete component');
    }
  };

  // 表格列定义
  const columns = [
    {
      title: t['toolComponent.componentName'],
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t['toolComponent.componentDescription'],
      dataIndex: 'description',
      key: 'description',
      render: (description: string) => (
        <span style={{ color: description ? 'var(--color-text-1)' : 'var(--color-text-3)' }}>
          {description || '-'}
        </span>
      ),
    },
    {
      title: t['toolComponent.componentType'],
      dataIndex: 'type',
      key: 'type',
      render: (type: ToolComponentType) => {
        let typeLabel = '';
        if (type === 'asset') {
          typeLabel = t['toolComponent.componentType.asset'];
        } else if (type === 'service') {
          typeLabel = t['toolComponent.componentType.service'];
        } else if (type === 'trigger') {
          typeLabel = t['toolComponent.componentType.trigger'];
        }
        return <Tag>{typeLabel}</Tag>;
      },
    },
    {
      title: t['toolComponent.assetId'],
      dataIndex: 'assetId',
      key: 'assetId',
      render: (assetId: string, record: ToolComponent) => (
        <span>{record.type === 'asset' ? assetId || '-' : '-'}</span>
      ),
    },
    {
      title: t['toolComponent.serviceUrl'],
      dataIndex: 'serviceUrl',
      key: 'serviceUrl',
      render: (serviceUrl: string, record: ToolComponent) => (
        <span style={{ maxWidth: '300px', display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {record.type === 'service' ? serviceUrl || '-' : '-'}
        </span>
      ),
    },
    {
      title: t['toolComponent.paramDesc'],
      dataIndex: 'paramDesc',
      key: 'paramDesc',
      render: (paramDesc: string, record: ToolComponent) => (
        <span style={{ maxWidth: '300px', display: 'inline-block', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {record.type === 'service' ? paramDesc || '-' : '-'}
        </span>
      ),
    },
    {
      title: t['toolComponent.actions'],
      key: 'actions',
      render: (_: any, record: ToolComponent) => (
        <Space>
          <Button
            type="text"
            size="small"
            icon={<IconEdit />}
            onClick={() => handleEdit(record)}
          >
            {t['toolComponent.edit']}
          </Button>
          <Popconfirm
            title={t['toolComponent.deleteConfirm']}
            onOk={() => handleDelete(record.id)}
          >
            <Button type="text" status="danger" size="small" icon={<IconDelete />}>
              {t['toolComponent.delete']}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h2>{t['toolComponent.title']}</h2>
        <Button
          type="primary"
          icon={<IconPlus />}
          onClick={() => {
            form.resetFields();
            setVisible(true);
          }}
        >
          {t['toolComponent.addComponent']}
        </Button>
      </div>

      <div className={styles.content}>
        <Spin loading={loading} style={{ width: '100%' }}>
          <Table
            columns={columns}
            data={components}
            pagination={{ pageSize: 10 }}
            noDataElement={<div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-3)' }}>{t['toolComponent.noComponents']}</div>}
          />
        </Spin>
      </div>

      {/* 添加组件弹窗 */}
      <Modal
        title={t['toolComponent.addComponent']}
        visible={visible}
        onCancel={() => {
          setVisible(false);
          form.resetFields();
        }}
        onOk={handleAdd}
        className={styles.addModal}
        okText={t['toolComponent.save']}
        cancelText={t['toolComponent.cancel']}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            label={t['toolComponent.componentName']}
            field="name"
            rules={[{ required: true, message: t['toolComponent.nameRequired'] }]}
          >
            <Input placeholder={t['toolComponent.namePlaceholder']} />
          </Form.Item>
          <Form.Item
            label={t['toolComponent.componentDescription']}
            field="description"
          >
            <TextArea
              placeholder={t['toolComponent.descriptionPlaceholder']}
              autoSize={{ minRows: 2, maxRows: 4 }}
            />
          </Form.Item>
          <Form.Item
            label={t['toolComponent.componentType']}
            field="type"
            rules={[{ required: true, message: t['toolComponent.typeRequired'] }]}
          >
            <RadioGroup>
              <Radio value="asset">{t['toolComponent.componentType.asset']}</Radio>
              <Radio value="service">{t['toolComponent.componentType.service']}</Radio>
              <Radio value="trigger">{t['toolComponent.componentType.trigger']}</Radio>
            </RadioGroup>
          </Form.Item>
          <Form.Item
            noStyle
            shouldUpdate={(prev, next) => prev.type !== next.type}
          >
            {(values) => {
              const type = values.type;
              if (type === 'asset') {
                return (
                  <Form.Item
                    label={t['toolComponent.assetId']}
                    field="asset_id"
                    rules={[{ required: true, message: t['toolComponent.assetIdRequired'] }]}
                  >
                    <Select placeholder={t['toolComponent.selectAsset']}>
                      {assets.map((asset) => (
                        <Option key={asset.asset_id} value={asset.asset_id}>
                          {asset.name} ({asset.asset_id})
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>
                );
              }
              if (type === 'service') {
                return (
                  <>
                    <Form.Item
                      label={t['toolComponent.serviceUrl']}
                      field="service_url"
                      rules={[{ required: true, message: t['toolComponent.serviceUrlRequired'] }]}
                    >
                      <Input placeholder={t['toolComponent.serviceUrlPlaceholder']} />
                    </Form.Item>
                    <Form.Item
                      label={t['toolComponent.paramDesc']}
                      field="param_desc"
                    >
                      <TextArea
                        placeholder={t['toolComponent.paramDescPlaceholder']}
                        autoSize={{ minRows: 3, maxRows: 6 }}
                      />
                    </Form.Item>
                  </>
                );
              }
              if (type === 'trigger') {
                return (
                  <Form.Item
                    label={t['toolComponent.cronExpression']}
                    field="cron_expression"
                    rules={[{ required: true, message: t['toolComponent.cronExpressionRequired'] }]}
                    help={t['toolComponent.cronExpressionHelp']}
                  >
                    <Input placeholder={t['toolComponent.cronExpressionPlaceholder']} />
                  </Form.Item>
                );
              }
              return null;
            }}
          </Form.Item>
        </Form>
      </Modal>

      {/* 编辑组件弹窗 */}
      <Modal
        title={t['toolComponent.editComponent']}
        visible={editVisible}
        onCancel={handleCancelEdit}
        onOk={handleSaveEdit}
        className={styles.addModal}
        okText={t['toolComponent.save']}
        cancelText={t['toolComponent.cancel']}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item
            label={t['toolComponent.componentName']}
            field="name"
            rules={[{ required: true, message: t['toolComponent.nameRequired'] }]}
          >
            <Input placeholder={t['toolComponent.namePlaceholder']} />
          </Form.Item>
          <Form.Item
            label={t['toolComponent.componentDescription']}
            field="description"
          >
            <TextArea
              placeholder={t['toolComponent.descriptionPlaceholder']}
              autoSize={{ minRows: 2, maxRows: 4 }}
            />
          </Form.Item>
          {editingComponent?.type === 'asset' && (
            <Form.Item
              label={t['toolComponent.assetId']}
              field="asset_id"
              rules={[{ required: true, message: t['toolComponent.assetIdRequired'] }]}
            >
              <Select placeholder={t['toolComponent.selectAsset']}>
                {assets.map((asset) => (
                  <Option key={asset.asset_id} value={asset.asset_id}>
                    {asset.name} ({asset.asset_id})
                  </Option>
                ))}
              </Select>
            </Form.Item>
          )}
          {editingComponent?.type === 'service' && (
            <>
              <Form.Item
                label={t['toolComponent.serviceUrl']}
                field="service_url"
                rules={[{ required: true, message: t['toolComponent.serviceUrlRequired'] }]}
              >
                <Input placeholder={t['toolComponent.serviceUrlPlaceholder']} />
              </Form.Item>
              <Form.Item
                label={t['toolComponent.paramDesc']}
                field="param_desc"
              >
                <TextArea
                  placeholder={t['toolComponent.paramDescPlaceholder']}
                  autoSize={{ minRows: 3, maxRows: 6 }}
                />
              </Form.Item>
            </>
          )}
          {editingComponent?.type === 'trigger' && (
            <Form.Item
              label={t['toolComponent.cronExpression']}
              field="cron_expression"
              rules={[{ required: true, message: t['toolComponent.cronExpressionRequired'] }]}
              help={t['toolComponent.cronExpressionHelp']}
            >
              <Input placeholder={t['toolComponent.cronExpressionPlaceholder']} />
            </Form.Item>
          )}
        </Form>
      </Modal>
    </div>
  );
}

export default ToolComponentManagement;

