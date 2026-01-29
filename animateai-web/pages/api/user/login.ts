import type { NextApiRequest, NextApiResponse } from 'next';
import bcrypt from 'bcryptjs';
import { prisma } from '@/utils/prisma';

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    res.setHeader('Allow', ['POST']);
    return res
      .status(405)
      .json({ status: 'error', msg: `Method ${req.method} Not Allowed` });
  }

  try {
    const { userName, password } = req.body || {};
    if (!userName) return res.status(400).json({ status: 'error', msg: '用户名不能为空' });
    if (!password) return res.status(400).json({ status: 'error', msg: '密码不能为空' });

    const user = await prisma.user.findUnique({ where: { userName: String(userName) } });
    if (!user) return res.status(401).json({ status: 'error', msg: '账号或者密码错误' });

    const ok = await bcrypt.compare(String(password), user.passwordHash);
    if (!ok) return res.status(401).json({ status: 'error', msg: '账号或者密码错误' });

    return res.status(200).json({ status: 'ok' });
  } catch (e) {
    return res.status(500).json({ status: 'error', msg: '登录失败' });
  }
}
