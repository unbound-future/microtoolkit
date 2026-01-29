import type { NextApiRequest, NextApiResponse } from 'next';
import bcrypt from 'bcryptjs';
import { prisma } from '@/utils/prisma';

export async function requireAuthedUser(req: NextApiRequest) {
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

export async function requireAdmin(req: NextApiRequest, res: NextApiResponse) {
  const user = await requireAuthedUser(req);
  if (!user) {
    res.status(401).json({ msg: 'Unauthorized' });
    return null;
  }
  if (user.role !== 'admin') {
    res.status(403).json({ msg: 'Forbidden' });
    return null;
  }
  return user;
}
