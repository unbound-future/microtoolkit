import React, { useEffect, useMemo, useState } from 'react';
import {
  Card,
  Space,
  Typography,
  Table,
  Button,
  Message,
  Tag,
  Modal,
} from '@arco-design/web-react';
import internalApi from '@/utils/internalApi';

type OrderStatus = 'pending' | 'approved' | 'rejected';

type Order = {
  id: string;
  planKey: string;
  planName: string;
  status: OrderStatus;
  createdAt: string;
  paid?: boolean;
  paidAt?: string;
};

export default function OrderQueryPage() {
  const [data, setData] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);
  const [actingId, setActingId] = useState<string | null>(null);

  const fetchList = async () => {
    try {
      setLoading(true);
      const res = await internalApi.get('/api/orders');
      setData(res?.data || []);
    } catch (e) {
      Message.error('获取订单列表失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchList();
  }, []);

  const updateStatus = async (id: string, status: OrderStatus) => {
    try {
      setActingId(id);
      await internalApi.patch('/api/orders', { id, status });
      Message.success('操作成功');
      await fetchList();
    } catch (e) {
      Message.error('操作失败');
    } finally {
      setActingId(null);
    }
  };

  const columns = useMemo(
    () => [
      {
        title: '订单号',
        dataIndex: 'id',
      },
      {
        title: '套餐',
        dataIndex: 'planName',
      },
      {
        title: '状态',
        dataIndex: 'status',
        render: (v: OrderStatus) => {
          if (v === 'pending') return <Tag color="orange">待审批</Tag>;
          if (v === 'approved') return <Tag color="green">已通过</Tag>;
          return <Tag color="red">已拒绝</Tag>;
        },
      },
      {
        title: '支付',
        dataIndex: 'paid',
        render: (v: boolean) => (v ? <Tag color="green">已支付</Tag> : <Tag color="gray">未支付</Tag>),
      },
      {
        title: '支付时间',
        dataIndex: 'paidAt',
      },
      {
        title: '创建时间',
        dataIndex: 'createdAt',
      },
      {
        title: '操作',
        dataIndex: 'actions',
        render: (_: unknown, row: Order) => {
          const disabled = row.status !== 'pending';
          return (
            <Space>
              <Button
                type="primary"
                size="small"
                disabled={disabled}
                loading={actingId === row.id}
                onClick={() => {
                  Modal.confirm({
                    title: '审批通过',
                    content: `确认通过订单 ${row.id} 吗？`,
                    onOk: () => updateStatus(row.id, 'approved'),
                  });
                }}
              >
                通过
              </Button>
              <Button
                status="danger"
                size="small"
                disabled={disabled}
                loading={actingId === row.id}
                onClick={() => {
                  Modal.confirm({
                    title: '审批拒绝',
                    content: `确认拒绝订单 ${row.id} 吗？`,
                    onOk: () => updateStatus(row.id, 'rejected'),
                  });
                }}
              >
                拒绝
              </Button>
            </Space>
          );
        },
      },
    ],
    [actingId]
  );

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          <Typography.Title heading={5} style={{ margin: 0 }}>
            订单查询
          </Typography.Title>
          <Typography.Text type="secondary">
            管理员查看用户购买申请，并进行审批。
          </Typography.Text>
        </Space>
      </Card>

      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Typography.Text>订单列表</Typography.Text>
          <Button onClick={fetchList} loading={loading}>
            刷新
          </Button>
        </Space>
        <Table
          rowKey="id"
          loading={loading}
          columns={columns}
          data={data}
          pagination={{ pageSize: 10 }}
          style={{ marginTop: 12 }}
        />
      </Card>
    </Space>
  );
}
