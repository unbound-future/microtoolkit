import type { NextApiRequest, NextApiResponse } from 'next';
import bcrypt from 'bcryptjs';
import { prisma } from '@/utils/prisma';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    res.setHeader('Allow', ['POST']);
    return res.status(405).json({ status: 'error', msg: `Method ${req.method} Not Allowed` });
  }

  try {
    const { userName, password, nickname, email, phone } = req.body || {};

    if (!userName) return res.status(400).json({ status: 'error', msg: '用户名不能为空' });
    if (!password) return res.status(400).json({ status: 'error', msg: '密码不能为空' });

    const existed = await prisma.user.findUnique({ where: { userName: String(userName) } });
    if (existed) {
      return res.status(409).json({ status: 'error', msg: '用户名已存在' });
    }

    const passwordHash = await bcrypt.hash(String(password), 10);

    await prisma.user.create({
      data: {
        userName: String(userName),
        passwordHash,
        role: 'user',
        nickname: nickname ? String(nickname) : undefined,
        email: email ? String(email) : undefined,
        phone: phone ? String(phone) : undefined,
      },
    });

    return res.status(200).json({ status: 'ok' });
  } catch (e) {
    // avoid leaking internal errors
    return res.status(500).json({ status: 'error', msg: '注册失败' });
  }
}
