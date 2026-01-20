import React, { useState, useCallback, useEffect } from 'react';
import { Button, Input, Message, Space, Tag, Spin, Select, Modal, Table, Popconfirm } from '@arco-design/web-react';
import { IconSave, IconPlus, IconArrowLeft, IconDelete, IconEdit } from '@arco-design/web-react/icon';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import styles from './style/index.module.less';
import request from '@/utils/request';
import type { WorkflowTemplate, BackendWorkflowTemplate } from './types';

const { Option } = Select;
const { TextArea } = Input;

function WorkflowTemplateManagement() {
  const t = useLocale(locale);
  const [viewMode, setViewMode] = useState<'list' | 'edit'>('list'); // 视图模式：列表视图或编辑视图
  const [currentTemplateId, setCurrentTemplateId] = useState<string | null>(null);
  const [currentTemplateName, setCurrentTemplateName] = useState<string>('');
  const [currentTemplateDescription, setCurrentTemplateDescription] = useState<string>('');
  const [templates, setTemplates] = useState<Array<{ template_id: string; name: string; description?: string; asset_id?: string; created_at: string; updated_at: string }>>([]);
  const [isSavingTemplate, setIsSavingTemplate] = useState(false);
  const [isLoadingTemplate, setIsLoadingTemplate] = useState(false);
  const [isLoadingTemplateList, setIsLoadingTemplateList] = useState(false);
  
  // 模版数据（从工作流选择或编辑）
  const [templateData, setTemplateData] = useState<any>(null);
  
  // 工作流列表（用于从工作流创建模版）
  const [workflows, setWorkflows] = useState<Array<{ flow_id: string; name: string }>>([]);
  const [selectedWorkflowId, setSelectedWorkflowId] = useState<string>('');
  const [createTemplateVisible, setCreateTemplateVisible] = useState(false);
  const [newTemplateName, setNewTemplateName] = useState('');
  const [newTemplateDescription, setNewTemplateDescription] = useState('');

  // 生成唯一ID
  const generateUniqueID = (): string => {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(2, 15);
    return `${timestamp}_${random}`;
  };

  // 加载模版列表
  const fetchTemplates = useCallback(async () => {
    setIsLoadingTemplateList(true);
    try {
      const response = await request.get<{ status: string; data: Array<BackendWorkflowTemplate> }>('/api/workflow-template/list');
      if (response.data.status === 'ok' && response.data.data) {
        setTemplates(response.data.data.map(t => ({
          template_id: t.template_id,
          name: t.name,
          description: t.description,
          asset_id: t.asset_id,
          created_at: t.created_at,
          updated_at: t.updated_at,
        })));
      } else {
        Message.error(response.data.msg || t['workflowTemplate.loadTemplateListFailed'] || '加载模版列表失败');
      }
    } catch (error: any) {
      console.error('Failed to fetch templates:', error);
      Message.error(error.response?.data?.msg || t['workflowTemplate.loadTemplateListFailed'] || '加载模版列表失败');
    } finally {
      setIsLoadingTemplateList(false);
    }
  }, [t]);

  // 加载工作流列表（用于从工作流创建模版）
  const fetchWorkflows = useCallback(async () => {
    try {
      const response = await request.get<{ status: string; data: Array<{ flow_id: string; name: string }> }>('/api/agent-flow/list');
      if (response.data.status === 'ok' && response.data.data) {
        setWorkflows(response.data.data.map(w => ({
          flow_id: w.flow_id,
          name: w.name,
        })));
      }
    } catch (error) {
      console.error('Failed to fetch workflows:', error);
    }
  }, []);

  // 页面加载时获取模版列表和工作流列表
  useEffect(() => {
    if (viewMode === 'list') {
      fetchTemplates();
    }
    fetchWorkflows();
  }, [viewMode, fetchTemplates, fetchWorkflows]);

  // 创建新模版（从工作流选择）
  const handleCreateNewTemplate = () => {
    setSelectedWorkflowId('');
    setNewTemplateName('');
    setNewTemplateDescription('');
    setCreateTemplateVisible(true);
  };

  // 确认从工作流创建模版
  const handleConfirmCreateFromWorkflow = async () => {
    if (!selectedWorkflowId) {
      Message.warning(t['workflowTemplate.selectWorkflow'] || '请选择工作流');
      return;
    }
    if (!newTemplateName.trim()) {
      Message.warning(t['workflowTemplate.templateNameRequired'] || '请输入模版名称');
      return;
    }

    setIsLoadingTemplate(true);
    try {
      // 加载选中的工作流
      const response = await request.get<{ status: string; data?: { flow_data: any } }>(
        `/api/agent-flow/${selectedWorkflowId}`
      );
      if (response.data.status === 'ok' && response.data.data) {
        const flowData = response.data.data.flow_data;
        if (flowData) {
          // 使用工作流数据创建模版
          setTemplateData(flowData);
          setCurrentTemplateId(null);
          setCurrentTemplateName(newTemplateName.trim());
          setCurrentTemplateDescription(newTemplateDescription.trim());
          setCreateTemplateVisible(false);
          setViewMode('edit');
          Message.success(t['workflowTemplate.newTemplateCreated'] || '新模版已创建');
        } else {
          Message.error(t['workflowTemplate.invalidTemplateData'] || '工作流数据格式无效');
        }
      } else {
        Message.error(response.data.msg || t['workflowTemplate.loadTemplateFailed'] || '加载工作流失败');
      }
    } catch (error: any) {
      console.error('Failed to load workflow:', error);
      Message.error(error.response?.data?.msg || t['workflowTemplate.loadTemplateFailed'] || '加载工作流失败');
    } finally {
      setIsLoadingTemplate(false);
    }
  };

  // 返回列表视图
  const handleBackToList = () => {
    setViewMode('list');
    setCurrentTemplateId(null);
    setCurrentTemplateName('');
    setCurrentTemplateDescription('');
    setTemplateData(null);
  };

  // 内部保存模版方法
  const handleSaveTemplateInternal = async () => {
    if (!templateData) {
      Message.warning(t['workflowTemplate.invalidTemplateData'] || '模版数据无效');
      return;
    }

    setIsSavingTemplate(true);
    try {
      // 提取资产ID（从模版数据中获取）
      let assetId = '';
      if (templateData.nodes && Array.isArray(templateData.nodes)) {
        for (const node of templateData.nodes) {
          if (node.data && node.data.assetId) {
            assetId = node.data.assetId;
            break;
          }
        }
      }

      // 如果没有资产ID，生成一个唯一的资产ID
      if (!assetId) {
        assetId = generateUniqueID();
      }

      if (currentTemplateId) {
        // 更新现有模版
        const response = await request.put<{ status: string; msg?: string }>(
          `/api/workflow-template/${currentTemplateId}`,
          {
            name: currentTemplateName,
            description: currentTemplateDescription,
            asset_id: assetId,
            template_data: templateData,
          }
        );
        if (response.data.status === 'ok') {
          Message.success(t['workflowTemplate.templateSaved'] || '模版保存成功');
          await fetchTemplates(); // 刷新列表
        } else {
          Message.error(response.data.msg || t['workflowTemplate.saveTemplateFailed'] || '保存模版失败');
        }
      } else {
        // 创建新模版
        const response = await request.post<{ status: string; data?: { template_id: string }; msg?: string }>(
          '/api/workflow-template',
          {
            name: currentTemplateName,
            description: currentTemplateDescription,
            asset_id: assetId,
            template_data: templateData,
          }
        );
        if (response.data.status === 'ok' && response.data.data) {
          setCurrentTemplateId(response.data.data.template_id);
          Message.success(t['workflowTemplate.templateSaved'] || '模版保存成功');
          await fetchTemplates(); // 刷新列表
        } else {
          Message.error(response.data.msg || t['workflowTemplate.saveTemplateFailed'] || '保存模版失败');
        }
      }
    } catch (error: any) {
      console.error('Failed to save template:', error);
      Message.error(error.response?.data?.msg || t['workflowTemplate.saveTemplateFailed'] || '保存模版失败');
    } finally {
      setIsSavingTemplate(false);
    }
  };

  // 保存模版
  const handleSaveTemplate = async () => {
    if (!currentTemplateName.trim()) {
      // 如果名称为空，弹出输入框让用户输入
      setNewTemplateName(currentTemplateName);
      setNewTemplateDescription(currentTemplateDescription);
      setCreateTemplateVisible(true);
      return;
    }
    await handleSaveTemplateInternal();
  };

  // 确认创建模版名称（用于保存时输入名称）
  const handleConfirmCreateTemplateName = async () => {
    if (!newTemplateName.trim()) {
      Message.warning(t['workflowTemplate.templateNameRequired'] || '请输入模版名称');
      return;
    }
    setCurrentTemplateName(newTemplateName.trim());
    setCurrentTemplateDescription(newTemplateDescription.trim());
    setCreateTemplateVisible(false);
    setNewTemplateName('');
    setNewTemplateDescription('');
    // 继续保存模版
    await handleSaveTemplateInternal();
  };

  // 加载模版并进入编辑视图
  const handleLoadTemplate = async (templateId: string) => {
    setIsLoadingTemplate(true);
    try {
      const response = await request.get<{ status: string; data?: { template_id: string; name: string; description?: string; template_data: any }; msg?: string }>(
        `/api/workflow-template/${templateId}`
      );
      if (response.data.status === 'ok' && response.data.data) {
        const templateData = response.data.data.template_data;
        if (templateData) {
          setTemplateData(templateData);
          setCurrentTemplateId(templateId);
          setCurrentTemplateName(response.data.data.name);
          setCurrentTemplateDescription(response.data.data.description || '');
          setViewMode('edit'); // 切换到编辑视图
          Message.success(t['workflowTemplate.templateLoaded'] || '模版加载成功');
        } else {
          Message.error(t['workflowTemplate.invalidTemplateData'] || '模版数据格式无效');
        }
      } else {
        Message.error(response.data.msg || t['workflowTemplate.loadTemplateFailed'] || '加载模版失败');
      }
    } catch (error: any) {
      console.error('Failed to load template:', error);
      Message.error(error.response?.data?.msg || t['workflowTemplate.loadTemplateFailed'] || '加载模版失败');
    } finally {
      setIsLoadingTemplate(false);
    }
  };

  // 删除模版
  const handleDeleteTemplate = async (e: React.MouseEvent, templateId: string) => {
    e.stopPropagation(); // 阻止事件冒泡，防止触发行点击
    try {
      const response = await request.delete<{ status: string; msg?: string }>(`/api/workflow-template/${templateId}`);
      if (response.data.status === 'ok') {
        Message.success(t['workflowTemplate.templateDeleted'] || '模版删除成功');
        await fetchTemplates(); // 刷新列表
        // 如果删除的是当前模版，返回列表视图
        if (currentTemplateId === templateId) {
          handleBackToList();
        }
      } else {
        Message.error(response.data.msg || t['workflowTemplate.deleteTemplateFailed'] || '删除模版失败');
      }
    } catch (error: any) {
      console.error('Failed to delete template:', error);
      Message.error(error.response?.data?.msg || t['workflowTemplate.deleteTemplateFailed'] || '删除模版失败');
    }
  };

  // 模版列表视图
  if (viewMode === 'list') {
    return (
      <div className={styles.container}>
        <div className={styles.templateListContainer}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>
              {t['workflowTemplate.templateList'] || '模版列表'}
            </h2>
            <Button
              type="primary"
              icon={<IconPlus />}
              onClick={handleCreateNewTemplate}
            >
              {t['workflowTemplate.createNewTemplate'] || '创建新模版'}
            </Button>
          </div>
          <Spin loading={isLoadingTemplateList}>
            <Table
              data={templates}
              rowKey="template_id"
              style={{ width: '100%', minWidth: '100%' }}
              onRow={(record) => ({
                onClick: () => handleLoadTemplate(record.template_id),
                style: { cursor: 'pointer' },
              })}
              columns={[
                {
                  title: t['workflowTemplate.templateName'] || '模版名称',
                  dataIndex: 'name',
                  key: 'name',
                  render: (name: string) => (
                    <span style={{ fontWeight: 500 }}>{name}</span>
                  ),
                },
                {
                  title: t['workflowTemplate.templateDescription'] || '模版描述',
                  dataIndex: 'description',
                  key: 'description',
                  render: (description: string) => description || (
                    <span style={{ color: 'var(--color-text-3)' }}>-</span>
                  ),
                },
                {
                  title: t['workflowTemplate.assetId'] || '资产ID',
                  dataIndex: 'asset_id',
                  key: 'asset_id',
                  render: (assetId: string) => assetId ? (
                    <Tag color="blue" size="small">{assetId}</Tag>
                  ) : (
                    <span style={{ color: 'var(--color-text-3)' }}>-</span>
                  ),
                },
                {
                  title: t['workflowTemplate.createdAt'] || '创建时间',
                  dataIndex: 'created_at',
                  key: 'created_at',
                  render: (createdAt: string) => new Date(createdAt).toLocaleString(),
                },
                {
                  title: t['workflowTemplate.updatedAt'] || '更新时间',
                  dataIndex: 'updated_at',
                  key: 'updated_at',
                  render: (updatedAt: string) => new Date(updatedAt).toLocaleString(),
                },
                {
                  title: t['workflowTemplate.actions'] || '操作',
                  key: 'actions',
                  render: (_: any, record: { template_id: string }) => (
                    <Space>
                      <Button
                        type="text"
                        size="small"
                        icon={<IconEdit />}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleLoadTemplate(record.template_id);
                        }}
                      >
                        {t['workflowTemplate.edit'] || '编辑'}
                      </Button>
                      <Popconfirm
                        title={t['workflowTemplate.deleteTemplateConfirm'] || '确定要删除这个模版吗？'}
                        onOk={(e) => handleDeleteTemplate(e as any, record.template_id)}
                      >
                        <Button
                          type="text"
                          size="small"
                          status="danger"
                          icon={<IconDelete />}
                          onClick={(e) => e.stopPropagation()}
                        >
                          {t['workflowTemplate.delete'] || '删除'}
                        </Button>
                      </Popconfirm>
                    </Space>
                  ),
                },
              ]}
              pagination={{
                pageSize: 10,
                showTotal: (total) => `${t['workflowTemplate.total'] || '共'} ${total} ${t['workflowTemplate.items'] || '条'}`,
              }}
            />
          </Spin>
        </div>

        {/* 创建新模版 Modal（从工作流选择） */}
        <Modal
          title={t['workflowTemplate.createNewTemplate'] || '创建新模版'}
          visible={createTemplateVisible}
          onOk={handleConfirmCreateFromWorkflow}
          onCancel={() => setCreateTemplateVisible(false)}
          okText={t['workflowTemplate.confirm'] || '确认'}
          cancelText={t['workflowTemplate.cancel'] || '取消'}
          confirmLoading={isLoadingTemplate}
        >
          <div className={styles.formItem}>
            <label>{t['workflowTemplate.selectWorkflow'] || '选择工作流'}：</label>
            <Select
              value={selectedWorkflowId}
              onChange={(value) => setSelectedWorkflowId(value)}
              placeholder={t['workflowTemplate.selectWorkflowPlaceholder'] || '请选择工作流以创建模版'}
              style={{ width: '100%' }}
            >
              {workflows.map((workflow) => (
                <Option key={workflow.flow_id} value={workflow.flow_id}>
                  {workflow.name}
                </Option>
              ))}
            </Select>
          </div>
          <div className={styles.formItem}>
            <label>{t['workflowTemplate.templateName'] || '模版名称'}：</label>
            <Input
              value={newTemplateName}
              onChange={(value) => setNewTemplateName(value)}
              placeholder={t['workflowTemplate.templateNamePlaceholder'] || '请输入模版名称'}
            />
          </div>
          <div className={styles.formItem}>
            <label>{t['workflowTemplate.templateDescription'] || '模版描述'}：</label>
            <TextArea
              value={newTemplateDescription}
              onChange={(value) => setNewTemplateDescription(value)}
              placeholder={t['workflowTemplate.templateDescriptionPlaceholder'] || '请输入模版描述（可选）'}
              rows={4}
            />
          </div>
        </Modal>
      </div>
    );
  }

  // 编辑视图（显示模版信息，可以编辑名称和描述）
  return (
    <div className={styles.container}>
      <div style={{ padding: '16px', borderBottom: '1px solid var(--color-border-2)' }}>
        <Space>
          <Button 
            icon={<IconArrowLeft />}
            onClick={handleBackToList}
            disabled={isSavingTemplate || isLoadingTemplate}
          >
            {t['workflowTemplate.backToList'] || '返回列表'}
          </Button>
          <Button 
            icon={<IconSave />}
            onClick={handleSaveTemplate}
            loading={isSavingTemplate}
            disabled={isSavingTemplate || isLoadingTemplate || !currentTemplateName.trim()}
          >
            {t['workflowTemplate.saveTemplate'] || '保存模版'}
          </Button>
          {currentTemplateName ? (
            <Tag color="green">
              {t['workflowTemplate.currentTemplate'] || '当前模版'}: {currentTemplateName}
            </Tag>
          ) : (
            <Tag color="orange">
              {t['workflowTemplate.noTemplate'] || '未命名模版'}
            </Tag>
          )}
        </Space>
      </div>
      <div style={{ padding: '24px', overflowY: 'auto', height: '100%' }}>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div className={styles.formItem}>
            <label>{t['workflowTemplate.templateName'] || '模版名称'}：</label>
            <Input
              value={currentTemplateName}
              onChange={(value) => setCurrentTemplateName(value)}
              placeholder={t['workflowTemplate.templateNamePlaceholder'] || '请输入模版名称'}
            />
          </div>
          <div className={styles.formItem}>
            <label>{t['workflowTemplate.templateDescription'] || '模版描述'}：</label>
            <TextArea
              value={currentTemplateDescription}
              onChange={(value) => setCurrentTemplateDescription(value)}
              placeholder={t['workflowTemplate.templateDescriptionPlaceholder'] || '请输入模版描述（可选）'}
              rows={6}
            />
          </div>
          <div className={styles.formItem}>
            <label>模版数据预览：</label>
            <div style={{ 
              padding: '16px', 
              background: 'var(--color-fill-1)', 
              borderRadius: '4px',
              maxHeight: '400px',
              overflow: 'auto',
              fontFamily: 'monospace',
              fontSize: '12px'
            }}>
              <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                {JSON.stringify(templateData, null, 2)}
              </pre>
            </div>
          </div>
        </Space>
      </div>

      {/* 创建模版名称 Modal（用于保存时输入名称） */}
      <Modal
        title={t['workflowTemplate.templateName'] || '模版名称'}
        visible={createTemplateVisible && viewMode === 'edit'}
        onOk={handleConfirmCreateTemplateName}
        onCancel={() => setCreateTemplateVisible(false)}
        okText={t['workflowTemplate.confirm'] || '确认'}
        cancelText={t['workflowTemplate.cancel'] || '取消'}
      >
        <div className={styles.formItem}>
          <label>{t['workflowTemplate.templateName'] || '模版名称'}：</label>
          <Input
            value={newTemplateName}
            onChange={(value) => setNewTemplateName(value)}
            placeholder={t['workflowTemplate.templateNamePlaceholder'] || '请输入模版名称'}
            onPressEnter={handleConfirmCreateTemplateName}
          />
        </div>
        <div className={styles.formItem}>
          <label>{t['workflowTemplate.templateDescription'] || '模版描述'}：</label>
          <TextArea
            value={newTemplateDescription}
            onChange={(value) => setNewTemplateDescription(value)}
            placeholder={t['workflowTemplate.templateDescriptionPlaceholder'] || '请输入模版描述（可选）'}
            rows={4}
          />
        </div>
      </Modal>
    </div>
  );
}

export default WorkflowTemplateManagement;
