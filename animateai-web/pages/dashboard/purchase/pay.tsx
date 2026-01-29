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
  paid?: boolean;
  paidAt?: string;
};

export default function PurchasePayPage() {
  const router = useRouter();
  const id = typeof router.query.id === 'string' ? router.query.id : '';

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);
  const [paying, setPaying] = useState(false);
  const [secondsLeft, setSecondsLeft] = useState(120);

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

  useEffect(() => {
    if (!id) return;
    const timer = window.setInterval(() => {
      setSecondsLeft((s) => (s > 0 ? s - 1 : 0));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [id]);

  const pay = async () => {
    if (!id) return;
    try {
      setPaying(true);
      await internalApi.patch('/api/orders', { id, paid: true });
      router.push(`/dashboard/purchase/pay-success?id=${encodeURIComponent(id)}`);
    } catch (e) {
      Message.error('支付失败');
    } finally {
      setPaying(false);
    }
  };

  const expired = secondsLeft <= 0;

  const subtitle = useMemo(() => {
    if (!order) return '加载中...';
    return `订单号：${order.id} · 套餐：${order.planName}`;
  }, [order]);

  if (!id) {
    return (
      <Card>
        <Result status="warning" title="缺少订单号" subTitle="请从购买流程进入本页" />
      </Card>
    );
  }

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          <Typography.Title heading={5} style={{ margin: 0 }}>
            模拟支付
          </Typography.Title>
          <Typography.Text type="secondary">{subtitle}</Typography.Text>
        </Space>
      </Card>

      <Card>
        <Result
          status={expired ? 'warning' : 'info'}
          title={expired ? '支付已超时' : '请完成支付'}
          subTitle={expired ? '请返回重新下单' : `请在 ${secondsLeft}s 内完成支付（纯模拟）`}
          extra={
            <Space>
              <Button type="primary" disabled={expired} loading={paying} onClick={pay}>
                立即支付
              </Button>
              <Button
                type="secondary"
                onClick={() => router.push('/dashboard/purchase')}
              >
                取消
              </Button>
              <Button onClick={fetchOrder} loading={loading}>
                刷新订单
              </Button>
            </Space>
          }
        />
      </Card>
    </Space>
  );
}
