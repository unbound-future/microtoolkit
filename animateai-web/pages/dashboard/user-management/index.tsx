import React, { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Form,
  Input,
  Message,
  Modal,
  Select,
  Space,
  Table,
  Typography,
} from '@arco-design/web-react';
import internalApi from '@/utils/internalApi';

type UserRow = {
  id: string;
  userName: string;
  role: string;
  nickname?: string | null;
  email?: string | null;
  phone?: string | null;
  createdAt: string;
  updatedAt: string;
};

type FormValue = {
  userName: string;
  password?: string;
  role: 'admin' | 'user';
  nickname?: string;
  email?: string;
  phone?: string;
};

export default function UserManagementPage() {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<UserRow[]>([]);
  const [q, setQ] = useState('');

  const [createVisible, setCreateVisible] = useState(false);
  const [editVisible, setEditVisible] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [current, setCurrent] = useState<UserRow | null>(null);

  const [createForm] = Form.useForm<FormValue>();
  const [editForm] = Form.useForm<FormValue>();

  const fetchList = async () => {
    try {
      setLoading(true);
      const res = await internalApi.get('/api/admin/users', {
        params: q ? { q } : undefined,
      });
      setData(res?.data || []);
    } catch (e) {
      Message.error('获取用户列表失败（需要管理员权限）');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, []);

  const columns = useMemo(
    () => [
      { title: '用户名', dataIndex: 'userName' },
      {
        title: '角色',
        dataIndex: 'role',
        render: (v: string) => (v === 'admin' ? '管理员' : '用户'),
      },
      { title: '昵称', dataIndex: 'nickname' },
      { title: '邮箱', dataIndex: 'email' },
      { title: '手机', dataIndex: 'phone' },
      { title: '创建时间', dataIndex: 'createdAt' },
      { title: '更新时间', dataIndex: 'updatedAt' },
      {
        title: '操作',
        dataIndex: 'actions',
        render: (_: unknown, row: UserRow) => (
          <Space>
            <Button
              size="small"
              onClick={() => {
                setCurrent(row);
                editForm.setFieldsValue({
                  userName: row.userName,
                  role: (row.role === 'admin' ? 'admin' : 'user') as any,
                  nickname: row.nickname || '',
                  email: row.email || '',
                  phone: row.phone || '',
                });
                setEditVisible(true);
              }}
            >
              编辑
            </Button>
            <Button
              size="small"
              status="danger"
              onClick={() => {
                Modal.confirm({
                  title: '删除用户',
                  content: `确认删除用户 ${row.userName} 吗？`,
                  onOk: async () => {
                    try {
                      await internalApi.delete('/api/admin/users', {
                        params: { id: row.id },
                      });
                      Message.success('删除成功');
                      fetchList();
                    } catch {
                      Message.error('删除失败');
                    }
                  },
                });
              }}
            >
              删除
            </Button>
          </Space>
        ),
      },
    ],
    [editForm]
  );

  const onCreate = async () => {
    try {
      const values = await createForm.validate();
      setSubmitting(true);
      await internalApi.post('/api/admin/users', values);
      Message.success('创建成功');
      setCreateVisible(false);
      createForm.resetFields();
      fetchList();
    } catch (e) {
      if ((e as any)?.errorFields) return;
      Message.error('创建失败');
    } finally {
      setSubmitting(false);
    }
  };

  const onEdit = async () => {
    if (!current?.id) return;
    try {
      const values = await editForm.validate();
      setSubmitting(true);
      await internalApi.patch('/api/admin/users', {
        id: current.id,
        ...values,
      });
      Message.success('更新成功');
      setEditVisible(false);
      setCurrent(null);
      fetchList();
    } catch (e) {
      if ((e as any)?.errorFields) return;
      Message.error('更新失败');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          <Typography.Title heading={5} style={{ margin: 0 }}>
            用户管理
          </Typography.Title>
          <Typography.Text type="secondary">
            管理员对用户信息进行增删改查。接口基于本地 SQLite（dev.db）与 Prisma。
          </Typography.Text>
        </Space>
      </Card>

      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Space>
            <Input
              style={{ width: 280 }}
              value={q}
              onChange={setQ}
              placeholder="搜索 userName / nickname / email / phone"
              allowClear
            />
            <Button onClick={fetchList} loading={loading}>
              搜索/刷新
            </Button>
          </Space>
          <Button type="primary" onClick={() => setCreateVisible(true)}>
            新增用户
          </Button>
        </Space>

        <Table
          rowKey="id"
          loading={loading}
          columns={columns as any}
          data={data}
          pagination={{ pageSize: 10 }}
          style={{ marginTop: 12 }}
        />
      </Card>

      <Modal
        title="新增用户"
        visible={createVisible}
        onCancel={() => setCreateVisible(false)}
        onOk={onCreate}
        confirmLoading={submitting}
      >
        <Form form={createForm} layout="vertical" initialValues={{ role: 'user' }}>
          <Form.Item
            field="userName"
            label="用户名"
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input />
          </Form.Item>
          <Form.Item
            field="password"
            label="密码"
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password />
          </Form.Item>
          <Form.Item field="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={[
                { label: '用户', value: 'user' },
                { label: '管理员', value: 'admin' },
              ]}
            />
          </Form.Item>
          <Form.Item field="nickname" label="昵称">
            <Input />
          </Form.Item>
          <Form.Item field="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item field="phone" label="手机">
            <Input />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="编辑用户"
        visible={editVisible}
        onCancel={() => {
          setEditVisible(false);
          setCurrent(null);
        }}
        onOk={onEdit}
        confirmLoading={submitting}
      >
        <Form form={editForm} layout="vertical">
          <Form.Item field="userName" label="用户名">
            <Input disabled />
          </Form.Item>
          <Form.Item field="password" label="重置密码（可选）">
            <Input.Password placeholder="不填则不修改" />
          </Form.Item>
          <Form.Item field="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={[
                { label: '用户', value: 'user' },
                { label: '管理员', value: 'admin' },
              ]}
            />
          </Form.Item>
          <Form.Item field="nickname" label="昵称">
            <Input />
          </Form.Item>
          <Form.Item field="email" label="邮箱">
            <Input />
          </Form.Item>
          <Form.Item field="phone" label="手机">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </Space>
  );
}
