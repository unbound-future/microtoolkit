import React, { useMemo, useState } from 'react';
import { Card, Space, Typography, Button, Grid, Message, Steps, Result } from '@arco-design/web-react';
import { useRouter } from 'next/router';
import internalApi from '@/utils/internalApi';
import ConfirmOrderModal from './confirm-order-modal';

type PlanKey = '7d' | '30d' | '90d';

type OrderStatus = 'pending' | 'approved' | 'rejected';

type Order = {
  id: string;
  planKey: PlanKey;
  planName: string;
  status: OrderStatus;
  createdAt: string;
};

type Plan = { key: PlanKey; title: string; subtitle: string; priceText?: string };

const plans: Plan[] = [
  { key: '7d', title: '7天 AI-VIP', subtitle: '短期体验', priceText: '¥9.9' },
  { key: '30d', title: '30天 AI-VIP', subtitle: '月度订阅', priceText: '¥29.9' },
  { key: '90d', title: '3个月 AI-VIP', subtitle: '季度订阅', priceText: '¥79.9' },
];

export default function PurchasePage() {
  const router = useRouter();

  const [creating, setCreating] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [order, setOrder] = useState<Order | null>(null);
  const [confirmVisible, setConfirmVisible] = useState(false);
  const [selectedPlan, setSelectedPlan] = useState<PlanKey>('30d');



  const createOrder = async (planKey: PlanKey) => {
    try {
      setCreating(true);
      const plan = plans.find((p) => p.key === planKey);
      const res = await internalApi.post('/api/orders', {
        planKey,
        planName: plan?.title || planKey,
      });
      if (res?.data) {
        setOrder(res.data);
        setConfirmVisible(false);
        Message.success('下单成功，请完成模拟支付');
        router.push(`/dashboard/purchase/pay?id=${encodeURIComponent(res.data.id)}`);
      }
    } catch (e) {
      Message.error('提交失败，请稍后重试');
    } finally {
      setCreating(false);
    }
  };

  const refreshOrder = async () => {
    if (!order?.id) return;
    try {
      setRefreshing(true);
      const res = await internalApi.get(`/api/orders?id=${encodeURIComponent(order.id)}`);
      if (res?.data) {
        setOrder(res.data);
      }
    } catch (e) {
      Message.error('刷新失败');
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <Space direction="vertical" style={{ width: '100%' }} size={16}>
      <Card>
        <Space direction="vertical" style={{ width: '100%' }} size={8}>
          <Typography.Title heading={5} style={{ margin: 0 }}>
            购买 AI-VIP
          </Typography.Title>
          <Typography.Text type="secondary">
            选择套餐并提交申请，管理员审批通过后将展示账号信息（此处可用“账号已显示”提示模拟）。
          </Typography.Text>
        </Space>
      </Card>

      <Card>
        <Steps current={0}>
          <Steps.Step title="选择套餐" />
          <Steps.Step title="确认下单" />
          <Steps.Step title="等待审批" />
          <Steps.Step title="领取账号" />
        </Steps>
      </Card>

      <Grid.Row gutter={16}>
        {plans.map((p) => (
          <Grid.Col key={p.key} xs={24} sm={24} md={8} lg={8} xl={8}>
            <Card hoverable>
              <Space direction="vertical" style={{ width: '100%' }} size={12}>
                <Space direction="vertical" size={2}>
                  <Typography.Title heading={6} style={{ margin: 0 }}>
                    {p.title}
                  </Typography.Title>
                  <Typography.Text type="secondary">{p.subtitle}</Typography.Text>
                  {p.priceText ? <Typography.Text>{p.priceText}</Typography.Text> : null}
                </Space>
                <Button
                  type="primary"
                  long
                  onClick={() => {
                    setSelectedPlan(p.key);
                    setConfirmVisible(true);
                  }}
                >
                  选择并继续
                </Button>
              </Space>
            </Card>
          </Grid.Col>
        ))}
      </Grid.Row>

      <Card>
        <Result
          status="info"
          title="已下单/等待审批"
          subTitle="如果你已经下单，可以从订单号进入等待页面。"
          extra={
            <Space>
              <Button
                onClick={() => {
                  const id = order?.id;
                  if (id) router.push(`/dashboard/purchase/waiting?id=${encodeURIComponent(id)}`);
                  else Message.info('暂无订单');
                }}
              >
                查看当前订单
              </Button>
              <Button type="secondary" onClick={() => setOrder(null)}>
                清空本地订单
              </Button>
            </Space>
          }
        />
      </Card>

      <ConfirmOrderModal
        visible={confirmVisible}
        plans={plans}
        defaultPlanKey={selectedPlan}
        onCancel={() => setConfirmVisible(false)}
        onConfirm={(planKey) => createOrder(planKey)}
      />
    </Space>
  );
}
