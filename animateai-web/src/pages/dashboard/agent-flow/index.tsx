import React, { useState, useCallback, useEffect, useRef } from 'react';
import dynamic from 'next/dynamic';
import { Button, Input, Drawer, Message, Space, Collapse, Tag, Spin, Select, Modal, Table, Popconfirm } from '@arco-design/web-react';
import { IconLock, IconSave, IconPlus, IconArrowLeft, IconDelete, IconEdit } from '@arco-design/web-react/icon';
import type { Node, Edge, Connection, NodeTypes } from 'reactflow';
import { addEdge, useNodesState, useEdgesState } from 'reactflow';
import useLocale from '@/utils/useLocale';
import locale from './locale';
import CustomNode from './custom-node';
import NodeConnectionEditor from '@/components/AgentFlow/NodeConnectionEditor';
import ContextInteractionEditor from '@/components/AgentFlow/ContextInteractionEditor';
import UpstreamCallDescriptionEditor from '@/components/AgentFlow/UpstreamCallDescriptionEditor';
import nodeIdManager from './utils/nodeIdManager';
import { acquireLock, releaseLock, checkLockStatus, clearLocalLock, getCurrentUserId, type LockInfo } from './utils/lockService';
import type { NodeConfig, NodeVariable, NodeConnection, ContextInteractionMode, ComponentInputParam } from './types';
import styles from './style/index.module.less';
import componentStyles from '@/components/AgentFlow/style/components.module.less';
import request from '@/utils/request';

const { Option } = Select;

// 动态导入 ReactFlow 以避免 SSR 问题
const ReactFlow = dynamic(
  () => import('reactflow').then((mod) => mod.ReactFlow),
  { ssr: false }
);

const Background = dynamic(
  () => import('reactflow').then((mod) => mod.Background),
  { ssr: false }
);

const Controls = dynamic(
  () => import('reactflow').then((mod) => mod.Controls),
  { ssr: false }
);

const MiniMap = dynamic(
  () => import('reactflow').then((mod) => mod.MiniMap),
  { ssr: false }
);

// 导入样式（需要在客户端加载）
if (typeof window !== 'undefined') {
  require('reactflow/dist/style.css');
}

const nodeTypes: NodeTypes = {
  custom: CustomNode,
};

// 初始化节点ID管理器
const initNodes: Node[] = [
  {
    id: 'node-0',
    type: 'custom',
    position: { x: 250, y: 100 },
    data: {
      label: '开始节点',
      description: '这是流程的开始节点，你可以拖拽它',
      variables: [
        { name: 'startTime', type: 'string', value: '', description: '开始时间' },
      ],
    } as NodeConfig,
  },
  {
    id: 'node-1',
    type: 'custom',
    position: { x: 500, y: 200 },
    data: {
      label: '处理节点',
      description: '这是处理节点，用于处理业务逻辑',
      connections: [
        {
          targetNodeId: 'node-2',
          logicDescription: '处理完成后进入结束节点',
        },
      ],
    } as NodeConfig,
  },
  {
    id: 'node-2',
    type: 'custom',
    position: { x: 250, y: 300 },
    data: {
      label: '结束节点',
      description: '这是流程的结束节点',
    } as NodeConfig,
  },
];

// 加载现有节点ID到管理器
nodeIdManager.loadFromNodes(initNodes.map((n) => n.id));

const initialNodes: Node[] = initNodes;

const initialEdges: Edge[] = [
  {
    id: 'edge-0-1',
    source: 'node-0',
    target: 'node-1',
    type: 'smoothstep',
  },
  {
    id: 'edge-1-2',
    source: 'node-1',
    target: 'node-2',
    type: 'smoothstep',
  },
];

function AgentFlow() {
  const t = useLocale(locale);
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [visible, setVisible] = useState(false);
  const [editingNodeId, setEditingNodeId] = useState<string | null>(null);
  
  // 节点配置状态
  const [nodeName, setNodeName] = useState('');
  const [nodeDescription, setNodeDescription] = useState('');
  const [nodeAssetId, setNodeAssetId] = useState<string>('');
  const [assets, setAssets] = useState<Array<{ asset_id: string; name: string }>>([]);
  const [nodeComponentId, setNodeComponentId] = useState<string>('');
  const [components, setComponents] = useState<Array<{ component_id: string; name: string; type: string; service_url?: string; param_desc?: string; cron_expression?: string }>>([]);
  const [componentInputParams, setComponentInputParams] = useState<Array<{ name: string; value: string; description?: string }>>([]);
  const [selectedComponent, setSelectedComponent] = useState<{ component_id: string; type: string; param_desc?: string } | null>(null);
  const [upstreamCallDescriptions, setUpstreamCallDescriptions] = useState<string[]>([]);
  const [contextInteractionMode, setContextInteractionMode] = useState<ContextInteractionMode>('full');
  const [nodeVariables, setNodeVariables] = useState<NodeVariable[]>([]);
  const [nodeConnections, setNodeConnections] = useState<NodeConnection[]>([]);
  
  const [mounted, setMounted] = useState(false);
  
  // 锁定状态
  const [lockInfo, setLockInfo] = useState<LockInfo>({ locked: false });
  const [isLocking, setIsLocking] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const lockIdRef = useRef<string | undefined>(undefined);
  
  // 工作流管理状态
  const [viewMode, setViewMode] = useState<'list' | 'edit'>('list'); // 视图模式：列表视图或编辑视图
  const [currentFlowId, setCurrentFlowId] = useState<string | null>(null);
  const [currentFlowName, setCurrentFlowName] = useState<string>('');
  const [workflows, setWorkflows] = useState<Array<{ flow_id: string; name: string; asset_id?: string; template_id?: string; created_at: string; updated_at: string }>>([]);
  const [workflowListVisible, setWorkflowListVisible] = useState(false);
  const [createWorkflowVisible, setCreateWorkflowVisible] = useState(false);
  const [newWorkflowName, setNewWorkflowName] = useState('');
  const [isSavingWorkflow, setIsSavingWorkflow] = useState(false);
  const [isLoadingWorkflow, setIsLoadingWorkflow] = useState(false);
  const [isLoadingWorkflowList, setIsLoadingWorkflowList] = useState(false);
  
  // 页面卸载时释放锁
  useEffect(() => {
    return () => {
      if (lockIdRef.current) {
        console.log('[AgentFlow] useEffect cleanup - 页面卸载，释放锁', {
          lockId: lockIdRef.current,
        });
        releaseLock(lockIdRef.current).catch((error) => {
          console.error('[AgentFlow] useEffect cleanup - 释放锁失败', error);
        });
        clearLocalLock();
      }
    };
  }, []);

  useEffect(() => {
    setMounted(true);
    // 加载现有节点ID到管理器
    nodeIdManager.loadFromNodes(nodes.map((n) => n.id));
  }, [nodes]);

  // 加载资产列表
  useEffect(() => {
    const fetchAssets = async () => {
      try {
        const response = await request.get<{ status: string; data: Array<{ asset_id: string; name: string }> }>('/api/asset/list');
        if (response.data.status === 'ok' && response.data.data) {
          setAssets(response.data.data.map(asset => ({ asset_id: asset.asset_id, name: asset.name })));
        }
      } catch (error) {
        console.error('Failed to fetch assets:', error);
      }
    };
    fetchAssets();
  }, []);

  // 加载工具组件列表（用于组件选择）
  useEffect(() => {
    const fetchComponents = async () => {
      try {
        const response = await request.get<{ status: string; data: Array<{ component_id: string; name: string; type: string; service_url?: string; param_desc?: string; cron_expression?: string }> }>('/api/tool-component/list');
        if (response.data.status === 'ok' && response.data.data) {
          setComponents(response.data.data.map(comp => ({
            component_id: comp.component_id,
            name: comp.name,
            type: comp.type,
            service_url: comp.service_url,
            param_desc: comp.param_desc,
            cron_expression: comp.cron_expression,
          })));
        }
      } catch (error) {
        console.error('Failed to fetch components:', error);
      }
    };
    fetchComponents();
  }, []);

  // 加载工作流列表
  const fetchWorkflows = useCallback(async () => {
    setIsLoadingWorkflowList(true);
    try {
      const response = await request.get<{ status: string; data: Array<{ flow_id: string; name: string; asset_id?: string; template_id?: string; created_at: string; updated_at: string }> }>('/api/agent-flow/list');
      if (response.data.status === 'ok' && response.data.data) {
        setWorkflows(response.data.data);
      } else {
        Message.error(response.data.msg || t['agentFlow.loadWorkflowListFailed'] || '加载工作流列表失败');
      }
    } catch (error: any) {
      console.error('Failed to fetch workflows:', error);
      Message.error(error.response?.data?.msg || t['agentFlow.loadWorkflowListFailed'] || '加载工作流列表失败');
    } finally {
      setIsLoadingWorkflowList(false);
    }
  }, [t]);

  // 页面加载时获取工作流列表
  useEffect(() => {
    if (viewMode === 'list') {
      fetchWorkflows();
    }
  }, [viewMode, fetchWorkflows]);

  // 创建新工作流
  const handleCreateNewWorkflow = () => {
    // 清空当前工作流，创建新工作流
    setCurrentFlowId(null);
    setCurrentFlowName('');
    setNodes(initNodes);
    setEdges(initialEdges);
    nodeIdManager.reset();
    nodeIdManager.loadFromNodes(initNodes.map((n) => n.id));
    setViewMode('edit');
    Message.success(t['agentFlow.newWorkflowCreated'] || '新工作流已创建');
  };

  // 确认创建工作流名称（用于保存时输入名称）
  const handleConfirmCreateWorkflowName = async () => {
    if (!newWorkflowName.trim()) {
      Message.warning(t['agentFlow.workflowNameRequired'] || '请输入工作流名称');
      return;
    }
    setCurrentFlowName(newWorkflowName.trim());
    setCreateWorkflowVisible(false);
    setNewWorkflowName('');
    // 继续保存工作流
    await handleSaveWorkflowInternal();
  };

  // 生成唯一ID
  const generateUniqueID = (): string => {
    const timestamp = Date.now();
    const random = Math.random().toString(36).substring(2, 15);
    return `${timestamp}_${random}`;
  };

  // 内部保存工作流方法（不检查名称）
  const handleSaveWorkflowInternal = async () => {
    setIsSavingWorkflow(true);
    try {
      const flowData = {
        nodes: nodes,
        edges: edges,
      };

      // 提取资产ID（从节点中获取第一个关联的资产ID）
      let assetId = '';
      for (const node of nodes) {
        const nodeData = node.data as NodeConfig;
        if (nodeData.assetId) {
          assetId = nodeData.assetId;
          break;
        }
      }

      // 如果没有资产ID，生成一个唯一的资产ID
      if (!assetId) {
        assetId = generateUniqueID();
      }

      // 生成唯一的模版ID
      const templateId = generateUniqueID();

      if (currentFlowId) {
        // 更新现有工作流（保留原有的模版ID和资产ID，如果原来没有则使用新的）
        const response = await request.put<{ status: string; msg?: string }>(
          `/api/agent-flow/${currentFlowId}`,
          {
            name: currentFlowName,
            asset_id: assetId,
            template_id: templateId,
            flow_data: flowData,
          }
        );
        if (response.data.status === 'ok') {
          Message.success(t['agentFlow.workflowSaved'] || '工作流保存成功');
          await fetchWorkflows(); // 刷新列表
        } else {
          Message.error(response.data.msg || t['agentFlow.saveWorkflowFailed'] || '保存工作流失败');
        }
      } else {
        // 创建新工作流（必须生成唯一的模版ID和资产ID）
        const response = await request.post<{ status: string; data?: { flow_id: string }; msg?: string }>(
          '/api/agent-flow',
          {
            name: currentFlowName,
            asset_id: assetId,
            template_id: templateId,
            flow_data: flowData,
          }
        );
        if (response.data.status === 'ok' && response.data.data) {
          setCurrentFlowId(response.data.data.flow_id);
          Message.success(t['agentFlow.workflowSaved'] || '工作流保存成功');
          await fetchWorkflows(); // 刷新列表
        } else {
          Message.error(response.data.msg || t['agentFlow.saveWorkflowFailed'] || '保存工作流失败');
        }
      }
    } catch (error: any) {
      console.error('Failed to save workflow:', error);
      Message.error(error.response?.data?.msg || t['agentFlow.saveWorkflowFailed'] || '保存工作流失败');
    } finally {
      setIsSavingWorkflow(false);
    }
  };

  // 返回列表视图
  const handleBackToList = () => {
    setViewMode('list');
    setCurrentFlowId(null);
    setCurrentFlowName('');
  };

  // 保存工作流（已废弃，使用 handleSaveWorkflowInternal）
  const handleSaveWorkflow = async () => {
    if (!currentFlowName.trim()) {
      // 如果名称为空，弹出输入框让用户输入
      setNewWorkflowName('');
      setCreateWorkflowVisible(true);
      return;
    }
    await handleSaveWorkflowInternal();
  };

  // 加载工作流并进入编辑视图
  const handleLoadWorkflow = async (flowId: string) => {
    setIsLoadingWorkflow(true);
    try {
      const response = await request.get<{ status: string; data?: { flow_id: string; name: string; flow_data: any }; msg?: string }>(
        `/api/agent-flow/${flowId}`
      );
      if (response.data.status === 'ok' && response.data.data) {
        const flowData = response.data.data.flow_data;
        if (flowData && flowData.nodes && flowData.edges) {
          setNodes(flowData.nodes);
          setEdges(flowData.edges);
          // 重新加载节点ID到管理器
          nodeIdManager.reset();
          nodeIdManager.loadFromNodes(flowData.nodes.map((n: Node) => n.id));
          setCurrentFlowId(flowId);
          setCurrentFlowName(response.data.data.name);
          setViewMode('edit'); // 切换到编辑视图
          Message.success(t['agentFlow.workflowLoaded'] || '工作流加载成功');
        } else {
          Message.error(t['agentFlow.invalidWorkflowData'] || '工作流数据格式无效');
        }
      } else {
        Message.error(response.data.msg || t['agentFlow.loadWorkflowFailed'] || '加载工作流失败');
      }
    } catch (error: any) {
      console.error('Failed to load workflow:', error);
      Message.error(error.response?.data?.msg || t['agentFlow.loadWorkflowFailed'] || '加载工作流失败');
    } finally {
      setIsLoadingWorkflow(false);
    }
  };

  // 删除工作流
  const handleDeleteWorkflow = async (e: React.MouseEvent, flowId: string) => {
    e.stopPropagation(); // 阻止事件冒泡，防止触发卡片点击
    try {
      const response = await request.delete<{ status: string; msg?: string }>(`/api/agent-flow/${flowId}`);
      if (response.data.status === 'ok') {
        Message.success(t['agentFlow.workflowDeleted'] || '工作流删除成功');
        await fetchWorkflows(); // 刷新列表
        // 如果删除的是当前工作流，返回列表视图
        if (currentFlowId === flowId) {
          handleBackToList();
        }
      } else {
        Message.error(response.data.msg || t['agentFlow.deleteWorkflowFailed'] || '删除工作流失败');
      }
    } catch (error: any) {
      console.error('Failed to delete workflow:', error);
      Message.error(error.response?.data?.msg || t['agentFlow.deleteWorkflowFailed'] || '删除工作流失败');
    }
  };

  const onConnect = useCallback(
    (params: Connection) => {
      setEdges((eds) => addEdge(params, eds));
    },
    [setEdges]
  );

  const resetForm = () => {
      setNodeName('');
      setNodeDescription('');
      setNodeAssetId('');
      setNodeComponentId('');
      setComponentInputParams([]);
      setSelectedComponent(null);
      setUpstreamCallDescriptions([]);
      setContextInteractionMode('full');
      setNodeVariables([]);
      setNodeConnections([]);
      setEditingNodeId(null);
  };

  const handleAddNode = async () => {
    console.log('[AgentFlow] handleAddNode - 开始添加节点', {
      visible,
      lockInfo,
      lockIdRef: lockIdRef.current,
    });

    // 严格检查：如果编辑面板已打开，不允许添加新节点
    if (visible) {
      console.log('[AgentFlow] handleAddNode - 编辑面板已打开，阻止操作');
      Message.warning(t['agentFlow.alreadyEditing'] || '当前正在编辑中，请先完成或取消当前编辑');
      return;
    }

    // 严格检查：如果已经持有锁，不允许添加新节点
    if (lockInfo.locked && lockIdRef.current) {
      console.log('[AgentFlow] handleAddNode - 已持有锁，阻止操作', {
        lockInfo,
        lockIdRef: lockIdRef.current,
      });
      Message.warning(t['agentFlow.alreadyEditing'] || '当前正在编辑中，请先完成或取消当前编辑');
      return;
    }

    // 检查服务器端的锁状态（防止并发问题）
    console.log('[AgentFlow] handleAddNode - 检查服务器端锁状态');
    const currentLockStatus = await checkLockStatus();
    console.log('[AgentFlow] handleAddNode - 服务器端锁状态检查结果', {
      currentLockStatus,
      localLockInfo: lockInfo,
      localLockId: lockIdRef.current,
    });
    if (currentLockStatus.locked) {
      // 如果服务器端显示已锁定
      const currentUserId = getCurrentUserId();
      // 检查是否是当前用户持有的锁
      if (currentLockStatus.lockedBy === currentUserId && currentLockStatus.lockId === lockIdRef.current) {
        // 如果是当前用户持有的锁，说明正在编辑中，不允许添加新节点
        Message.warning(t['agentFlow.alreadyEditing'] || '当前正在编辑中，请先完成或取消当前编辑');
      } else {
        // 被其他用户锁定
        Message.warning(
          currentLockStatus.lockedBy
            ? t['agentFlow.lockedByOther']?.replace('{user}', currentLockStatus.lockedBy) || `已被 ${currentLockStatus.lockedBy} 锁定`
            : t['agentFlow.lockFailed'] || '获取编辑锁失败'
        );
      }
      // 同步本地状态
      if (currentLockStatus.lockId) {
        setLockInfo({
          locked: true,
          lockedBy: currentLockStatus.lockedBy,
          lockedAt: currentLockStatus.lockedAt,
          lockId: currentLockStatus.lockId,
        });
        lockIdRef.current = currentLockStatus.lockId;
      }
      return;
    }

    // 尝试获取锁（新建节点也需要锁定，避免其他用户同时编辑）
    setIsLocking(true);
    try {
      console.log('[AgentFlow] handleAddNode - 尝试获取锁');
      const lock = await acquireLock();
      console.log('[AgentFlow] handleAddNode - 获取锁结果', {
        lock,
        localLockInfo: lockInfo,
        localLockId: lockIdRef.current,
      });
      if (lock.locked) {
        console.log('[AgentFlow] handleAddNode - 成功获取锁，更新本地状态', {
          lock,
        });
        setLockInfo(lock);
        lockIdRef.current = lock.lockId;
        resetForm();
        setEditingNodeId(null);
        setVisible(true);
      } else {
        console.log('[AgentFlow] handleAddNode - 获取锁失败', {
          lock,
        });
        Message.warning(
          lock.lockedBy
            ? t['agentFlow.lockedByOther']?.replace('{user}', lock.lockedBy) || `已被 ${lock.lockedBy} 锁定`
            : t['agentFlow.lockFailed'] || '获取编辑锁失败'
        );
        // 同步锁状态
        if (lock.lockedBy) {
          setLockInfo({
            locked: false,
            lockedBy: lock.lockedBy,
            lockedAt: lock.lockedAt,
          });
        }
      }
    } catch (error) {
      Message.error(t['agentFlow.lockError'] || '获取编辑锁时出错');
    } finally {
      setIsLocking(false);
    }
  };

  const handleEditNode = async (nodeId: string) => {
    console.log('[AgentFlow] handleEditNode - 开始编辑节点', {
      nodeId,
      editingNodeId,
      lockInfo,
      lockIdRef: lockIdRef.current,
    });

    const node = nodes.find((n) => n.id === nodeId);
    if (!node) {
      console.log('[AgentFlow] handleEditNode - 节点不存在', { nodeId });
      return;
    }

    // 严格检查：如果已经持有锁，不允许编辑其他节点
    if (lockInfo.locked && lockIdRef.current) {
      // 如果正在编辑的节点不同，提示用户
      if (editingNodeId !== nodeId) {
        console.log('[AgentFlow] handleEditNode - 已持有锁但节点不同，阻止操作', {
          editingNodeId,
          nodeId,
        });
        Message.warning(t['agentFlow.alreadyEditing'] || '当前正在编辑中，请先完成或取消当前编辑');
        return;
      }
      // 如果是同一个节点，直接打开编辑面板（可能只是重新点击）
      const data = node.data as NodeConfig;
      setNodeName(data.label || '');
      setNodeDescription(data.description || '');
      setNodeAssetId(data.assetId || '');
      if (data.upstreamCallDescriptions && data.upstreamCallDescriptions.length > 0) {
        setUpstreamCallDescriptions(data.upstreamCallDescriptions);
      } else if ((data as any).externalCallDescription) {
        setUpstreamCallDescriptions([(data as any).externalCallDescription]);
      } else {
        setUpstreamCallDescriptions([]);
      }
      setContextInteractionMode(data.contextInteractionMode || 'full');
      setNodeVariables(data.variables || []);
      if (data.connections && data.connections.length > 0) {
        setNodeConnections(data.connections);
      } else {
        const oldDownstreamIds = (data as any).downstreamNodeIds || (data as any).logicSelector?.downstreamNodeIds || [];
        if (oldDownstreamIds.length > 0) {
          setNodeConnections(
            oldDownstreamIds.map((nodeId: string) => ({
              targetNodeId: nodeId,
              logicDescription: '',
            }))
          );
        } else {
          setNodeConnections([]);
        }
      }
      setVisible(true);
      return;
    }

    // 检查服务器端的锁状态
    console.log('[AgentFlow] handleEditNode - 检查服务器端锁状态');
    const currentLockStatus = await checkLockStatus();
    console.log('[AgentFlow] handleEditNode - 服务器端锁状态检查结果', {
      currentLockStatus,
      localLockInfo: lockInfo,
      localLockId: lockIdRef.current,
    });
    if (currentLockStatus.locked) {
      console.log('[AgentFlow] handleEditNode - 服务器端已锁定，阻止操作');
      Message.warning(
        currentLockStatus.lockedBy
          ? t['agentFlow.lockedByOther']?.replace('{user}', currentLockStatus.lockedBy) || `已被 ${currentLockStatus.lockedBy} 锁定`
          : t['agentFlow.lockFailed'] || '获取编辑锁失败'
      );
      if (currentLockStatus.lockId) {
        setLockInfo({
          locked: true,
          lockedBy: currentLockStatus.lockedBy,
          lockedAt: currentLockStatus.lockedAt,
          lockId: currentLockStatus.lockId,
        });
        lockIdRef.current = currentLockStatus.lockId;
      }
      return;
    }

    // 先尝试获取锁
    setIsLocking(true);
    try {
      console.log('[AgentFlow] handleEditNode - 尝试获取锁');
      const lock = await acquireLock();
      console.log('[AgentFlow] handleEditNode - 获取锁结果', {
        lock,
        localLockInfo: lockInfo,
        localLockId: lockIdRef.current,
      });
      if (!lock.locked) {
        Message.warning(
          lock.lockedBy
            ? t['agentFlow.lockedByOther']?.replace('{user}', lock.lockedBy) || `已被 ${lock.lockedBy} 锁定`
            : t['agentFlow.lockFailed'] || '获取编辑锁失败'
        );
        // 同步锁状态
        if (lock.lockedBy) {
          setLockInfo({
            locked: false,
            lockedBy: lock.lockedBy,
            lockedAt: lock.lockedAt,
          });
        }
        setIsLocking(false);
        return;
      }

      // 锁定成功，加载节点数据
      console.log('[AgentFlow] handleEditNode - 成功获取锁，更新本地状态', {
        lock,
      });
      setLockInfo(lock);
      lockIdRef.current = lock.lockId;

      const data = node.data as NodeConfig;
      setNodeName(data.label || '');
      setNodeDescription(data.description || '');
      setNodeAssetId(data.assetId || '');
      setNodeComponentId(data.componentId || '');
      setComponentInputParams(data.componentInputParams || []);
      
      // 如果选择了组件，加载组件信息
      if (data.componentId) {
        const comp = components.find(c => c.component_id === data.componentId);
        if (comp) {
          setSelectedComponent({
            component_id: comp.component_id,
            type: comp.type,
            param_desc: comp.param_desc,
          });
        } else {
          setSelectedComponent(null);
        }
      } else {
        setSelectedComponent(null);
      }
      
      // 兼容旧数据：如果有 externalCallDescription，转换为数组
      if (data.upstreamCallDescriptions && data.upstreamCallDescriptions.length > 0) {
        setUpstreamCallDescriptions(data.upstreamCallDescriptions);
      } else if ((data as any).externalCallDescription) {
        setUpstreamCallDescriptions([(data as any).externalCallDescription]);
      } else {
        setUpstreamCallDescriptions([]);
      }
      // 加载上下文交互模式，默认为全量传递
      setContextInteractionMode(data.contextInteractionMode || 'full');
      setNodeVariables(data.variables || []);
      
      // 加载节点关联关系，兼容旧数据结构
      if (data.connections && data.connections.length > 0) {
        setNodeConnections(data.connections);
      } else {
        // 兼容旧数据：如果有旧的downstreamNodeIds，转换为新格式
        const oldDownstreamIds = (data as any).downstreamNodeIds || (data as any).logicSelector?.downstreamNodeIds || [];
        if (oldDownstreamIds.length > 0) {
          setNodeConnections(
            oldDownstreamIds.map((nodeId: string) => ({
              targetNodeId: nodeId,
              logicDescription: '',
            }))
          );
        } else {
          setNodeConnections([]);
        }
      }
      
      setEditingNodeId(nodeId);
      setVisible(true);
    } catch (error) {
      Message.error(t['agentFlow.lockError'] || '获取编辑锁时出错');
    } finally {
      setIsLocking(false);
    }
  };

  const handleConfirm = async () => {
    console.log('[AgentFlow] handleConfirm - 开始保存', {
      nodeName,
      editingNodeId,
      lockInfo,
      lockIdRef: lockIdRef.current,
    });

    if (!nodeName.trim()) {
      console.log('[AgentFlow] handleConfirm - 节点名称为空，阻止保存');
      Message.warning(t['agentFlow.nodeNameRequired']);
      return;
    }

    // 检查本地锁状态
    if (!lockInfo.locked || !lockIdRef.current) {
      console.log('[AgentFlow] handleConfirm - 本地锁状态检查失败，阻止保存', {
        lockInfo,
        lockIdRef: lockIdRef.current,
      });
      Message.warning(t['agentFlow.noLock'] || '没有编辑权限，请重新打开编辑');
      return;
    }

    // 保存前再次验证服务器端的锁状态，确保锁仍然有效
    console.log('[AgentFlow] handleConfirm - 第一次验证服务器端锁状态');
    const currentLockStatus = await checkLockStatus();
    console.log('[AgentFlow] handleConfirm - 第一次锁状态验证结果', {
      currentLockStatus,
      localLockInfo: lockInfo,
      localLockId: lockIdRef.current,
    });
    if (!currentLockStatus.locked) {
      // 锁已经被释放
      Message.warning(t['agentFlow.noLock'] || '编辑锁已失效，请重新打开编辑');
      setLockInfo({ locked: false });
      lockIdRef.current = undefined;
      setVisible(false);
      resetForm();
      return;
    }

    // 验证锁ID是否匹配
    if (currentLockStatus.lockId !== lockIdRef.current) {
      console.log('[AgentFlow] handleConfirm - 锁ID不匹配，阻止保存', {
        serverLockId: currentLockStatus.lockId,
        localLockId: lockIdRef.current,
        currentLockStatus,
      });
      // 锁已经被其他用户获取
      Message.warning(
        currentLockStatus.lockedBy
          ? t['agentFlow.lockedByOther']?.replace('{user}', currentLockStatus.lockedBy) || `已被 ${currentLockStatus.lockedBy} 锁定`
          : t['agentFlow.noLock'] || '编辑锁已被其他用户获取'
      );
      setLockInfo({
        locked: true,
        lockedBy: currentLockStatus.lockedBy,
        lockedAt: currentLockStatus.lockedAt,
        lockId: currentLockStatus.lockId,
      });
      lockIdRef.current = currentLockStatus.lockId;
      setVisible(false);
      resetForm();
      return;
    }

    // 验证是否是当前用户持有的锁
    const currentUserId = getCurrentUserId();
    console.log('[AgentFlow] handleConfirm - 验证锁所有者', {
      currentUserId,
      serverLockedBy: currentLockStatus.lockedBy,
    });
    if (currentLockStatus.lockedBy && currentLockStatus.lockedBy !== currentUserId) {
      console.log('[AgentFlow] handleConfirm - 锁所有者不匹配，阻止保存', {
        currentUserId,
        serverLockedBy: currentLockStatus.lockedBy,
      });
      // 锁被其他用户持有
      Message.warning(
        t['agentFlow.lockedByOther']?.replace('{user}', currentLockStatus.lockedBy) || `已被 ${currentLockStatus.lockedBy} 锁定`
      );
      setLockInfo({
        locked: true,
        lockedBy: currentLockStatus.lockedBy,
        lockedAt: currentLockStatus.lockedAt,
        lockId: currentLockStatus.lockId,
      });
      lockIdRef.current = currentLockStatus.lockId;
      setVisible(false);
      resetForm();
      return;
    }

    setIsSaving(true);
    try {
      const nodeConfig: NodeConfig = {
        label: nodeName,
        description: nodeDescription,
        assetId: nodeAssetId || undefined,
        componentId: nodeComponentId || undefined,
        componentInputParams: componentInputParams.length > 0 ? componentInputParams : undefined,
        upstreamCallDescriptions: upstreamCallDescriptions.length > 0 ? upstreamCallDescriptions : undefined,
        contextInteractionMode: contextInteractionMode || 'full',
        // 只有在增量传递模式下才保存变量
        variables: contextInteractionMode === 'incremental' && nodeVariables.length > 0 ? nodeVariables : undefined,
        connections: nodeConnections.length > 0 ? nodeConnections : undefined,
      };

      // 在保存之前，最后一次验证锁状态（防止在验证和保存之间的时间窗口内锁状态发生变化）
      console.log('[AgentFlow] handleConfirm - 最后一次验证锁状态（保存前）');
      const finalLockCheck = await checkLockStatus();
      console.log('[AgentFlow] handleConfirm - 最终锁状态验证结果', {
        finalLockCheck,
        localLockId: lockIdRef.current,
      });
      if (!finalLockCheck.locked || finalLockCheck.lockId !== lockIdRef.current) {
        console.log('[AgentFlow] handleConfirm - 最终验证失败，锁状态已改变', {
          finalLockCheck,
          localLockId: lockIdRef.current,
        });
        // 锁状态已改变，不允许保存
        Message.error(t['agentFlow.lockChangedDuringSave'] || '保存过程中锁状态已改变，保存失败');
        setLockInfo({ locked: false });
        lockIdRef.current = undefined;
        setVisible(false);
        resetForm();
        return;
      }

      // 验证锁的所有者
      const finalCurrentUserId = getCurrentUserId();
      console.log('[AgentFlow] handleConfirm - 验证最终锁所有者', {
        finalCurrentUserId,
        finalLockCheckLockedBy: finalLockCheck.lockedBy,
      });
      if (finalLockCheck.lockedBy && finalLockCheck.lockedBy !== finalCurrentUserId) {
        console.log('[AgentFlow] handleConfirm - 最终锁所有者验证失败', {
          finalCurrentUserId,
          finalLockCheckLockedBy: finalLockCheck.lockedBy,
        });
        Message.error(
          t['agentFlow.lockedByOther']?.replace('{user}', finalLockCheck.lockedBy) || `已被 ${finalLockCheck.lockedBy} 锁定，保存失败`
        );
        setLockInfo({
          locked: true,
          lockedBy: finalLockCheck.lockedBy,
          lockedAt: finalLockCheck.lockedAt,
          lockId: finalLockCheck.lockId,
        });
        lockIdRef.current = finalLockCheck.lockId;
        setVisible(false);
        resetForm();
        return;
      }

      // TODO: 调用后端 API 保存数据
      // await axios.post('/api/agent-flow/save', { nodeConfig, nodeId: editingNodeId });

      if (editingNodeId) {
        // 编辑现有节点
        setNodes((nds) =>
          nds.map((node) =>
            node.id === editingNodeId
              ? { ...node, data: { ...node.data, ...nodeConfig } }
              : node
          )
        );
        Message.success(t['agentFlow.nodeUpdated']);
      } else {
        // 创建新节点，使用全局ID管理器生成唯一ID
        const newNodeId = nodeIdManager.generateId();
        const newNode: Node = {
          id: newNodeId,
          type: 'custom',
          position: {
            x: Math.random() * 400,
            y: Math.random() * 400,
          },
          data: nodeConfig,
        };
        setNodes((nds) => [...nds, newNode]);
        Message.success(t['agentFlow.nodeCreated']);
      }

      console.log('[AgentFlow] handleConfirm - 保存成功', {
        currentLockId: lockIdRef.current,
        lockInfo,
      });

      // 保存成功后关闭编辑面板（不立即释放锁，锁在面板关闭时释放）
      // 关闭编辑面板
      setVisible(false);
      resetForm();
      console.log('[AgentFlow] handleConfirm - 保存完成，面板已关闭');
    } catch (error) {
      Message.error(t['agentFlow.saveError'] || '保存失败，请重试');
    } finally {
      setIsSaving(false);
    }
  };

  const handleCancel = async () => {
    console.log('[AgentFlow] handleCancel - 开始取消操作', {
      lockInfo,
      lockIdRef: lockIdRef.current,
    });

    // 取消时关闭编辑面板（不立即释放锁，锁在面板关闭时释放）
    setVisible(false);
    resetForm();
    console.log('[AgentFlow] handleCancel - 取消完成，面板已关闭');
  };

  // 当编辑面板关闭时，释放锁
  useEffect(() => {
    if (!visible) {
      const currentLockId = lockIdRef.current;
      if (currentLockId) {
        console.log('[AgentFlow] useEffect visible change - 编辑面板已关闭，释放锁', {
          currentLockId,
        });
        
        // 异步释放锁
        releaseLock(currentLockId)
          .then(() => {
            console.log('[AgentFlow] useEffect visible change - 锁已释放');
          })
          .catch((error) => {
            console.error('[AgentFlow] useEffect visible change - 释放锁失败:', error);
          })
          .finally(() => {
            // 无论是否成功，都清除本地状态
            console.log('[AgentFlow] useEffect visible change - 清除本地锁状态', {
              previousLockId: lockIdRef.current,
            });
            setLockInfo({ locked: false });
            lockIdRef.current = undefined;
            console.log('[AgentFlow] useEffect visible change - 本地锁状态已清除');
          });
      }
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible]);

  // 处理节点点击事件（单击选择节点，打开编辑面板）
  const handleNodeClick = useCallback(
    (event: React.MouseEvent, node: Node) => {
      handleEditNode(node.id);
    },
    [nodes] // 依赖 nodes 以获取最新数据
  );

  // 工作流列表视图
  if (viewMode === 'list') {
    return (
      <div className={styles.container}>
        <div className={styles.workflowListContainer}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h2 style={{ margin: 0, fontSize: '20px', fontWeight: 600 }}>
              {t['agentFlow.workflowList'] || '工作流列表'}
            </h2>
            <Button
              type="primary"
              icon={<IconPlus />}
              onClick={handleCreateNewWorkflow}
            >
              {t['agentFlow.createNewWorkflow'] || '创建新工作流'}
            </Button>
          </div>
          <Spin loading={isLoadingWorkflowList}>
            <Table
              data={workflows}
              rowKey="flow_id"
              style={{ width: '100%', minWidth: '100%' }}
              onRow={(record) => ({
                onClick: () => handleLoadWorkflow(record.flow_id),
                style: { cursor: 'pointer' },
              })}
              columns={[
                {
                  title: t['agentFlow.workflowName'] || '工作流名称',
                  dataIndex: 'name',
                  key: 'name',
                  render: (name: string) => (
                    <span style={{ fontWeight: 500 }}>{name}</span>
                  ),
                },
                {
                  title: t['agentFlow.assetId'] || '资产ID',
                  dataIndex: 'asset_id',
                  key: 'asset_id',
                  render: (assetId: string) => assetId ? (
                    <Tag color="blue" size="small">{assetId}</Tag>
                  ) : (
                    <span style={{ color: 'var(--color-text-3)' }}>-</span>
                  ),
                },
                {
                  title: t['agentFlow.templateId'] || '模版ID',
                  dataIndex: 'template_id',
                  key: 'template_id',
                  render: (templateId: string) => templateId ? (
                    <Tag color="green" size="small">{templateId}</Tag>
                  ) : (
                    <span style={{ color: 'var(--color-text-3)' }}>-</span>
                  ),
                },
                {
                  title: t['agentFlow.createdAt'] || '创建时间',
                  dataIndex: 'created_at',
                  key: 'created_at',
                  render: (createdAt: string) => new Date(createdAt).toLocaleString(),
                },
                {
                  title: t['agentFlow.updatedAt'] || '更新时间',
                  dataIndex: 'updated_at',
                  key: 'updated_at',
                  render: (updatedAt: string) => new Date(updatedAt).toLocaleString(),
                },
                {
                  title: t['agentFlow.actions'] || '操作',
                  key: 'actions',
                  render: (_: any, record: { flow_id: string }) => (
                    <Space>
                      <Button
                        type="text"
                        size="small"
                        icon={<IconEdit />}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleLoadWorkflow(record.flow_id);
                        }}
                      >
                        {t['agentFlow.edit'] || '编辑'}
                      </Button>
                      <Popconfirm
                        title={t['agentFlow.deleteWorkflowConfirm'] || '确定要删除这个工作流吗？'}
                        onOk={(e) => handleDeleteWorkflow(e as any, record.flow_id)}
                      >
                        <Button
                          type="text"
                          size="small"
                          status="danger"
                          icon={<IconDelete />}
                          onClick={(e) => e.stopPropagation()}
                        >
                          {t['agentFlow.delete'] || '删除'}
                        </Button>
                      </Popconfirm>
                    </Space>
                  ),
                },
              ]}
              pagination={{
                pageSize: 10,
                showTotal: (total) => `${t['agentFlow.total'] || '共'} ${total} ${t['agentFlow.items'] || '条'}`,
              }}
            />
          </Spin>
        </div>
      </div>
    );
  }

  // 编辑视图
  return (
    <div className={styles.container}>
      <div className={styles.toolbar}>
        <Space>
          <Button 
            icon={<IconArrowLeft />}
            onClick={handleBackToList}
            disabled={isSavingWorkflow || isLoadingWorkflow}
          >
            {t['agentFlow.backToList'] || '返回列表'}
          </Button>
          <Button 
            icon={<IconSave />}
            onClick={handleSaveWorkflow}
            loading={isSavingWorkflow}
            disabled={isSavingWorkflow || isLoadingWorkflow || !currentFlowName.trim()}
          >
            {t['agentFlow.saveWorkflow'] || '保存工作流'}
          </Button>
          {currentFlowName ? (
            <Tag color="green">
              {t['agentFlow.currentWorkflow'] || '当前工作流'}: {currentFlowName}
            </Tag>
          ) : (
            <Tag color="orange">
              {t['agentFlow.noWorkflow'] || '未命名工作流'}
            </Tag>
          )}
          <Button 
            type="primary" 
            onClick={handleAddNode} 
            loading={isLocking}
            disabled={isLocking || isSaving || visible || (lockInfo.locked && !!lockIdRef.current)}
          >
            {t['agentFlow.addNode']}
          </Button>
          {lockInfo.locked && (
            <Tag color="blue" icon={<IconLock />}>
              {t['agentFlow.lockedByYou'] || '已锁定'} ({lockInfo.lockedBy})
            </Tag>
          )}
        </Space>
      </div>
      <div className={styles.flowContainer}>
        {mounted && (
          <ReactFlow
            nodes={nodes}
            edges={edges}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onNodeClick={handleNodeClick}
            nodeTypes={nodeTypes}
            fitView
            nodesDraggable={!visible || !lockInfo.locked}
            nodesConnectable={!visible || !lockInfo.locked}
            elementsSelectable={true}
            panOnDrag={[1, 2]}
            defaultViewport={{ x: 0, y: 0, zoom: 1 }}
          >
            <Background color="#aaa" gap={16} />
            <Controls />
            <MiniMap />
          </ReactFlow>
        )}
      </div>
      <Drawer
        width={480}
        title={
          <Space>
            {editingNodeId ? t['agentFlow.editNode'] : t['agentFlow.createNode']}
            {lockInfo.locked && (
              <Tag color="blue" size="small" icon={<IconLock />}>
                {t['agentFlow.locked'] || '已锁定'}
              </Tag>
            )}
          </Space>
        }
        visible={visible}
        onCancel={handleCancel}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 12 }}>
            <Button onClick={handleCancel} disabled={isSaving}>
              {t['agentFlow.cancel']}
            </Button>
            <Button type="primary" onClick={handleConfirm} loading={isSaving}>
              {editingNodeId ? t['agentFlow.save'] : t['agentFlow.createNode']}
            </Button>
          </div>
        }
        placement="right"
        className={styles.nodeEditorDrawer}
      >
        <Collapse
          defaultActiveKey={['basic']}
          expandIconPosition="right"
          accordion={false}
        >
          <Collapse.Item
            header={t['agentFlow.basicInfo']}
            name="basic"
          >
            <Space direction="vertical" size="medium" style={{ width: '100%' }}>
              <div className={styles.formItem}>
                <label>{t['agentFlow.nodeName']}：</label>
                <Input
                  value={nodeName}
                  onChange={(value) => setNodeName(value)}
                  placeholder={t['agentFlow.nodeNamePlaceholder']}
                />
              </div>
              <div className={styles.formItem}>
                <label>{t['agentFlow.nodeDescription']}：</label>
                <Input.TextArea
                  value={nodeDescription}
                  onChange={(value) => setNodeDescription(value)}
                  placeholder={t['agentFlow.nodeDescriptionPlaceholder']}
                  rows={4}
                />
              </div>
              <div className={styles.formItem}>
                <label>{t['agentFlow.assetAssociation'] || '资产关联'}：</label>
                <Select
                  value={nodeAssetId || undefined}
                  onChange={(value) => setNodeAssetId(value || '')}
                  placeholder={t['agentFlow.selectAsset'] || '请选择资产（可选）'}
                  allowClear
                  showSearch
                  filterOption={(inputValue, option) => {
                    const label = option?.props?.children || '';
                    return String(label).toLowerCase().indexOf(inputValue.toLowerCase()) >= 0;
                  }}
                >
                  {assets.map((asset) => (
                    <Option key={asset.asset_id} value={asset.asset_id}>
                      {asset.name} ({asset.asset_id})
                    </Option>
                  ))}
                </Select>
              </div>
              <div className={styles.formItem}>
                <label>{t['agentFlow.componentSelection'] || '组件选择'}：</label>
                <Select
                  value={nodeComponentId || undefined}
                  onChange={(value) => {
                    setNodeComponentId(value || '');
                    if (value) {
                      const comp = components.find(c => c.component_id === value);
                      if (comp) {
                        setSelectedComponent({
                          component_id: comp.component_id,
                          type: comp.type,
                          param_desc: comp.param_desc,
                        });
                        // 解析参数说明，初始化输入参数
                        if (comp.param_desc) {
                          // 尝试从参数说明中提取参数名（简单解析，例如：query: string, limit: number）
                          const paramNames = comp.param_desc.split(',').map(p => {
                            const parts = p.trim().split(':');
                            return parts[0].trim();
                          }).filter(Boolean);
                          setComponentInputParams(
                            paramNames.map(name => ({
                              name,
                              value: '',
                              description: comp.param_desc,
                            }))
                          );
                        } else {
                          setComponentInputParams([]);
                        }
                      }
                    } else {
                      setSelectedComponent(null);
                      setComponentInputParams([]);
                    }
                  }}
                  placeholder={t['agentFlow.selectComponent'] || '请选择组件（可选）'}
                  allowClear
                  showSearch
                  filterOption={(inputValue, option) => {
                    const label = option?.props?.children || '';
                    return String(label).toLowerCase().indexOf(inputValue.toLowerCase()) >= 0;
                  }}
                >
                  {components.map((comp) => (
                    <Option key={comp.component_id} value={comp.component_id}>
                      {comp.name} ({comp.type === 'asset' ? '资产' : comp.type === 'service' ? '服务' : comp.type === 'trigger' ? '触发器' : comp.type})
                    </Option>
                  ))}
                </Select>
              </div>
              {selectedComponent && (
                <div className={styles.formItem}>
                  <label>{t['agentFlow.componentInputParams'] || '组件输入参数'}：</label>
                  <Space direction="vertical" size="small" style={{ width: '100%' }}>
                    {selectedComponent.type === 'trigger' ? (
                      <span style={{ color: 'var(--color-text-3)', fontSize: '12px' }}>
                        {t['agentFlow.triggerNoParams'] || '时间触发器无需输入参数'}
                      </span>
                    ) : componentInputParams.length > 0 ? (
                      <>
                        {componentInputParams.map((param, index) => (
                          <div key={index} style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
                            <Input
                              style={{ flex: 1 }}
                              addBefore={param.name}
                              value={param.value}
                              onChange={(value) => {
                                const newParams = [...componentInputParams];
                                newParams[index] = { ...newParams[index], value };
                                setComponentInputParams(newParams);
                              }}
                              placeholder={t['agentFlow.paramValuePlaceholder'] || '请输入参数值'}
                            />
                            <Button
                              type="text"
                              status="danger"
                              size="small"
                              icon={<IconDelete />}
                              onClick={() => {
                                const newParams = componentInputParams.filter((_, i) => i !== index);
                                setComponentInputParams(newParams);
                              }}
                            />
                          </div>
                        ))}
                        <Button
                          type="dashed"
                          size="small"
                          icon={<IconPlus />}
                          onClick={() => {
                            const paramName = prompt(t['agentFlow.addParamName'] || '请输入参数名：');
                            if (paramName && paramName.trim()) {
                              setComponentInputParams([
                                ...componentInputParams,
                                {
                                  name: paramName.trim(),
                                  value: '',
                                  description: selectedComponent.param_desc,
                                },
                              ]);
                            }
                          }}
                          style={{ width: '100%' }}
                        >
                          {t['agentFlow.addParam'] || '添加参数'}
                        </Button>
                      </>
                    ) : (
                      <>
                        <span style={{ color: 'var(--color-text-3)', fontSize: '12px' }}>
                          {t['agentFlow.noComponentParams'] || '该组件无需输入参数'}
                        </span>
                        <Button
                          type="dashed"
                          size="small"
                          icon={<IconPlus />}
                          onClick={() => {
                            const paramName = prompt(t['agentFlow.addParamName'] || '请输入参数名：');
                            if (paramName && paramName.trim()) {
                              setComponentInputParams([
                                {
                                  name: paramName.trim(),
                                  value: '',
                                  description: selectedComponent.param_desc,
                                },
                              ]);
                            }
                          }}
                          style={{ width: '100%' }}
                        >
                          {t['agentFlow.addParam'] || '添加参数'}
                        </Button>
                      </>
                    )}
                    {selectedComponent.param_desc && (
                      <div style={{ fontSize: '12px', color: 'var(--color-text-3)', marginTop: '4px' }}>
                        {t['agentFlow.paramDesc'] || '参数说明'}：{selectedComponent.param_desc}
                      </div>
                    )}
                  </Space>
                </div>
              )}
            </Space>
          </Collapse.Item>

          <Collapse.Item
            header={t['agentFlow.upstreamCallDescriptions'] || '上游调用描述'}
            name="upstreamCallDescriptions"
          >
            <UpstreamCallDescriptionEditor
              descriptions={upstreamCallDescriptions}
              onChange={setUpstreamCallDescriptions}
            />
          </Collapse.Item>

          <Collapse.Item
            header={t['agentFlow.contextInteraction'] || '上下文交互'}
            name="contextInteraction"
          >
            <ContextInteractionEditor
              mode={contextInteractionMode}
              variables={nodeVariables}
              onModeChange={setContextInteractionMode}
              onVariablesChange={setNodeVariables}
            />
          </Collapse.Item>

          <Collapse.Item
            header={t['agentFlow.nodeConnections'] || '节点关联'}
            name="connections"
          >
            <NodeConnectionEditor
              key={`node-editor-${editingNodeId || 'new'}-${nodes.length}`}
              nodes={nodes}
              connections={nodeConnections}
              onChange={setNodeConnections}
              excludeNodeId={editingNodeId || undefined}
            />
          </Collapse.Item>
        </Collapse>
      </Drawer>

    </div>
  );
}

export default AgentFlow;


