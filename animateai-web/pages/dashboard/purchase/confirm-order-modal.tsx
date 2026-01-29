import React, { useMemo, useState } from 'react';
import { Modal, Space, Typography, Radio, Button } from '@arco-design/web-react';

type PlanKey = '7d' | '30d' | '90d';

type Plan = {
  key: PlanKey;
  title: string;
  subtitle: string;
  priceText?: string;
};

type Props = {
  visible: boolean;
  plans: Plan[];
  defaultPlanKey?: PlanKey;
  onCancel: () => void;
  onConfirm: (planKey: PlanKey) => Promise<void> | void;
};

export default function ConfirmOrderModal({
  visible,
  plans,
  defaultPlanKey,
  onCancel,
  onConfirm,
}: Props) {
  const initial = useMemo<PlanKey>(() => {
    return defaultPlanKey || plans[0]?.key || '30d';
  }, [defaultPlanKey, plans]);

  const [planKey, setPlanKey] = useState<PlanKey>(initial);
  const [submitting, setSubmitting] = useState(false);

  const plan = useMemo(() => plans.find((p) => p.key === planKey), [plans, planKey]);

  const handleOk = async () => {
    try {
      setSubmitting(true);
      await onConfirm(planKey);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Modal
      title="确认购买"
      visible={visible}
      onCancel={onCancel}
      footer={
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button type="primary" loading={submitting} onClick={handleOk}>
            确认下单
          </Button>
        </Space>
      }
    >
      <Space direction="vertical" style={{ width: '100%' }} size={12}>
        <Typography.Text type="secondary">
          这是一个演示流程：提交订单后进入“等待审批”页面，管理员审批通过后可“显示账号”。
        </Typography.Text>

        <Radio.Group value={planKey} onChange={(v) => setPlanKey(v)} direction="vertical">
          {plans.map((p) => (
            <Radio value={p.key} key={p.key}>
              <Space direction="vertical" size={0}>
                <Typography.Text>{p.title}</Typography.Text>
                <Typography.Text type="secondary">{p.subtitle}{p.priceText ? ` · ${p.priceText}` : ''}</Typography.Text>
              </Space>
            </Radio>
          ))}
        </Radio.Group>

        {plan ? (
          <Typography.Text>
            你将购买：{plan.title}
          </Typography.Text>
        ) : null}
      </Space>
    </Modal>
  );
}
