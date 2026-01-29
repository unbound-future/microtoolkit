import type { NextApiRequest, NextApiResponse } from 'next';
import bcrypt from 'bcryptjs';
import { prisma } from '@/utils/prisma';

function isAdminRole(role: unknown) {
  return role === 'admin';
}

async function getAuthedUser(req: NextApiRequest) {
  const auth = req.headers.authorization;
  if (!auth) return null;
  const [type, token] = auth.split(' ');
  if (type !== 'Basic' || !token) return null;
  try {
    const decoded = Buffer.from(token, 'base64').toString('utf8');
    const idx = decoded.indexOf(':');
    if (idx === -1) return null;
    const userName = decoded.slice(0, idx);
    const password = decoded.slice(idx + 1);
    const user = await prisma.user.findUnique({ where: { userName } });
    if (!user) return null;
    const ok = await bcrypt.compare(password, user.passwordHash);
    if (!ok) return null;
    return user;
  } catch {
    return null;
  }
}

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  const authed = await getAuthedUser(req);
  if (!authed || !isAdminRole(authed.role)) {
    return res.status(403).json({ msg: 'Forbidden' });
  }

  if (req.method === 'GET') {
    const { q } = req.query;
    const where =
      typeof q === 'string' && q
        ? {
            OR: [
              { userName: { contains: q } },
              { nickname: { contains: q } },
              { email: { contains: q } },
              { phone: { contains: q } },
            ],
          }
        : undefined;

    const users = await prisma.user.findMany({
      where,
      orderBy: { createdAt: 'desc' },
      select: {
        id: true,
        userName: true,
        role: true,
        nickname: true,
        email: true,
        phone: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    return res.status(200).json(users);
  }

  if (req.method === 'POST') {
    const { userName, password, role, nickname, email, phone } = req.body || {};
    if (!userName) return res.status(400).json({ msg: 'userName is required' });
    if (!password) return res.status(400).json({ msg: 'password is required' });

    const existed = await prisma.user.findUnique({ where: { userName: String(userName) } });
    if (existed) return res.status(409).json({ msg: 'userName already exists' });

    const passwordHash = await bcrypt.hash(String(password), 10);
    const created = await prisma.user.create({
      data: {
        userName: String(userName),
        passwordHash,
        role: role === 'admin' ? 'admin' : 'user',
        nickname: nickname ? String(nickname) : undefined,
        email: email ? String(email) : undefined,
        phone: phone ? String(phone) : undefined,
      },
      select: {
        id: true,
        userName: true,
        role: true,
        nickname: true,
        email: true,
        phone: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    return res.status(201).json(created);
  }

  if (req.method === 'PATCH') {
    const { id, role, nickname, email, phone, password } = req.body || {};
    if (!id) return res.status(400).json({ msg: 'id is required' });

    const data: any = {};
    if (typeof role !== 'undefined') data.role = role === 'admin' ? 'admin' : 'user';
    if (typeof nickname !== 'undefined') data.nickname = nickname ? String(nickname) : null;
    if (typeof email !== 'undefined') data.email = email ? String(email) : null;
    if (typeof phone !== 'undefined') data.phone = phone ? String(phone) : null;
    if (typeof password !== 'undefined' && String(password)) {
      data.passwordHash = await bcrypt.hash(String(password), 10);
    }

    const updated = await prisma.user.update({
      where: { id: String(id) },
      data,
      select: {
        id: true,
        userName: true,
        role: true,
        nickname: true,
        email: true,
        phone: true,
        createdAt: true,
        updatedAt: true,
      },
    });

    return res.status(200).json(updated);
  }

  if (req.method === 'DELETE') {
    const { id } = req.query;
    if (typeof id !== 'string' || !id) return res.status(400).json({ msg: 'id is required' });

    await prisma.user.delete({ where: { id } });
    return res.status(200).json({ ok: true });
  }

  res.setHeader('Allow', ['GET', 'POST', 'PATCH', 'DELETE']);
  return res.status(405).json({ msg: `Method ${req.method} Not Allowed` });
}
