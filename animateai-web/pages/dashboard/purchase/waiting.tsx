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

export default function PurchaseWaitingPage() {
  const router = useRouter();
  const id = typeof router.query.id === 'string' ? router.query.id : '';

  const [order, setOrder] = useState<Order | null>(null);
  const [loading, setLoading] = useState(false);
  const [polling, setPolling] = useState(true);

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
    if (!polling || !id) return;
    const timer = window.setInterval(() => {
      fetchOrder();
    }, 2500);
    return () => window.clearInterval(timer);
  }, [polling, id]);

  const viewState = useMemo(() => {
    const status = order?.status;
    if (!status || status === 'pending') return 'pending';
    if (status === 'approved') return 'approved';
    return 'rejected';
  }, [order]);

  const title =
    viewState === 'pending'
      ? '等待管理员审批'
      : viewState === 'approved'
        ? '审批通过'
        : '审批被拒绝';

  const subTitle =
    viewState === 'pending'
      ? '我们已收到你的购买申请，请稍后…'
      : viewState === 'approved'
        ? '你可以前往领取/显示账号'
        : '如有疑问，请联系管理员';

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          <Typography.Title heading={5} style={{ margin: 0 }}>
            购买进度
          </Typography.Title>
          <Typography.Text type="secondary">订单号：{id || '-'}</Typography.Text>
        </Space>
      </Card>

      <Card>
        <Result
          status={viewState === 'rejected' ? 'error' : 'info'}
          title={title}
          subTitle={subTitle}
          extra={
            <Space>
              <Button onClick={fetchOrder} loading={loading}>
                手动刷新
              </Button>
              <Button
                type="secondary"
                onClick={() => setPolling((v) => !v)}
              >
                {polling ? '停止自动刷新' : '开启自动刷新'}
              </Button>
              {viewState === 'approved' ? (
                <Button
                  type="primary"
                  onClick={() => router.push(`/dashboard/purchase/redeem?id=${encodeURIComponent(id)}`)}
                >
                  领取账号
                </Button>
              ) : null}
            </Space>
          }
        />
      </Card>
    </Space>
  );
}
