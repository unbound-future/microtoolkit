import type { NextApiRequest, NextApiResponse } from 'next';

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

declare global {
  // eslint-disable-next-line no-var
  var __ORDERS_STORE__: { orders: Order[] } | undefined;
}

function getStore() {
  if (!global.__ORDERS_STORE__) {
    global.__ORDERS_STORE__ = { orders: [] };
  }
  return global.__ORDERS_STORE__;
}

function genId() {
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export default function handler(req: NextApiRequest, res: NextApiResponse) {
  const store = getStore();

  if (req.method === 'GET') {
    const { id } = req.query;
    if (typeof id === 'string' && id) {
      const order = store.orders.find((o) => o.id === id);
      if (!order) {
        return res.status(404).json({ message: 'Order not found' });
      }
      return res.status(200).json(order);
    }
    return res.status(200).json(store.orders);
  }

  if (req.method === 'POST') {
    const { planKey, planName } = req.body || {};
    if (!planKey) {
      return res.status(400).json({ message: 'planKey is required' });
    }

    const order: Order = {
      id: genId(),
      planKey: String(planKey),
      planName: String(planName || planKey),
      status: 'pending',
      createdAt: new Date().toISOString(),
    };

    store.orders.unshift(order);
    return res.status(201).json(order);
  }

  if (req.method === 'PATCH') {
    const { id, status, paid } = req.body || {};
    if (!id) {
      return res.status(400).json({ message: 'id is required' });
    }

    if (
      typeof status !== 'undefined' &&
      status !== 'pending' &&
      status !== 'approved' &&
      status !== 'rejected'
    ) {
      return res.status(400).json({ message: 'invalid status' });
    }

    const idx = store.orders.findIndex((o) => o.id === String(id));
    if (idx === -1) {
      return res.status(404).json({ message: 'Order not found' });
    }

    const next = { ...store.orders[idx] };
    if (typeof status !== 'undefined') {
      next.status = status;
    }
    if (typeof paid !== 'undefined') {
      next.paid = Boolean(paid);
      next.paidAt = next.paid ? new Date().toISOString() : undefined;
    }

    store.orders[idx] = next;
    return res.status(200).json(store.orders[idx]);
  }

  res.setHeader('Allow', ['GET', 'POST', 'PATCH']);
  return res.status(405).json({ message: `Method ${req.method} Not Allowed` });
}
