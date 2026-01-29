import React, { useEffect, useState } from 'react';
import { Card, Space, Button, Result } from '@arco-design/web-react';
import { useRouter } from 'next/router';
import internalApi from '@/utils/internalApi';

export default function PurchasePaySuccessPage() {
  const router = useRouter();
  const id = typeof router.query.id === 'string' ? router.query.id : '';

  const [seconds, setSeconds] = useState(3);

  useEffect(() => {
    if (!id) return;
    const timer = window.setInterval(() => {
      setSeconds((s) => (s > 0 ? s - 1 : 0));
    }, 1000);
    return () => window.clearInterval(timer);
  }, [id]);

  useEffect(() => {
    if (!id) return;
    if (seconds !== 0) return;
    router.push(`/dashboard/purchase/waiting?id=${encodeURIComponent(id)}`);
  }, [seconds, id]);

  const refreshPaid = async () => {
    if (!id) return;
    await internalApi.get(`/api/orders?id=${encodeURIComponent(id)}`);
  };

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
        <Result
          status="success"
          title="支付成功（模拟）"
          subTitle={`即将跳转到等待审批（${seconds}s）`}
          extra={
            <Space>
              <Button type="primary" onClick={() => router.push(`/dashboard/purchase/waiting?id=${encodeURIComponent(id)}`)}>
                立即进入等待审批
              </Button>
              <Button onClick={refreshPaid}>刷新</Button>
            </Space>
          }
        />
      </Card>
    </Space>
  );
}
