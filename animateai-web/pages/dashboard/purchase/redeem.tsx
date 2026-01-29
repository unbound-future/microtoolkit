import React, { useEffect, useMemo, useState } from 'react';
import { Card, Space, Typography, Button, Message, Result } from '@arco-design/web-react';
import { useRouter } from 'next/router';
import internalApi from '@/utils/internalApi';

type OrderStatus = 'pending' | 'approved' | 'rejected';

type Order = {
  id: string;
  planKey: string;
  planName: string;
  status: OrderStatus;
  createdAt: string;
};

export default function PurchaseRedeemPage() {
  const router = useRouter();
  const id = typeof router.query.id === 'string' ? router.query.id : '';

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);
  const [revealed, setRevealed] = useState(false);

  const fetchOrder = async () => {
    if (!id) return;
    try {
      setLoading(true);
      const res = await internalApi.get(`/api/orders?id=${encodeURIComponent(id)}`);
      setOrder(res?.data || null);
    } catch (e) {
      Message.error('获取订单失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchOrder();
  }, [id]);

  const state = useMemo(() => {
    if (!order?.status) return 'loading';
    if (order.status === 'approved') return 'approved';
    if (order.status === 'pending') return 'pending';
    return 'rejected';
  }, [order]);

  if (!id) {
    return (
      <Card>
        <Result status="warning" title="缺少订单号" subTitle="请从购买流程进入本页" />
      </Card>
    );
  }

  if (state === 'pending') {
    return (
      <Space direction="vertical" style={{ width: '100%' }} size={16}>
        <Card>
          <Result
            status="info"
            title="尚未审批"
            subTitle="管理员尚未审批通过，请先返回等待页面。"
            extra={
              <Space>
                <Button onClick={fetchOrder} loading={loading}>
                  刷新
                </Button>
                <Button type="primary" onClick={() => router.push(`/dashboard/purchase/waiting?id=${encodeURIComponent(id)}`)}>
                  返回等待审批
                </Button>
              </Space>
            }
          />
        </Card>
      </Space>
    );
  }

  if (state === 'rejected') {
    return (
      <Card>
        <Result
          status="error"
          title="审批被拒绝"
          subTitle="请重新下单或联系管理员"
          extra={
            <Space>
              <Button type="primary" onClick={() => router.push('/dashboard/purchase')}>
                返回购买
              </Button>
            </Space>
          }
        />
      </Card>
    );
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          <Typography.Title heading={5} style={{ margin: 0 }}>
            账号领取
          </Typography.Title>
          <Typography.Text type="secondary">订单号：{id}</Typography.Text>
        </Space>
      </Card>

      <Card>
        {!revealed ? (
          <Result
            status="success"
            title="审批已通过"
            subTitle="点击下方按钮显示账号（演示）"
            extra={
              <Space>
                <Button
                  type="primary"
                  onClick={() => {
                    setRevealed(true);
                    Message.success('账号已显示');
                  }}
                >
                  显示账号
                </Button>
                <Button onClick={() => router.push('/dashboard/purchase')}>返回</Button>
              </Space>
            }
          />
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size={12}>
            <Typography.Text>
              账号：已显示
            </Typography.Text>
            <Typography.Text type="secondary">
              （演示：这里不展示真实账号，只提示“账号已显示”）
            </Typography.Text>
            <Space>
              <Button type="primary" onClick={() => router.push('/dashboard/purchase')}>
                完成
              </Button>
            </Space>
          </Space>
        )}
      </Card>
    </Space>
  );
}
